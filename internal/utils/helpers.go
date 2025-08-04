package utils

import (
	"fmt"
	"time"
)

// Helper functions for type conversion
func GetString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func GetFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func GetTime(v interface{}) time.Time {
	switch val := v.(type) {
	case float64:
		return time.Unix(int64(val/1000), 0)
	case int64:
		return time.Unix(val/1000, 0)
	case string:
		t, err := time.Parse(time.RFC3339, val)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

// GetBool chuyển interface{} sang bool
func GetBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val == "true" || val == "1"
	default:
		return false
	}
}
