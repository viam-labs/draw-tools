package lib

import (
	"fmt"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/spatialmath"
)

func ParsePose(poseData any) (spatialmath.Pose, error) {
	poseMap, ok := poseData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected pose object, got %T", poseData)
	}

	x := GetFloat64(poseMap["x"], 0.0)
	y := GetFloat64(poseMap["y"], 0.0)
	z := GetFloat64(poseMap["z"], 0.0)

	oX := GetFloat64(poseMap["o_x"], 0.0)
	oY := GetFloat64(poseMap["o_y"], 0.0)
	oZ := GetFloat64(poseMap["o_z"], 0.0)
	theta := GetFloat64(poseMap["theta"], 0.0)

	point := r3.Vector{X: x, Y: y, Z: z}
	orientation := &spatialmath.OrientationVectorDegrees{
		OX: oX, OY: oY, OZ: oZ, Theta: theta,
	}

	return spatialmath.NewPose(point, orientation), nil
}
