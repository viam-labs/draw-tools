package drawmeshbutton

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
	DrawMesh = resource.NewModel("viam-viz", "draw-tools", "draw-mesh-button")
)

func init() {
	resource.RegisterComponent(button.API, DrawMesh,
		resource.Registration[button.Button, *Config]{
			Constructor: newDrawMeshButton,
		},
	)
}

type Config struct {
	ServiceName string    `json:"service_name"`
	ModelPath   string    `json:"model_path"`
	Color       lib.Color `json:"color"`
}

func (config *Config) Validate(path string) ([]string, []string, error) {
	if config.ServiceName == "" {
		return nil, nil, resource.NewConfigValidationFieldRequiredError(path, "service_name")
	}

	if config.ModelPath == "" {
		return nil, nil, errors.New("model_path is required")
	}

	return nil, nil, nil
}

type drawMeshButton struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	config *Config

	cancelCtx  context.Context
	cancelFunc func()

	service worldstatestore.Service
}

func newDrawMeshButton(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
	logger logging.Logger,
) (button.Button, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewDrawMeshButton(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewDrawMeshButton(
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

	component := &drawMeshButton{
		name:       name,
		logger:     logger,
		config:     conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		service:    service,
	}

	return component, nil
}

func (s *drawMeshButton) Name() resource.Name {
	return s.name
}

func (s *drawMeshButton) Push(ctx context.Context, extra map[string]interface{}) error {

	color := lib.Color{R: 0, G: 0, B: 255}
	if s.config.Color != (lib.Color{}) {
		color = s.config.Color
	}
	result, err := s.service.DoCommand(ctx, map[string]interface{}{
		"draw": map[string]interface{}{
			"model_path": s.config.ModelPath,
			"color":      color,
		},
	})
	if err != nil {
		return err
	}

	if result["success"] != true {
		return fmt.Errorf("Failed to draw mesh: %s", result["error"])
	}

	return nil
}

func (s *drawMeshButton) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Not implemented, use service DoCommand instead")
}

func (s *drawMeshButton) Close(context.Context) error {
	s.cancelFunc()
	return nil
}
