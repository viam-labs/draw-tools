package drawmesh

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/viam-labs/draw-tools/lib"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/google/uuid"
	commonPB "go.viam.com/api/common/v1"
	v1 "go.viam.com/api/service/worldstatestore/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
	"go.viam.com/rdk/spatialmath"
)

var (
	WorldState = resource.NewModel("viam-viz", "draw-tools", "draw-mesh-world-state")
)

func init() {
	resource.RegisterService(worldstatestore.API, WorldState,
		resource.Registration[worldstatestore.Service, *Config]{
			Constructor: newWorldStateService,
		},
	)
}

type Config struct {
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	return []string{}, nil, nil
}

type worldStateService struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	config *Config

	cancelCtx  context.Context
	cancelFunc func()

	transforms      map[string]*commonPB.Transform
	transformsMutex sync.RWMutex

	changeStream chan worldstatestore.TransformChange

	workers sync.WaitGroup
}

func newWorldStateService(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
	logger logging.Logger,
) (worldstatestore.Service, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewWorldStateService(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewWorldStateService(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	conf *Config,
	logger logging.Logger,
) (worldstatestore.Service, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	service := &worldStateService{
		name:         name,
		logger:       logger,
		config:       conf,
		cancelCtx:    cancelCtx,
		cancelFunc:   cancelFunc,
		transforms:   make(map[string]*commonPB.Transform),
		changeStream: make(chan worldstatestore.TransformChange, 100),
	}

	return service, nil
}

func (service *worldStateService) Name() resource.Name {
	return service.name
}

func (service *worldStateService) ListUUIDs(ctx context.Context, extra map[string]any) ([][]byte, error) {
	service.transformsMutex.RLock()
	defer service.transformsMutex.RUnlock()

	uuids := make([][]byte, 0, len(service.transforms))
	for _, transform := range service.transforms {
		parsedId, err := uuid.FromBytes(transform.Uuid)
		if err != nil {
			service.logger.Errorw("Failed to parse UUID", "error", err.Error())
			return nil, err
		}
		uuids = append(uuids, parsedId[:])
	}

	return uuids, nil
}

func (service *worldStateService) GetTransform(ctx context.Context, id []byte, extra map[string]any) (*commonPB.Transform, error) {
	service.transformsMutex.RLock()
	defer service.transformsMutex.RUnlock()

	uuidString, err := uuid.FromBytes(id)
	if err != nil {
		service.logger.Errorw("Failed to parse UUID", "error", err.Error())
		return nil, err
	}

	transform, ok := service.transforms[uuidString.String()]
	if !ok {
		return nil, fmt.Errorf("transform not found for UUID: %x", uuidString)
	}

	return transform, nil
}

func (service *worldStateService) StreamTransformChanges(ctx context.Context, extra map[string]any) (*worldstatestore.TransformChangeStream, error) {
	subscriberChan := make(chan worldstatestore.TransformChange, 10)
	go func() {
		defer close(subscriberChan)
		for {
			select {
			case change := <-service.changeStream:
				select {
				case subscriberChan <- change:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return worldstatestore.NewTransformChangeStreamFromChannel(ctx, subscriberChan), nil
}

func (s *worldStateService) draw(meshPath string, color lib.Color) error {
	// Read PLY from the file specified in the config (ModelPath)
	file, err := os.Open(meshPath)
	if err != nil {
		return err
	}
	defer file.Close()

	mesh, err := spatialmath.NewMeshFromPLYFile(meshPath)
	if err != nil {
		s.logger.Errorw("Error creating mesh from PLY file:", err)
		return err
	}

	s.logger.Infow("Successfully created mesh from PLY file:", meshPath)

	geometry := mesh.ToProtobuf()
	uuidBytes := lib.GenerateUUID()
	if err != nil {
		s.logger.Errorw("Failed to parse UUID", "error", err.Error())
		return err
	}

	metadata, err := structpb.NewStruct(map[string]any{
		"color": map[string]any{
			"r": int(color.R),
			"g": int(color.G),
			"b": int(color.B),
		},
	})

	transform := commonPB.Transform{
		ReferenceFrame: fmt.Sprintf("mesh-%s", uuidBytes.String()),
		PoseInObserverFrame: &commonPB.PoseInFrame{
			ReferenceFrame: "world",
			Pose: &commonPB.Pose{
				X:     0,
				Y:     0,
				Z:     0,
				OX:    0,
				OY:    0,
				OZ:    1,
				Theta: 0,
			},
		},
		Uuid:           uuidBytes.Bytes(),
		PhysicalObject: geometry,
		Metadata:       metadata,
	}

	s.transformsMutex.Lock()
	defer s.transformsMutex.Unlock()

	s.transforms[uuidBytes.String()] = &transform
	s.emitChange(worldstatestore.TransformChange{
		ChangeType: v1.TransformChangeType_TRANSFORM_CHANGE_TYPE_ADDED,
		Transform:  &transform,
	})
	s.logger.Infow("Successfully added transform to world state store:", uuidBytes.String())

	return nil
}

func (service *worldStateService) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	if drawCmd, ok := cmd["draw"]; ok {
		meshPath := drawCmd.(map[string]any)["model_path"].(string)
		colorJson := drawCmd.(map[string]any)["color"]
		color, err := lib.ParseColor(colorJson, lib.Color{R: 0, G: 0, B: 255})
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, err
		}
		err = service.draw(meshPath, color)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, err
		}

		return map[string]any{
			"success": true,
		}, nil
	}

	if _, ok := cmd["clear"]; ok {
		count, err := service.clear()
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, err
		}

		return map[string]any{
			"success":      true,
			"mesh_removed": count,
		}, nil
	}

	return nil, fmt.Errorf("Unknown command")
}

func (service *worldStateService) Close(context.Context) error {
	service.cancelFunc()
	service.workers.Wait()
	close(service.changeStream)
	return nil
}

func (service *worldStateService) emitChange(change worldstatestore.TransformChange) {
	select {
	case service.changeStream <- change:
		// Successfully sent
	default:
		service.logger.Warnw("Change stream buffer full, dropping change")
	}
}

func (service *worldStateService) clear() (int, error) {
	service.transformsMutex.Lock()
	defer service.transformsMutex.Unlock()

	count := len(service.transforms)
	for id := range service.transforms {
		parsedId, err := uuid.Parse(id)
		if err != nil {
			service.logger.Errorw("Failed to parse UUID", "error", err.Error())
			return 0, err
		}

		service.emitChange(worldstatestore.TransformChange{
			ChangeType: v1.TransformChangeType_TRANSFORM_CHANGE_TYPE_REMOVED,
			Transform: &commonPB.Transform{
				Uuid: parsedId[:],
			},
		})
	}

	service.transforms = make(map[string]*commonPB.Transform)
	return count, nil
}
