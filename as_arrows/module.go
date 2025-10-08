package as_arrows

import (
	"context"
	"drawmotionplan/lib"
	"fmt"
	"sync"

	"github.com/google/uuid"
	commonPB "go.viam.com/api/common/v1"
	v1 "go.viam.com/api/service/worldstatestore/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
	"go.viam.com/rdk/spatialmath"
)

var (
	AsArrows = resource.NewModel("viam-viz", "draw-motion-plan", "as-arrows")

	MinUpdateRateHz = 16.6667
)

func init() {
	resource.RegisterService(worldstatestore.API, AsArrows,
		resource.Registration[worldstatestore.Service, *Config]{
			Constructor: newDrawMotionPlanAsArrows,
		},
	)
}

type Config struct {
}

type arrow struct {
	Pose        spatialmath.Pose `json:"pose"`
	Color       lib.Color        `json:"color,omitempty"`        // optional, defaults to { R: 255, G: 255, B: 0 }
	ParentFrame string           `json:"parent_frame,omitempty"` // optional, defaults to "world"
}

// Validate ensures all parts of the config are valid and important fields exist.
func (cfg *Config) Validate(path string) ([]string, []string, error) {

	return []string{}, nil, nil
}

type drawMotionPlanAsArrows struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	config *Config

	cancelCtx  context.Context
	cancelFunc func()

	transforms      map[string]*storedTransform
	transformsMutex sync.RWMutex

	changeStream chan worldstatestore.TransformChange

	workers sync.WaitGroup
}

type storedTransform struct {
	UUID      uuid.UUID
	Transform *commonPB.Transform
}

func newDrawMotionPlanAsArrows(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
	logger logging.Logger,
) (worldstatestore.Service, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewAsArrows(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewAsArrows(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	conf *Config,
	logger logging.Logger,
) (worldstatestore.Service, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	service := &drawMotionPlanAsArrows{
		name:         name,
		logger:       logger,
		config:       conf,
		cancelCtx:    cancelCtx,
		cancelFunc:   cancelFunc,
		transforms:   make(map[string]*storedTransform),
		changeStream: make(chan worldstatestore.TransformChange, 100),
	}

	return service, nil
}

func (service *drawMotionPlanAsArrows) Name() resource.Name {
	return service.name
}

func (service *drawMotionPlanAsArrows) ListUUIDs(ctx context.Context, extra map[string]any) ([][]byte, error) {
	service.transformsMutex.RLock()
	defer service.transformsMutex.RUnlock()

	uuids := make([][]byte, 0, len(service.transforms))
	for _, transform := range service.transforms {
		uuids = append(uuids, transform.UUID[:])
	}

	return uuids, nil
}

func (service *drawMotionPlanAsArrows) GetTransform(ctx context.Context, uuid []byte, extra map[string]any) (*commonPB.Transform, error) {
	service.transformsMutex.RLock()
	defer service.transformsMutex.RUnlock()

	uuidString := string(uuid)
	transform, ok := service.transforms[uuidString]
	if !ok {
		return nil, fmt.Errorf("transform not found for UUID: %x", uuid)
	}

	return transform.Transform, nil
}

func (service *drawMotionPlanAsArrows) StreamTransformChanges(ctx context.Context, extra map[string]any) (*worldstatestore.TransformChangeStream, error) {
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

func (service *drawMotionPlanAsArrows) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	if drawData, ok := cmd["draw"]; ok {
		arrows, err := service.parseArrows(drawData)
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

	return nil, fmt.Errorf("unknown command")
}

func (service *drawMotionPlanAsArrows) Close(context.Context) error {
	service.cancelFunc()
	service.workers.Wait()
	close(service.changeStream)
	return nil
}

func (service *drawMotionPlanAsArrows) emitChange(change worldstatestore.TransformChange) {
	select {
	case service.changeStream <- change:
		// Successfully sent
	default:
		service.logger.Warnw("change stream buffer full, dropping change")
	}
}

func (service *drawMotionPlanAsArrows) draw(ctx context.Context, arrows []arrow) (int, error) {
	service.logger.Infow("Creating arrows from motion plan", "arrow_count", len(arrows))

	transforms, err := service.drawArrows(arrows)
	if err != nil {
		service.logger.Errorw("Failed to create arrow transform", "error", err)
		return 0, err
	}

	service.transformsMutex.Lock()
	defer service.transformsMutex.Unlock()

	for i := range transforms {
		transform := &transforms[i]
		uuidString := string(transform.Uuid)
		service.transforms[uuidString] = &storedTransform{
			UUID:      uuid.UUID(transform.Uuid),
			Transform: transform,
		}

		service.emitChange(worldstatestore.TransformChange{
			ChangeType: v1.TransformChangeType_TRANSFORM_CHANGE_TYPE_ADDED,
			Transform:  transform,
		})
	}

	service.logger.Infow("Successfully created arrows", "arrow_count", len(transforms))
	return len(transforms), nil
}

func (service *drawMotionPlanAsArrows) clear(ctx context.Context) (int, error) {
	service.transformsMutex.Lock()
	defer service.transformsMutex.Unlock()

	count := len(service.transforms)
	for uuid := range service.transforms {
		service.emitChange(worldstatestore.TransformChange{
			ChangeType: v1.TransformChangeType_TRANSFORM_CHANGE_TYPE_REMOVED,
			Transform: &commonPB.Transform{
				Uuid: []byte(uuid),
			},
		})
	}

	service.transforms = make(map[string]*storedTransform)
	return count, nil
}
