package lib

import (
	"testing"

	commonPB "go.viam.com/api/common/v1"
	"go.viam.com/test"
)

func TestParsePoseJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected func(*testing.T, *commonPB.Pose, error)
	}{
		{
			name: "valid pose",
			input: map[string]any{
				"x":     100.0,
				"y":     200.0,
				"z":     300.0,
				"o_x":   0.0,
				"o_y":   0.0,
				"o_z":   1.0,
				"theta": 45.0,
			},
			expected: func(t *testing.T, pose *commonPB.Pose, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, pose.X, test.ShouldEqual, 100.0)
				test.That(t, pose.Y, test.ShouldEqual, 200.0)
				test.That(t, pose.Z, test.ShouldEqual, 300.0)
				test.That(t, pose.OX, test.ShouldEqual, 0.0)
				test.That(t, pose.OY, test.ShouldEqual, 0.0)
				test.That(t, pose.OZ, test.ShouldEqual, 1.0)
				test.That(t, pose.Theta, test.ShouldEqual, 45.0)
			},
		},
		{
			name:  "not a map",
			input: "not a pose object",
			expected: func(t *testing.T, pose *commonPB.Pose, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "expected pose object")
			},
		},
		{
			name:  "nil input",
			input: nil,
			expected: func(t *testing.T, pose *commonPB.Pose, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "expected pose object")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePose(tt.input)
			tt.expected(t, result, err)
		})
	}
}

func TestPoseToMeters(t *testing.T) {
	t.Run("basic conversion from millimeters to meters", func(t *testing.T) {
		pose := PoseToMeters(&commonPB.Pose{
			X:     1000.0,
			Y:     2000.0,
			Z:     3000.0,
			OX:    0.0,
			OY:    0.0,
			OZ:    1.0,
			Theta: 45.0,
		})

		test.That(t, pose.X, test.ShouldEqual, 1.0)
		test.That(t, pose.Y, test.ShouldEqual, 2.0)
		test.That(t, pose.Z, test.ShouldEqual, 3.0)
		test.That(t, pose.OX, test.ShouldEqual, 0.0)
		test.That(t, pose.OY, test.ShouldEqual, 0.0)
		test.That(t, pose.OZ, test.ShouldEqual, 1.0)
		test.That(t, pose.Theta, test.ShouldEqual, 45.0)
	})
}
