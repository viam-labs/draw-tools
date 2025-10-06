package drawmotionplan

import (
	"context"
	"errors"
	"fmt"

	commonPB "go.viam.com/api/common/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

var (
	AsArrows         = resource.NewModel("viam-viz", "draw-motion-plan", "as-arrows")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterService(worldstatestore.API, AsArrows,
		resource.Registration[worldstatestore.Service, *Config]{
			Constructor: newDrawMotionPlanAsArrows,
		},
	)
}

type Config struct {
	/*
		Put config attributes here. There should be public/exported fields
		with a `json` parameter at the end of each attribute.

		Example config struct:
			type Config struct {
				Pin   string `json:"pin"`
				Board string `json:"board"`
				MinDeg *float64 `json:"min_angle_deg,omitempty"`
			}

		If your model does not need a config, replace *Config in the init
		function with resource.NoNativeConfig
	*/
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit required (first return) and optional (second return) dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	// Add config validation code here
	return nil, nil, nil
}

type drawMotionPlanAsArrows struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()
}

func newDrawMotionPlanAsArrows(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (worldstatestore.Service, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewAsArrows(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewAsArrows(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (worldstatestore.Service, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &drawMotionPlanAsArrows{
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *drawMotionPlanAsArrows) Name() resource.Name {
	return s.name
}

func (s *drawMotionPlanAsArrows) ListUUIDs(ctx context.Context, extra map[string]interface{}) ([][]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *drawMotionPlanAsArrows) GetTransform(ctx context.Context, uuid []byte, extra map[string]interface{}) (*commonPB.Transform, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *drawMotionPlanAsArrows) StreamTransformChanges(ctx context.Context, extra map[string]interface{}) (*worldstatestore.TransformChangeStream, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *drawMotionPlanAsArrows) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *drawMotionPlanAsArrows) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
