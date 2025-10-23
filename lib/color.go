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

// GetProgressColor maps progress to color spectrum using RGB interpolation
//
//	0.0 -> 0.2: Blue to Cyan (0,0,255) -> (0,255,255)
//	0.2 -> 0.4: Cyan to Green (0,255,255) -> (0,255,0)
//	0.4 -> 0.6: Green to Yellow (0,255,0) -> (255,255,0)
//	0.6 -> 0.8: Yellow to Orange (255,255,0) -> (255,165,0)
//	0.8 -> 1.0: Orange to Red (255,165,0) -> (255,0,0)
func GetProgressColor(current, total int) Color {
	if total <= 1 {
		return Color{R: 0, G: 0, B: 255}
	}

	progress := float64(current) / float64(total-1)
	if progress <= 0.2 {
		mod := progress / 0.2
		return Color{
			R: 0,
			G: uint8(255 * mod),
			B: 255,
		}
	} else if progress <= 0.4 {
		mod := (progress - 0.2) / 0.2
		return Color{
			R: 0,
			G: 255,
			B: uint8(255 * (1 - mod)),
		}
	} else if progress <= 0.6 {
		mod := (progress - 0.4) / 0.2
		return Color{
			R: uint8(255 * mod),
			G: 255,
			B: 0,
		}
	} else if progress <= 0.8 {
		mod := (progress - 0.6) / 0.2
		return Color{
			R: 255,
			G: uint8(255 - 90*mod),
			B: 0,
		}
	} else {
		mod := (progress - 0.8) / 0.2
		return Color{
			R: 255,
			G: uint8(165 - 165*mod),
			B: 0,
		}
	}
}
