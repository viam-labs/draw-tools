package as_arrows

import (
	"context"
	"drawmotionplan/colors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	commonPB "go.viam.com/api/common/v1"
	v1 "go.viam.com/api/service/worldstatestore/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/motion"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
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
	MotionService string        `json:"motion_service"`           // Required: name of motion service to use
	UpdateRateHz  *float64      `json:"update_rate_hz,omitempty"` // Optional: rate to fetch motion plan (Hz)
	Color         *colors.Color `json:"color,omitempty"`          // Optional: color of the arrows (default is black)
	ParentFrame   string        `json:"parent_frame,omitempty"`   // Optional: parent frame of the arrows (default is world)
}

// Validate ensures all parts of the config are valid and important fields exist.
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.MotionService == "" {
		return nil, nil, fmt.Errorf("motion_service is required")
	}

	if cfg.UpdateRateHz != nil && *cfg.UpdateRateHz <= MinUpdateRateHz {
		return nil, nil, fmt.Errorf("update_rate_hz must be greater than or equal to %f", MinUpdateRateHz)
	}

	return []string{cfg.MotionService}, nil, nil
}

type drawMotionPlanAsArrows struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	config *Config

	cancelCtx  context.Context
	cancelFunc func()

	motionService motion.Service
	color         colors.Color
	parentFrame   string

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
	motionService, err := motion.FromDependencies(deps, conf.MotionService)
	if err != nil {
		return nil, fmt.Errorf("failed to get motion service %q: %w", conf.MotionService, err)
	}

	color := conf.Color
	if color == nil {
		color = &colors.Color{
			R: 0,
			G: 0,
			B: 0,
		}
	}

	parentFrame := conf.ParentFrame
	if parentFrame == "" {
		parentFrame = "world"
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	service := &drawMotionPlanAsArrows{
		name:          name,
		logger:        logger,
		config:        conf,
		cancelCtx:     cancelCtx,
		cancelFunc:    cancelFunc,
		motionService: motionService,
		color:         *color,
		parentFrame:   parentFrame,
		transforms:    make(map[string]*storedTransform),
		changeStream:  make(chan worldstatestore.TransformChange, 100), // Simple buffered channel
	}

	if conf.UpdateRateHz != nil {
		service.startBackgroundWorker()
	}

	return service, nil
}

func (service *drawMotionPlanAsArrows) Name() resource.Name {
	return service.name
}

func (service *drawMotionPlanAsArrows) ListUUIDs(ctx context.Context, extra map[string]interface{}) ([][]byte, error) {
	service.transformsMutex.RLock()
	defer service.transformsMutex.RUnlock()

	uuids := make([][]byte, 0, len(service.transforms))
	for _, transform := range service.transforms {
		uuids = append(uuids, transform.UUID[:])
	}

	return uuids, nil
}

func (service *drawMotionPlanAsArrows) GetTransform(ctx context.Context, uuid []byte, extra map[string]interface{}) (*commonPB.Transform, error) {
	service.transformsMutex.RLock()
	defer service.transformsMutex.RUnlock()

	uuidString := string(uuid)
	transform, ok := service.transforms[uuidString]
	if !ok {
		return nil, fmt.Errorf("transform not found for UUID: %x", uuid)
	}

	return transform.Transform, nil
}

func (service *drawMotionPlanAsArrows) StreamTransformChanges(ctx context.Context, extra map[string]interface{}) (*worldstatestore.TransformChangeStream, error) {
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

func (service *drawMotionPlanAsArrows) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	if drawCommand, ok := cmd["draw_motion_plan"]; ok {
		var componentName string
		var executionID string

		if params, ok := drawCommand.(map[string]interface{}); ok {
			if cn, ok := params["component_name"].(string); ok {
				componentName = cn
			}
			if eid, ok := params["execution_id"].(string); ok {
				executionID = eid
			}
		}

		count, err := service.drawMotionPlan(ctx, componentName, executionID)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, err
		}

		return map[string]interface{}{
			"success":      true,
			"arrows_added": count,
		}, nil
	}

	if _, ok := cmd["clear_arrows"]; ok {
		count, err := service.clearAllArrows(ctx)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, err
		}

		return map[string]interface{}{
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

func (service *drawMotionPlanAsArrows) startBackgroundWorker() {
	service.workers.Add(1)
	go func() {
		defer service.workers.Done()

		interval := time.Duration(float64(time.Second) / *service.config.UpdateRateHz)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := service.updateMotionPlanArrows(service.cancelCtx); err != nil {
					service.logger.Errorw("failed to update motion plan", "error", err)
				}
			case <-service.cancelCtx.Done():
				return
			}
		}
	}()
}

func (service *drawMotionPlanAsArrows) updateMotionPlanArrows(ctx context.Context) error {

	// get the most recent motion plan
	// get the frame poses from the motion plan
	// create an arrow for each pose
	// add the arrow to the world state
	// emit a change stream event for each added/updated/removed arrow

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

func (service *drawMotionPlanAsArrows) drawMotionPlan(ctx context.Context, componentName, executionID string) (int, error) {
	count := 0

	// get motion plan by execution ID, or most recent plan if no execution ID is provided
	// get the frame poses from the motion plan
	// create an arrow for each pose using drawArrowsFromPoses
	// add the arrow to the world state
	// emit a change stream event for each arrow
	// return the number of arrows added

	return count, nil
}

func (service *drawMotionPlanAsArrows) clearAllArrows(ctx context.Context) (int, error) {
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

func generateUUID() []byte {
	id := uuid.New()
	return id[:]
}
