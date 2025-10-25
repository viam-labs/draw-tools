// Package lib provides utility functions and types for writing visualizations.
package lib

type float interface {
	float32 | float64
}

func parseFloat[T float](val any, defaultValue T) T {
	if val == nil {
		return defaultValue
	}

	if f32, ok := val.(float32); ok {
		return T(f32)
	}
	if f64, ok := val.(float64); ok {
		return T(f64)
	}

	if i, ok := val.(int); ok {
		return T(i)
	}
	if i8, ok := val.(int8); ok {
		return T(i8)
	}
	if i16, ok := val.(int16); ok {
		return T(i16)
	}
	if i32, ok := val.(int32); ok {
		return T(i32)
	}
	if i64, ok := val.(int64); ok {
		return T(i64)
	}

	if u, ok := val.(uint); ok {
		return T(u)
	}
	if u8, ok := val.(uint8); ok {
		return T(u8)
	}
	if u16, ok := val.(uint16); ok {
		return T(u16)
	}
	if u32, ok := val.(uint32); ok {
		return T(u32)
	}
	if u64, ok := val.(uint64); ok {
		return T(u64)
	}

	return defaultValue
}

type integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

func parseInt[T integer](val any, defaultValue T) T {
	if val == nil {
		return defaultValue
	}

	if i, ok := val.(int); ok {
		return T(i)
	}
	if i8, ok := val.(int8); ok {
		return T(i8)
	}
	if i16, ok := val.(int16); ok {
		return T(i16)
	}
	if i32, ok := val.(int32); ok {
		return T(i32)
	}
	if i64, ok := val.(int64); ok {
		return T(i64)
	}

	if u, ok := val.(uint); ok {
		return T(u)
	}
	if u8, ok := val.(uint8); ok {
		return T(u8)
	}
	if u16, ok := val.(uint16); ok {
		return T(u16)
	}
	if u32, ok := val.(uint32); ok {
		return T(u32)
	}
	if u64, ok := val.(uint64); ok {
		return T(u64)
	}

	if f32, ok := val.(float32); ok {
		return T(f32)
	}
	if f64, ok := val.(float64); ok {
		return T(f64)
	}

	return defaultValue
}
