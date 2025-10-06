package as_arrows

import (
	"context"
	"fmt"
	"sync"
	"time"

	commonPB "go.viam.com/api/common/v1"
	v1 "go.viam.com/api/service/worldstatestore/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/motion"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
	"go.viam.com/rdk/spatialmath"
)

var (
	AsArrows = resource.NewModel("viam-viz", "draw-motion-plan", "as-arrows")
)

func init() {
	resource.RegisterService(worldstatestore.API, AsArrows,
		resource.Registration[worldstatestore.Service, *Config]{
			Constructor: newDrawMotionPlanAsArrows,
		},
	)
}

type Config struct {
	MotionService string   `json:"motion_service"`           // Required: name of motion service to use
	UpdateRateHz  *float64 `json:"update_rate_hz,omitempty"` // Optional: rate to fetch motion plan (Hz)
}

// Validate ensures all parts of the config are valid and important fields exist.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.MotionService == "" {
		return nil, nil, fmt.Errorf("motion_service is required")
	}

	if cfg.UpdateRateHz != nil && *cfg.UpdateRateHz <= 0 {
		return nil, nil, fmt.Errorf("update_rate_hz must be greater than 0")
	}

	return []string{cfg.MotionService}, nil, nil
}

type drawMotionPlanAsArrows struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	motionService motion.Service

	transforms   map[string]*storedTransform
	transformsMu sync.RWMutex

	changeStreams   []*transformChangeStream
	changeStreamsMu sync.Mutex

	workers sync.WaitGroup
}

type storedTransform struct {
	UUID      []byte
	Transform *commonPB.Transform
}

type transformChangeStream struct {
	changes chan worldstatestore.TransformChange
	ctx     context.Context
	cancel  func()
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

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	s := &drawMotionPlanAsArrows{
		name:          name,
		logger:        logger,
		cfg:           conf,
		cancelCtx:     cancelCtx,
		cancelFunc:    cancelFunc,
		motionService: motionService,
		transforms:    make(map[string]*storedTransform),
		changeStreams: make([]*transformChangeStream, 0),
	}

	if conf.UpdateRateHz != nil {
		s.startBackgroundWorker()
	}

	return s, nil
}

func (s *drawMotionPlanAsArrows) Name() resource.Name {
	return s.name
}

func (s *drawMotionPlanAsArrows) ListUUIDs(ctx context.Context, extra map[string]interface{}) ([][]byte, error) {
	s.transformsMu.RLock()
	defer s.transformsMu.RUnlock()

	uuids := make([][]byte, 0, len(s.transforms))
	for _, t := range s.transforms {
		uuids = append(uuids, t.UUID)
	}

	return uuids, nil
}

func (s *drawMotionPlanAsArrows) GetTransform(ctx context.Context, uuid []byte, extra map[string]interface{}) (*commonPB.Transform, error) {
	s.transformsMu.RLock()
	defer s.transformsMu.RUnlock()

	uuidStr := string(uuid)
	t, ok := s.transforms[uuidStr]
	if !ok {
		return nil, fmt.Errorf("transform not found for UUID: %x", uuid)
	}

	return t.Transform, nil
}

func (s *drawMotionPlanAsArrows) StreamTransformChanges(ctx context.Context, extra map[string]interface{}) (*worldstatestore.TransformChangeStream, error) {
	s.changeStreamsMu.Lock()
	defer s.changeStreamsMu.Unlock()

	streamCtx, cancel := context.WithCancel(ctx)
	stream := &transformChangeStream{
		changes: make(chan worldstatestore.TransformChange, 100),
		ctx:     streamCtx,
		cancel:  cancel,
	}

	s.changeStreams = append(s.changeStreams, stream)
	changeStream := worldstatestore.NewTransformChangeStreamFromChannel(streamCtx, stream.changes)

	go func() {
		<-streamCtx.Done()
		s.removeChangeStream(stream)
	}()

	return changeStream, nil
}

