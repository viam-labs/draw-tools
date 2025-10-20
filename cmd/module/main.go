package main

import (
	drawarrows "github.com/viam-labs/draw-tools/drawarrows"
	cleararrowsbutton "github.com/viam-labs/draw-tools/drawarrows/clearbutton"
	drawarrowsbutton "github.com/viam-labs/draw-tools/drawarrows/drawbutton"
	"github.com/viam-labs/draw-tools/drawmesh"
	clearmeshbutton "github.com/viam-labs/draw-tools/drawmesh/clearbutton"
	drawmeshbutton "github.com/viam-labs/draw-tools/drawmesh/drawbutton"

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
		resource.APIModel{API: worldstatestore.API, Model: drawmesh.WorldState},
		resource.APIModel{API: button.API, Model: clearmeshbutton.ClearMesh},
		resource.APIModel{API: button.API, Model: drawmeshbutton.DrawMesh},
	)
}
