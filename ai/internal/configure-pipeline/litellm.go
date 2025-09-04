package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type keyGenerateResp struct {
	Key   string    `json:"key"`
	Error *apiError `json:"error"`
}

func generateTeamAndKey(auth, team, tier string, models []string) string {
	baseURL := "http://litellm.default.svc.cluster.local:4000"
	keyAlias := team + "-kratix-key"
	duration := "2d"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	hc := &http.Client{Timeout: 15 * time.Second}

	teamID, err := findTeamID(ctx, hc, baseURL, auth, team)
	if err != nil {
		log.Fatalf("list teams: %v", err)
	}
	if teamID == "" {
		teamID, err = createTeam(ctx, hc, baseURL, auth, team)
		if err != nil || teamID == "" {
			log.Fatalf("create team: %v", err)
		}
	}

	resp, err := generateKey(ctx, hc, baseURL, auth, teamID, keyAlias, duration, models)
	if err != nil {
		log.Fatalf("generate key: %v", err)
	}
	if resp.Error != nil && (resp.Error.Code == "400" || strings.Contains(strings.ToLower(resp.Error.Message), "already exists")) {
		_ = deleteKeyByAlias(ctx, hc, baseURL, auth, keyAlias)
		resp, err = generateKey(ctx, hc, baseURL, auth, teamID, keyAlias, duration, models)
		if err != nil {
			log.Fatalf("re-generate key: %v", err)
		}
	}

	virtualKey := resp.Key
	if virtualKey == "" {
		b, _ := json.Marshal(resp)
		log.Fatalf("failed to obtain key. response=%s", string(b))
	}

	return virtualKey
}

func findTeamID(ctx context.Context, hc *http.Client, baseURL, auth, team string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/team/list", nil)
	req.Header.Set("Authorization", "Bearer "+auth)
	res, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var any any
	if err := json.Unmarshal(body, &any); err != nil {
		return "", err
	}
	objects := normalizeToObjects(any)
	for _, obj := range objects {
		alias := firstString(obj, "team_alias", "alias", "name")
		if alias == team {
			return firstString(obj, "team_id", "id"), nil
		}
	}
	return "", nil
}

func createTeam(ctx context.Context, hc *http.Client, baseURL, auth, team string) (string, error) {
	payload := map[string]any{
		"team_alias": team,
		"metadata":   map[string]any{"owner": "jake"},
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/team/new", strings.NewReader(string(b)))
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Content-Type", "application/json")
	res, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	id := firstString(out, "team_id", "id")
	if id == "" {
		return "", errors.New("no team_id in response")
	}
	return id, nil
}

func generateKey(ctx context.Context, hc *http.Client, baseURL, auth, teamID, keyAlias, duration string, models []string) (*keyGenerateResp, error) {
	payload := map[string]any{
		"team_id":   teamID,
		"key_alias": keyAlias,
		"duration":  duration,
		"models":    models,
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/key/generate", strings.NewReader(string(b)))
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Content-Type", "application/json")

	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var out keyGenerateResp
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteKeyByAlias(ctx context.Context, hc *http.Client, baseURL, auth, keyAlias string) error {
	payload := map[string]any{"key_aliases": []string{keyAlias}}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/key/delete", strings.NewReader(string(b)))
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Content-Type", "application/json")
	res, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func mustParseModels(s string) []string {
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err == nil && len(arr) > 0 {
		return arr
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(p, `"'`))
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		out = []string{"local-tiny"}
	}
	return out
}

func normalizeToObjects(raw interface{}) []map[string]interface{} {
	switch v := raw.(type) {
	case []interface{}:
		return coerceSlice(v)
	case map[string]interface{}:
		for _, k := range []string{"teams", "data", "items"} {
			if inner, ok := v[k]; ok {
				return normalizeToObjects(inner)
			}
		}
		return []map[string]interface{}{v}
	default:
		return nil
	}
}

func coerceSlice(in []interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(in))
	for _, it := range in {
		if m, ok := it.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}

func firstString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func getenv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}
