package lib

func ParseFloat64(val any, defaultValue float64) float64 {
	if val == nil {
		return defaultValue
	}

	if f, ok := val.(float64); ok {
		return f
	}

	return defaultValue
}
