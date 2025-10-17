package main

import (
	drawarrows "github.com/viam-labs/draw-tools/drawarrows"
	cleararrowsbutton "github.com/viam-labs/draw-tools/drawarrows/clearbutton"
	drawarrowsbutton "github.com/viam-labs/draw-tools/drawarrows/drawbutton"

	"go.viam.com/rdk/components/button"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

func main() {
	// Register all models in the module
	module.ModularMain(
		resource.APIModel{API: worldstatestore.API, Model: drawarrows.WorldState},
		resource.APIModel{API: button.API, Model: cleararrowsbutton.ClearArrows},
		resource.APIModel{API: button.API, Model: drawarrowsbutton.DrawArrows},
	)
}
