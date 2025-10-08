package lib

import "fmt"

type Color struct {
	R uint8
	G uint8
	B uint8
}

func ParseColor(colorData any) (Color, error) {
	colorMap, ok := colorData.(map[string]any)
	if !ok {
		return Color{}, fmt.Errorf("expected color object, got %T", colorData)
	}

	r := GetFloat64(colorMap["r"], 0.0)
	g := GetFloat64(colorMap["g"], 0.0)
	b := GetFloat64(colorMap["b"], 0.0)

	return Color{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}, nil
}
