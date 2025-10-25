package lib

import (
	"fmt"

	commonPB "go.viam.com/api/common/v1"
)

// PoseJSON represents a pose configuration in JSON format.
// It contains position (in millimeters) and orientation information.
type PoseJSON struct {
	// millimeters from the origin
	X float64 `json:"x,omitempty"`
	// millimeters from the origin
	Y float64 `json:"y,omitempty"`
	// millimeters from the origin
	Z float64 `json:"z,omitempty"`
	// z component of a vector defining axis of rotation
	OX float64 `json:"o_x,omitempty"`
	// x component of a vector defining axis of rotation
	OY float64 `json:"o_y,omitempty"`
	// y component of a vector defining axis of rotation
	OZ float64 `json:"o_z,omitempty"`
	// degrees
	Theta float64 `json:"theta,omitempty"`
}

// ParsePose parses a pose from JSON data.
// It handles various numeric types and uses default values (0.0) for missing components.
//
// Parameters:
//   - poseData: JSON object containing pose data
//
// Returns the parsed pose or an error if parsing fails.
func ParsePose(data any) (*commonPB.Pose, error) {
	pose, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected pose object, got %T", data)
	}

	x := parseFloat(pose["x"], 0.0)
	y := parseFloat(pose["y"], 0.0)
	z := parseFloat(pose["z"], 0.0)
	oX := parseFloat(pose["o_x"], 0.0)
	oY := parseFloat(pose["o_y"], 0.0)
	oZ := parseFloat(pose["o_z"], 0.0)
	theta := parseFloat(pose["theta"], 0.0)

	return &commonPB.Pose{
		X:     x,
		Y:     y,
		Z:     z,
		OX:    oX,
		OY:    oY,
		OZ:    oZ,
		Theta: theta,
	}, nil
}

// PoseFromJSON converts a PoseJSON object to a commonPB.Pose.
// It returns a new pose with the position and orientation values from the JSON data.
//
// Parameters:
//   - data: PoseJSON object containing pose data
//
// Returns the converted pose.
func PoseFromJSON(data PoseJSON) *commonPB.Pose {
	return &commonPB.Pose{
		X:     data.X,
		Y:     data.Y,
		Z:     data.Z,
		OX:    data.OX,
		OY:    data.OY,
		OZ:    data.OZ,
		Theta: data.Theta,
	}
}

// PoseToMeters converts a pose's position from millimeters to meters.
// Only the position components (X, Y, Z) are converted; orientation values remain unchanged.
//
// Parameters:
//   - pose: Pose with position in millimeters
//
// Returns a new pose with position converted to meters.
func PoseToMeters(pose *commonPB.Pose) *commonPB.Pose {
	return &commonPB.Pose{
		X:     pose.X / 1000.0,
		Y:     pose.Y / 1000.0,
		Z:     pose.Z / 1000.0,
		OX:    pose.OX,
		OY:    pose.OY,
		OZ:    pose.OZ,
		Theta: pose.Theta,
	}
}
