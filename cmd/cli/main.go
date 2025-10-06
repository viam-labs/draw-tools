package main

import (
	"context"
	"drawmotionplan"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	worldstatestore "go.viam.com/rdk/services/worldstatestore"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	ctx := context.Background()
	logger := logging.NewLogger("cli")

	deps := resource.Dependencies{}
	// can load these from a remote machine if you need

	cfg := drawmotionplan.Config{}

	thing, err := drawmotionplan.NewAsArrows(ctx, deps, worldstatestore.Named("foo"), &cfg, logger)
	if err != nil {
		return err
	}
	defer thing.Close(ctx)

	return nil
}
