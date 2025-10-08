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

	x := ParseFloat64(poseMap["x"], 0.0)
	y := ParseFloat64(poseMap["y"], 0.0)
	z := ParseFloat64(poseMap["z"], 0.0)

	oX := ParseFloat64(poseMap["o_x"], 0.0)
	oY := ParseFloat64(poseMap["o_y"], 0.0)
	oZ := ParseFloat64(poseMap["o_z"], 0.0)
	theta := ParseFloat64(poseMap["theta"], 0.0)

	point := r3.Vector{X: x, Y: y, Z: z}
	orientation := &spatialmath.OrientationVectorDegrees{
		OX: oX, OY: oY, OZ: oZ, Theta: theta,
	}

	return spatialmath.NewPose(point, orientation), nil
}
