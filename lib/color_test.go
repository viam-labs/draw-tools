package lib

import (
	"testing"

	"go.viam.com/test"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		name         string
		colorData    any
		defaultValue Color
		expected     func(*testing.T, Color, error)
	}{
		{
			name: "valid color",
			colorData: map[string]any{
				"r": 255,
				"g": 128,
				"b": 64,
			},
			defaultValue: Color{R: 0, G: 0, B: 0},
			expected: func(t *testing.T, color Color, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, color.R, test.ShouldEqual, 255)
				test.That(t, color.G, test.ShouldEqual, 128)
				test.That(t, color.B, test.ShouldEqual, 64)
			},
		},
		{
			name: "default values",
			colorData: map[string]any{
				"r": 150,
				"g": nil,
				"b": 200,
			},
			defaultValue: Color{R: 10, G: 20, B: 30},
			expected: func(t *testing.T, color Color, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, color.R, test.ShouldEqual, 150)
				test.That(t, color.G, test.ShouldEqual, 20) // default value
				test.That(t, color.B, test.ShouldEqual, 200)
			},
		},
		{
			name: "clamped values",
			colorData: map[string]any{
				"r": 300,
				"g": -50,
				"b": 128,
			},
			defaultValue: Color{R: 0, G: 0, B: 0},
			expected: func(t *testing.T, color Color, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, color.R, test.ShouldEqual, 255)
				test.That(t, color.G, test.ShouldEqual, 0)
				test.That(t, color.B, test.ShouldEqual, 128)
			},
		},
		{
			name:         "nil color map returns error",
			colorData:    nil,
			defaultValue: Color{R: 100, G: 150, B: 200},
			expected: func(t *testing.T, color Color, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "expected color object")
				test.That(t, color.R, test.ShouldEqual, 100) // should return default
				test.That(t, color.G, test.ShouldEqual, 150)
				test.That(t, color.B, test.ShouldEqual, 200)
			},
		},
		{
			name:         "not a map returns error",
			colorData:    "not a color object",
			defaultValue: Color{R: 0, G: 0, B: 0},
			expected: func(t *testing.T, color Color, err error) {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "expected color object")
				test.That(t, color.R, test.ShouldEqual, 0) // should return default
				test.That(t, color.G, test.ShouldEqual, 0)
				test.That(t, color.B, test.ShouldEqual, 0)
			},
		},
		{
			name:         "empty color map uses defaults",
			colorData:    map[string]any{},
			defaultValue: Color{R: 50, G: 100, B: 150},
			expected: func(t *testing.T, color Color, err error) {
				test.That(t, err, test.ShouldBeNil)
				test.That(t, color.R, test.ShouldEqual, 50)
				test.That(t, color.G, test.ShouldEqual, 100)
				test.That(t, color.B, test.ShouldEqual, 150)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseColor(tt.colorData, tt.defaultValue)
			tt.expected(t, result, err)
		})
	}
}
