package util

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	kratix "github.com/syntasso/kratix-go"
)

type AppRequestContext struct {
	Resource     kratix.Resource
	Name         string
	Namespace    string
	DBDriver     string
	VaultEnabled bool
}

type ResourceConfigureContext struct {
	AppRequestContext
	Image        string
	ResourceName string
	ServicePort  string
}

func ReadAppRequestContext(sdk *kratix.KratixSDK) (AppRequestContext, error) {
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return AppRequestContext{}, fmt.Errorf("read input: %w", err)
	}

	return AppRequestContext{
		Resource:     res,
		Name:         mustString(get(res, "metadata.name")),
		Namespace:    mustString(get(res, "metadata.namespace")),
		DBDriver:     mustStringOrEmpty(get(res, "spec.dbDriver")),
		VaultEnabled: hasLabelTrue(res, VaultOptInLabelKey),
	}, nil
}

func ReadResourceConfigureContext(sdk *kratix.KratixSDK) (ResourceConfigureContext, error) {
	appCtx, err := ReadAppRequestContext(sdk)
	if err != nil {
		return ResourceConfigureContext{}, err
	}

	return ResourceConfigureContext{
		AppRequestContext: appCtx,
		Image:             mustString(get(appCtx.Resource, "spec.image")),
		ResourceName:      mustString(get(appCtx.Resource, "metadata.name")),
		ServicePort:       mustString(get(appCtx.Resource, "spec.service.port")),
	}, nil
}

func hasLabelTrue(res kratix.Resource, label string) bool {
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

func get(res kratix.Resource, path string) any {
	v, err := res.GetValue(path)
	if err != nil {
		log.Fatalf("get %s: %v", path, err)
	}
	return v
}

func mustString(v any) string {
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

func mustStringOrEmpty(v any) string {
	if v == nil {
		return ""
	}

	switch t := v.(type) {
	case string:
		return t
	default:
		return mustString(v)
	}
}
