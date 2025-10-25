package lib

import "fmt"

// Color represents an RGB color value with red, green, and blue components.
// Each component is an 8-bit unsigned integer (0-255).
type Color struct {
	R uint8 `json:"r"` // Red component (0-255)
	G uint8 `json:"g"` // Green component (0-255)
	B uint8 `json:"b"` // Blue component (0-255)
}

// ParseColor parses a color from JSON data with validation and clamping.
// It handles various numeric types and clamps RGB values to the valid range (0-255).
//
// Parameters:
//   - colorData: JSON object containing color data (r, g, b fields)
//   - defaultValue: Default color to use for missing values
//
// Returns the parsed color with values clamped to 0-255 range or an error if parsing fails.
func ParseColor(colorData any, defaultValue Color) (Color, error) {
	colorMap, ok := colorData.(map[string]any)
	if !ok {
		return defaultValue, fmt.Errorf("expected color object, got %T", colorData)
	}

	if colorMap == nil {
		return defaultValue, nil
	}

	r := parseInt(colorMap["r"], int(defaultValue.R))
	if r < 0 {
		r = 0
	} else if r > 255 {
		r = 255
	}

	g := parseInt(colorMap["g"], int(defaultValue.G))
	if g < 0 {
		g = 0
	} else if g > 255 {
		g = 255
	}

	b := parseInt(colorMap["b"], int(defaultValue.B))
	if b < 0 {
		b = 0
	} else if b > 255 {
		b = 255
	}

	return Color{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}, nil
}
