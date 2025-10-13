package lib

import "fmt"

type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

func ParseColor(colorData any, defaultValue Color) (Color, error) {
	colorMap, ok := colorData.(map[string]any)
	if !ok {
		return defaultValue, fmt.Errorf("expected color object, got %T", colorData)
	}

	if colorMap == nil {
		return defaultValue, nil
	}

	r := ParseFloat64(colorMap["r"], float64(defaultValue.R))
	g := ParseFloat64(colorMap["g"], float64(defaultValue.G))
	b := ParseFloat64(colorMap["b"], float64(defaultValue.B))

	return Color{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}, nil
}
