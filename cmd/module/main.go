package main

import (
	"drawmotionplan"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(resource.APIModel{ worldstatestore.API, drawmotionplan.AsArrows})
}