func (s *drawMotionPlanAsArrows) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	if addCmd, ok := cmd["draw_motion_plan"]; ok {
		var componentName string
		var executionID string

		if params, ok := addCmd.(map[string]interface{}); ok {
			if cn, ok := params["component_name"].(string); ok {
				componentName = cn
			}
			if eid, ok := params["execution_id"].(string); ok {
				executionID = eid
			}
		}

		count, err := s.drawMotionPlan(ctx, componentName, executionID)
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
		count, err := s.clearAllArrows(ctx)
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

func (s *drawMotionPlanAsArrows) Close(context.Context) error {
	s.cancelFunc()
	s.workers.Wait()

	s.changeStreamsMu.Lock()
	for _, stream := range s.changeStreams {
		stream.cancel()
		close(stream.changes)
	}
	s.changeStreams = nil
	s.changeStreamsMu.Unlock()

	s.transformsMu.Lock()
	s.transforms = nil
	s.transformsMu.Unlock()

	return nil
}

func (s *drawMotionPlanAsArrows) startBackgroundWorker() {
	s.workers.Add(1)
	go func() {
		defer s.workers.Done()

		interval := time.Duration(float64(time.Second) / *s.cfg.UpdateRateHz)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		s.logger.Infow("background worker started", "rate_hz", *s.cfg.UpdateRateHz, "interval", interval)

		for {
			select {
			case <-ticker.C:
				if err := s.updateMotionPlanArrows(s.cancelCtx); err != nil {
					s.logger.Errorw("failed to update motion plan", "error", err)
				}
			case <-s.cancelCtx.Done():
				s.logger.Info("background worker stopped")
				return
			}
		}
	}()
}

func (s *drawMotionPlanAsArrows) updateMotionPlanArrows(ctx context.Context) error {

	// get the most recent motion plan
	// get the frame poses from the motion plan
	// create an arrow for each pose
	// add the arrow to the world state
	// emit a change stream event for each added/updated/removed arrow

	return nil
}

func (s *drawMotionPlanAsArrows) removeChangeStream(stream *transformChangeStream) {
	s.changeStreamsMu.Lock()
	defer s.changeStreamsMu.Unlock()

	for i, st := range s.changeStreams {
		if st == stream {
			s.changeStreams = append(s.changeStreams[:i], s.changeStreams[i+1:]...)
			break
		}
	}
}

func (s *drawMotionPlanAsArrows) emitChange(change worldstatestore.TransformChange) {
	s.changeStreamsMu.Lock()
	defer s.changeStreamsMu.Unlock()

	for _, stream := range s.changeStreams {
		select {
		case stream.changes <- change:
			// Successfully sent
		case <-stream.ctx.Done():
			// Stream is closed, skip
		default:
			// Buffer full, skip (non-blocking)
			s.logger.Warnw("change stream buffer full, dropping change")
		}
	}
}

func createArrowFromPose(pose spatialmath.Pose) *commonPB.Geometry {
	// create an arrow transform from the pose
	// should use the same arrow drawing as motion-tools

	return nil
}

func (s *drawMotionPlanAsArrows) drawMotionPlan(ctx context.Context, componentName, executionID string) (int, error) {
	s.logger.Infow("adding motion plan arrows", "component_name", componentName, "execution_id", executionID)

	count := 0
	// get motion plan by execution ID, or most recent plan if no execution ID is provided
	// get the frame poses from the motion plan
	// create an arrow for each pose
	// add the arrow to the world state
	// emit a change stream event for each arrow
	// return the number of arrows added

	return count, nil
}

func (s *drawMotionPlanAsArrows) clearAllArrows(ctx context.Context) (int, error) {
	s.transformsMu.Lock()
	defer s.transformsMu.Unlock()

	count := len(s.transforms)
	for uuid := range s.transforms {
		s.emitChange(worldstatestore.TransformChange{
			ChangeType: v1.TransformChangeType_TRANSFORM_CHANGE_TYPE_REMOVED,
			Transform: &commonPB.Transform{
				Uuid: []byte(uuid),
			},
		})
	}

	s.transforms = make(map[string]*storedTransform)
	s.logger.Infow("cleared arrows", "count", count)
	return count, nil
}

func generateUUID(componentName, executionID string, index int) []byte {
	data := fmt.Sprintf("%s-%s-%d-%d", componentName, executionID, index, time.Now().UnixNano())
	hash := fmt.Sprintf("%x", data)
	return []byte(hash)
}
