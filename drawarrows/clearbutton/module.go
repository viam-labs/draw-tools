package cleararrowsbutton

import (
	"context"
	"fmt"

	button "go.viam.com/rdk/components/button"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

var (
	ClearArrows = resource.NewModel("viam-viz", "draw-tools", "clear-arrows-button")
)

func init() {
	resource.RegisterComponent(button.API, ClearArrows,
		resource.Registration[button.Button, *Config]{
			Constructor: newClearArrowsButton,
		},
	)

}

type Config struct {
	ServiceName string `json:"service_name"`
}

func (config *Config) Validate(path string) ([]string, []string, error) {
	if config.ServiceName == "" {
		return nil, nil, resource.NewConfigValidationFieldRequiredError(path, "service_name")
	}

	return nil, nil, nil
}

type clearArrowsButton struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	config *Config

	cancelCtx  context.Context
	cancelFunc func()

	service worldstatestore.Service
}

func newClearArrowsButton(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
	logger logging.Logger,
) (button.Button, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewClearArrowsButton(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewClearArrowsButton(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	conf *Config,
	logger logging.Logger,
) (button.Button, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	component := &clearArrowsButton{
		name:       name,
		logger:     logger,
		config:     conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}

	service, err := worldstatestore.FromDependencies(deps, conf.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get world state store %v: %w", conf.ServiceName, err)
	}

	component.service = service
	return component, nil
}

func (s *clearArrowsButton) Name() resource.Name {
	return s.name
}

func (s *clearArrowsButton) Push(ctx context.Context, extra map[string]interface{}) error {
	result, err := s.service.DoCommand(ctx, map[string]interface{}{
		"clear": map[string]interface{}{},
	})
	if err != nil {
		return err
	}

	if result["success"] != true {
		return fmt.Errorf("Failed to clear arrows: %s", result["error"])
	}

	return nil
}

func (s *clearArrowsButton) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Not implemented, use service DoCommand instead")
}

func (s *clearArrowsButton) Close(context.Context) error {
	s.cancelFunc()
	return nil
}
