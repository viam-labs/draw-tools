package lib

import (
	"testing"

	commonPB "go.viam.com/api/common/v1"
	"go.viam.com/test"
)

var testUUID = GenerateUUID()
var testUUIDBytes = testUUID.Bytes()

func TestCreateArrow(t *testing.T) {
	tests := []struct {
		name        string
		arrowName   string
		pose        *commonPB.Pose
		uuid        []byte
		color       *Color
		parentFrame string
		expected    func(*testing.T, *Arrow, error)
	}{
		{
			name:      "valid arrow",
			arrowName: "test-arrow",
			pose: &commonPB.Pose{
				X:     100.0,
				Y:     200.0,
				Z:     300.0,
				OX:    0.0,
				OY:    0.0,
				OZ:    1.0,
				Theta: 45.0,
			},
			uuid:        testUUIDBytes,
			color:       &Color{R: 255, G: 0, B: 0},
			parentFrame: "robot",
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, arrow, test.ShouldNotBeNil)
				test.That(t, arrow.ReferenceFrame, test.ShouldEqual, "test-arrow")
				test.That(t, arrow.PoseInObserverFrame.ReferenceFrame, test.ShouldEqual, "robot")
				test.That(t, arrow.PoseInObserverFrame.Pose.X, test.ShouldEqual, 100.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.Y, test.ShouldEqual, 200.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.Z, test.ShouldEqual, 300.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.OX, test.ShouldEqual, 0.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.OY, test.ShouldEqual, 0.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.OZ, test.ShouldEqual, 1.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.Theta, test.ShouldEqual, 45.0)

				test.That(t, arrow.Uuid, test.ShouldResemble, testUUIDBytes)

				test.That(t, arrow.Metadata, test.ShouldNotBeNil)
				shape, ok := arrow.Metadata.Fields["shape"]
				test.That(t, ok, test.ShouldBeTrue)
				test.That(t, shape.GetStringValue(), test.ShouldEqual, "arrow")

				color, ok := arrow.Metadata.Fields["color"]
				test.That(t, ok, test.ShouldBeTrue)
				colorStruct := color.GetStructValue()
				test.That(t, colorStruct.Fields["r"].GetNumberValue(), test.ShouldEqual, 255)
				test.That(t, colorStruct.Fields["g"].GetNumberValue(), test.ShouldEqual, 0)
				test.That(t, colorStruct.Fields["b"].GetNumberValue(), test.ShouldEqual, 0)
			},
		},
		{
			name: "defaults",
			pose: &commonPB.Pose{
				X:     50.0,
				Y:     75.0,
				Z:     100.0,
				OX:    1.0,
				OY:    0.0,
				OZ:    0.0,
				Theta: 90.0,
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, arrow, test.ShouldNotBeNil)
				test.That(t, arrow.ReferenceFrame, test.ShouldStartWith, "arrow-")
				test.That(t, arrow.PoseInObserverFrame.ReferenceFrame, test.ShouldEqual, "world")

				test.That(t, len(arrow.Uuid), test.ShouldEqual, 16)

				color, ok := arrow.Metadata.Fields["color"]
				test.That(t, ok, test.ShouldBeTrue)
				colorStruct := color.GetStructValue()
				test.That(t, colorStruct.Fields["r"].GetNumberValue(), test.ShouldEqual, defaultColor.R)
				test.That(t, colorStruct.Fields["g"].GetNumberValue(), test.ShouldEqual, defaultColor.G)
				test.That(t, colorStruct.Fields["b"].GetNumberValue(), test.ShouldEqual, defaultColor.B)
			},
		},
		{
			name:        "nil pose returns error",
			arrowName:   "test-arrow",
			pose:        nil,
			uuid:        nil,
			color:       nil,
			parentFrame: "",
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldEqual, "pose is required")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CreateArrow(tt.pose, tt.arrowName, tt.uuid, tt.color, tt.parentFrame)
			tt.expected(t, result, err)
		})
	}
}

func TestParseArrows(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected func(*testing.T, []*Arrow, error)
	}{
		{
			name: "valid array of arrows",
			input: []any{
				map[string]any{
					"name": "arrow1",
					"pose": map[string]any{
						"x":     100.0,
						"y":     200.0,
						"z":     300.0,
						"o_x":   0.0,
						"o_y":   0.0,
						"o_z":   1.0,
						"theta": 45.0,
					},
					"uuid":         "550e8400-e29b-41d4-a716-446655440000",
					"color":        map[string]any{"r": 255, "g": 0, "b": 0},
					"parent_frame": "robot",
				},
				map[string]any{
					"name": "arrow2",
					"pose": map[string]any{
						"x":     50.0,
						"y":     75.0,
						"z":     100.0,
						"o_x":   1.0,
						"o_y":   0.0,
						"o_z":   0.0,
						"theta": 90.0,
					},
				},
			},
			expected: func(t *testing.T, arrows []*Arrow, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, len(arrows), test.ShouldEqual, 2)

				test.That(t, arrows[0].ReferenceFrame, test.ShouldEqual, "arrow1")
				test.That(t, arrows[0].PoseInObserverFrame.ReferenceFrame, test.ShouldEqual, "robot")
				test.That(t, arrows[0].PoseInObserverFrame.Pose.X, test.ShouldEqual, 100.0)

				test.That(t, arrows[1].ReferenceFrame, test.ShouldEqual, "arrow2")
				test.That(t, arrows[1].PoseInObserverFrame.ReferenceFrame, test.ShouldEqual, "world")
				test.That(t, arrows[1].PoseInObserverFrame.Pose.X, test.ShouldEqual, 50.0)
			},
		},
		{
			name:  "empty array",
			input: []any{},
			expected: func(t *testing.T, arrows []*Arrow, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, len(arrows), test.ShouldEqual, 0)
			},
		},
		{
			name:  "not an array",
			input: "not an array",
			expected: func(t *testing.T, arrows []*Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Expected array of arrows")
				test.That(t, arrows, test.ShouldBeNil)
			},
		},
		{
			name: "array with invalid arrow",
			input: []any{
				map[string]any{
					"name": "valid-arrow",
					"pose": map[string]any{
						"x": 100.0,
						"y": 200.0,
						"z": 300.0,
					},
				},
				"not an arrow object",
			},
			expected: func(t *testing.T, arrows []*Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Failed to parse arrow at index 1")
				test.That(t, arrows, test.ShouldBeNil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseArrows(tt.input)
			tt.expected(t, result, err)
		})
	}
}

