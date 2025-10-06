package main

import (
	"drawmotionplan/as_arrows"
	// Future models will be imported here:
	// "drawmotionplan/as_points"
	// "drawmotionplan/as_paths"

	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

func main() {
	// Register all models in the module
	// ModularMain can take multiple APIModel arguments for multiple models
	module.ModularMain(
		resource.APIModel{API: worldstatestore.API, Model: as_arrows.AsArrows},
		// Future models will be added here:
		// resource.APIModel{API: worldstatestore.API, Model: as_points.AsPoints},
		// resource.APIModel{API: worldstatestore.API, Model: as_paths.AsPaths},
	)
}
