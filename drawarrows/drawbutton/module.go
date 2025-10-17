package drawarrowsbutton

import (
	"context"
	"errors"
	"fmt"

	"github.com/viam-labs/draw-tools/lib"

	button "go.viam.com/rdk/components/button"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

var (
	DrawArrows = resource.NewModel("viam-viz", "draw-tools", "draw-arrows-button")
)

func init() {
	resource.RegisterComponent(button.API, DrawArrows,
		resource.Registration[button.Button, *Config]{
			Constructor: newDrawArrowsButton,
		},
	)
}

type Config struct {
	ServiceName string      `json:"service_name"`
	Arrows      []lib.Arrow `json:"arrows"`
}

func (config *Config) Validate(path string) ([]string, []string, error) {
	if config.ServiceName == "" {
		return nil, nil, resource.NewConfigValidationFieldRequiredError(path, "service_name")
	}

	if config.Arrows == nil {
		return nil, nil, resource.NewConfigValidationFieldRequiredError(path, "arrows")
	}

	if len(config.Arrows) == 0 {
		return nil, nil, resource.NewConfigValidationError(path, errors.New("arrows must be a non-empty array"))
	}

	return nil, nil, nil
}

type drawArrowsButton struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	config *Config

	cancelCtx  context.Context
	cancelFunc func()

	service worldstatestore.Service
}

func newDrawArrowsButton(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
	logger logging.Logger,
) (button.Button, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewDrawArrowsButton(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewDrawArrowsButton(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	conf *Config,
	logger logging.Logger,
) (button.Button, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	serviceName := worldstatestore.Named(conf.ServiceName)
	service, err := worldstatestore.FromDependencies(deps, serviceName.Name)
	if err != nil {
		return nil, err
	}

	component := &drawArrowsButton{
		name:       name,
		logger:     logger,
		config:     conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		service:    service,
	}

	return component, nil
}

func (s *drawArrowsButton) Name() resource.Name {
	return s.name
}

func (s *drawArrowsButton) Push(ctx context.Context, extra map[string]interface{}) error {
	result, err := s.service.DoCommand(ctx, map[string]interface{}{
		"draw": s.config.Arrows,
	})
	if err != nil {
		return err
	}

	if result["success"] != true {
		return fmt.Errorf("Failed to draw arrows: %s", result["error"])
	}

	return nil
}

func (s *drawArrowsButton) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Not implemented, use service DoCommand instead")
}

func (s *drawArrowsButton) Close(context.Context) error {
	s.cancelFunc()
	return nil
}
