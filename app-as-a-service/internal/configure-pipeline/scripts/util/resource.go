package util

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	kratix "github.com/syntasso/kratix-go"
)

func Get(res kratix.Resource, path string) any {
	v, err := res.GetValue(path)
	if err != nil {
		log.Fatalf("get %s: %v", path, err)
	}
	return v
}

func MustString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		log.Fatalf("want string got %T", v)
		return ""
	}
}

func MustStringOrEmpty(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return MustString(v)
	}
}

func HasLabelTrue(res kratix.Resource, label string) bool {
	v, err := res.GetValue("metadata.labels")
	if err != nil || v == nil {
		return false
	}

	switch labels := v.(type) {
	case map[string]any:
		return isTruthy(labels[label])
	case map[string]string:
		value, ok := labels[label]
		if !ok {
			return false
		}
		return strings.EqualFold(strings.TrimSpace(value), "true")
	default:
		return false
	}
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}

	switch value := v.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true")
	default:
		return strings.EqualFold(strings.TrimSpace(fmt.Sprintf("%v", value)), "true")
	}
}
