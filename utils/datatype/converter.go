package datatype

import (
	"fmt"
	"strconv"
)

func ParseInt(data interface{}) int {
	switch v := data.(type) {
	default:
		fmt.Printf("unexpected datatype %T", v)
	case string:
		intVal, _ := strconv.Atoi(data.(string))
		return intVal
	case int:
		return data.(int)
	case float64:
		return int(data.(float64))
	case float32:
		return int(data.(float32))
	}
	return 0
}
func ParseUint(data interface{}) uint {
	switch v := data.(type) {
	default:
		fmt.Printf("unexpected datatype %T", v)
	case string:
		intVal, _ := strconv.ParseUint(data.(string), 10, 32)
		return uint(intVal)
	case int:
		return data.(uint)
	case float64:
		return uint(data.(float64))
	case float32:
		return uint(data.(float32))
	}
	return 0
}
func ParseUint64(data interface{}) uint64 {
	switch v := data.(type) {
	default:
		fmt.Printf("unexpected datatype %T", v)
	case string:
		intVal, _ := strconv.ParseUint(data.(string), 10, 64)
		return intVal
	case int:
		return data.(uint64)
	case float64:
		return uint64(data.(float64))
	case float32:
		return uint64(data.(float32))
	}
	return 0
}
func ParseFloat32(data interface{}) float32 {
	switch v := data.(type) {
	default:
		fmt.Printf("unexpected datatype %T", v)
	case string:
		floatVal, _ := strconv.ParseFloat(data.(string), 32)
		return float32(floatVal)
	case int:
		return data.(float32)
	case float64:
		return data.(float32)
	case float32:
		return data.(float32)
	}
	return 0.0
}
func ParseFloat64(data interface{}) float64 {
	switch v := data.(type) {
	default:
		fmt.Printf("unexpected datatype %T", v)
	case string:
		floatVal, _ := strconv.ParseFloat(data.(string), 64)
		return floatVal
	case int:
		return data.(float64)
	case float64:
		return data.(float64)
	case float32:
		return data.(float64)
	}
	return 0.0
}