func TestParseArrow(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected func(*testing.T, *Arrow, error)
	}{
		{
			name: "valid arrow",
			input: map[string]any{
				"name": "test-arrow",
				"pose": map[string]any{
					"x":     100.0,
					"y":     200.0,
					"z":     300.0,
					"o_x":   0.0,
					"o_y":   0.0,
					"o_z":   1.0,
					"theta": 45.0,
				},
				"uuid":         "550e8400-e29b-41d4-a716-446655440000",
				"color":        map[string]any{"r": 255, "g": 0, "b": 0},
				"parent_frame": "robot",
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, arrow, test.ShouldNotBeNil)
				test.That(t, arrow.ReferenceFrame, test.ShouldEqual, "test-arrow")
				test.That(t, arrow.PoseInObserverFrame.ReferenceFrame, test.ShouldEqual, "robot")
				test.That(t, arrow.PoseInObserverFrame.Pose.X, test.ShouldEqual, 100.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.Y, test.ShouldEqual, 200.0)
				test.That(t, arrow.PoseInObserverFrame.Pose.Z, test.ShouldEqual, 300.0)

				test.That(t, len(arrow.Uuid), test.ShouldEqual, 16)

				color, ok := arrow.Metadata.Fields["color"]
				test.That(t, ok, test.ShouldBeTrue)
				colorStruct := color.GetStructValue()
				test.That(t, colorStruct.Fields["r"].GetNumberValue(), test.ShouldEqual, 255)
				test.That(t, colorStruct.Fields["g"].GetNumberValue(), test.ShouldEqual, 0)
				test.That(t, colorStruct.Fields["b"].GetNumberValue(), test.ShouldEqual, 0)
			},
		},
		{
			name: "defaults",
			input: map[string]any{
				"pose": map[string]any{
					"x":     100.0,
					"y":     200.0,
					"z":     300.0,
					"o_x":   0.0,
					"o_y":   0.0,
					"o_z":   1.0,
					"theta": 45.0,
				},
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, arrow, test.ShouldNotBeNil)
				test.That(t, arrow.ReferenceFrame, test.ShouldStartWith, "arrow-")

				test.That(t, len(arrow.Uuid), test.ShouldEqual, 16)

				color, ok := arrow.Metadata.Fields["color"]
				test.That(t, ok, test.ShouldBeTrue)

				colorStruct := color.GetStructValue()
				test.That(t, colorStruct.Fields["r"].GetNumberValue(), test.ShouldEqual, defaultColor.R)
				test.That(t, colorStruct.Fields["g"].GetNumberValue(), test.ShouldEqual, defaultColor.G)
				test.That(t, colorStruct.Fields["b"].GetNumberValue(), test.ShouldEqual, defaultColor.B)

				test.That(t, arrow.PoseInObserverFrame.ReferenceFrame, test.ShouldEqual, "world")
			},
		},
		{
			name:  "not an object",
			input: "not an arrow object",
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Expected arrow object")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
		{
			name: "missing pose field",
			input: map[string]any{
				"name": "test-arrow",
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Missing required 'pose' field")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
		{
			name: "invalid pose data",
			input: map[string]any{
				"name": "test-arrow",
				"pose": "not a pose object",
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Failed to parse pose")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
		{
			name: "invalid name type",
			input: map[string]any{
				"pose": map[string]any{"x": 100.0, "y": 200.0, "z": 300.0},
				"name": 123,
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Expected string for name")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
		{
			name: "invalid parent frame type",
			input: map[string]any{
				"name":         "test-arrow",
				"pose":         map[string]any{"x": 100.0, "y": 200.0, "z": 300.0},
				"parent_frame": 123,
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Expected string for parent frame")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
		{
			name: "invalid color data",
			input: map[string]any{
				"name":  "test-arrow",
				"pose":  map[string]any{"x": 100.0, "y": 200.0, "z": 300.0},
				"color": "not a color object",
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Failed to parse color")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
		{
			name: "invalid UUID format",
			input: map[string]any{
				"name": "test-arrow",
				"pose": map[string]any{"x": 100.0, "y": 200.0, "z": 300.0},
				"uuid": "invalid-uuid-format",
			},
			expected: func(t *testing.T, arrow *Arrow, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Failed to parse UUID")
				test.That(t, arrow, test.ShouldBeNil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseArrow(tt.input)
			tt.expected(t, result, err)
		})
	}
}
