package lib

import (
	"fmt"

	"go.viam.com/rdk/spatialmath"
)

type Pose struct {
	// millimeters from the origin
	X float64 `protobuf:"fixed64,1,opt,name=x,proto3" json:"x,omitempty"`
	// millimeters from the origin
	Y float64 `protobuf:"fixed64,2,opt,name=y,proto3" json:"y,omitempty"`
	// millimeters from the origin
	Z float64 `protobuf:"fixed64,3,opt,name=z,proto3" json:"z,omitempty"`
	// z component of a vector defining axis of rotation
	OX float64 `protobuf:"fixed64,4,opt,name=o_x,json=oX,proto3" json:"o_x,omitempty"`
	// x component of a vector defining axis of rotation
	OY float64 `protobuf:"fixed64,5,opt,name=o_y,json=oY,proto3" json:"o_y,omitempty"`
	// y component of a vector defining axis of rotation
	OZ float64 `protobuf:"fixed64,6,opt,name=o_z,json=oZ,proto3" json:"o_z,omitempty"`
	// degrees
	Theta float64 `protobuf:"fixed64,7,opt,name=theta,proto3" json:"theta,omitempty"`
}

func ParsePose(poseData any) (Pose, error) {
	poseMap, ok := poseData.(map[string]any)
	if !ok {
		return Pose{}, fmt.Errorf("expected pose object, got %T", poseData)
	}

	x := ParseFloat64(poseMap["x"], 0.0)
	y := ParseFloat64(poseMap["y"], 0.0)
	z := ParseFloat64(poseMap["z"], 0.0)

	oX := ParseFloat64(poseMap["o_x"], 0.0)
	oY := ParseFloat64(poseMap["o_y"], 0.0)
	oZ := ParseFloat64(poseMap["o_z"], 0.0)
	theta := ParseFloat64(poseMap["theta"], 0.0)

	return Pose{
		X: x, Y: y, Z: z,
		OX: oX, OY: oY, OZ: oZ,
		Theta: theta,
	}, nil
}

func PoseFromSpatialMath(pose spatialmath.Pose) Pose {
	return Pose{
		X:     pose.Point().X,
		Y:     pose.Point().Y,
		Z:     pose.Point().Z,
		OX:    pose.Orientation().OrientationVectorRadians().OX,
		OY:    pose.Orientation().OrientationVectorRadians().OY,
		OZ:    pose.Orientation().OrientationVectorRadians().OZ,
		Theta: pose.Orientation().OrientationVectorRadians().Theta,
	}
}
