package lib

func GetFloat64(val any, defaultValue float64) float64 {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}

	return defaultValue
}
