package drawarrows

import (
	"context"
	"drawtools/lib"
	"fmt"
	"sync"

	"github.com/google/uuid"
	commonPB "go.viam.com/api/common/v1"
	v1 "go.viam.com/api/service/worldstatestore/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	WorldState = resource.NewModel("viam-viz", "draw-tools", "draw-arrows-world-state")
)

func init() {
	resource.RegisterService(worldstatestore.API, WorldState,
		resource.Registration[worldstatestore.Service, *Config]{
			Constructor: newWorldStateService,
		},
	)
}

type Config struct {
	Arrows []lib.Arrow `json:"arrows"`
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

	if conf.Arrows != nil {
		for _, toDraw := range conf.Arrows {
			service.draw(ctx, []lib.Arrow{toDraw})
		}
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

func (service *worldStateService) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	if drawData, ok := cmd["draw"]; ok {
		arrows, err := lib.ParseArrows(drawData)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, err
		}

		count, err := service.draw(ctx, arrows)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, err
		}

		return map[string]any{
			"success":      true,
			"arrows_added": count,
		}, nil
	}

	if _, ok := cmd["clear"]; ok {
		count, err := service.clear(ctx)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, err
		}

		return map[string]any{
			"success":        true,
			"arrows_removed": count,
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

func (service *worldStateService) draw(ctx context.Context, arrows []lib.Arrow) (int, error) {
	transforms, err := lib.DrawArrows(arrows)
	if err != nil {
		service.logger.Errorw("Failed to create arrow transform", "error", err.Error())
		return 0, err
	}

	service.transformsMutex.Lock()
	defer service.transformsMutex.Unlock()

	for i := range transforms {
		transform := &transforms[i]
		id, err := uuid.FromBytes(transform.Uuid)
		if err != nil {
			service.logger.Errorw("Failed to parse UUID", "error", err.Error())
			return 0, err
		}

		service.transforms[id.String()] = transform
		service.emitChange(worldstatestore.TransformChange{
			ChangeType: v1.TransformChangeType_TRANSFORM_CHANGE_TYPE_ADDED,
			Transform:  transform,
		})
	}

	return len(transforms), nil
}

func (service *worldStateService) clear(ctx context.Context) (int, error) {
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

func (service *worldStateService) updateArrowColor(id string, newColor lib.Color) {
	transform, exists := service.transforms[id]
	if !exists {
		return
	}

	metadata, err := structpb.NewStruct(map[string]any{
		"shape": "arrow",
		"color": map[string]any{
			"r": int(newColor.R),
			"g": int(newColor.G),
			"b": int(newColor.B),
		},
	})

	if err != nil {
		service.logger.Errorw("Failed to update arrow color", "uuid", fmt.Sprintf("%x", id), "error", err.Error())
		return
	}

	transform.Metadata = metadata
	partialTransform := &commonPB.Transform{
		Uuid:     transform.Uuid,
		Metadata: metadata,
	}

	change := worldstatestore.TransformChange{
		ChangeType:    v1.TransformChangeType_TRANSFORM_CHANGE_TYPE_UPDATED,
		Transform:     partialTransform,
		UpdatedFields: []string{"metadata"},
	}

	service.emitChange(change)
}
