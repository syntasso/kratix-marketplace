package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const baseURL = "http://litellm.default.svc.cluster.local:4000"

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type keyGenerateResp struct {
	Key   string    `json:"key"`
	Error *apiError `json:"error"`
}

type llmClient struct {
	auth string
	hc   *http.Client
}

type tierCfg struct {
	RPM            int
	TPM            int
	Budget         float64
	BudgetDuration string
}

func generateTeamAndKey(auth, team, tier string, models []string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := newClient(auth)

	teamID, err := client.findTeamID(ctx, team)
	if err != nil {
		log.Fatalf("list teams: %v", err)
	}
	if teamID == "" {
		teamID, err = client.createTeam(ctx, team, tier)
		if err != nil || teamID == "" {
			log.Fatalf("create team: %v", err)
		}
	}

	keyAlias := team + "-kratix-key"
	resp, err := client.generateKey(ctx, teamID, keyAlias, "2d", models)
	if err != nil {
		log.Fatalf("generate key: %v", err)
	}
	if resp.Error != nil && (resp.Error.Code == "400" || strings.Contains(strings.ToLower(resp.Error.Message), "already exists")) {
		_ = client.deleteKeyByAlias(ctx, keyAlias)
		resp, err = client.generateKey(ctx, teamID, keyAlias, "2d", models)
		if err != nil {
			log.Fatalf("re-generate key: %v", err)
		}
	}

	if resp.Key == "" {
		b, _ := json.Marshal(resp)
		log.Fatalf("failed to obtain key. response=%s", string(b))
	}
	return resp.Key
}

func (c *llmClient) findTeamID(ctx context.Context, team string) (string, error) {
	var raw any
	if err := c.get(ctx, "/team/list", &raw); err != nil {
		return "", err
	}
	for _, obj := range normalizeToObjects(raw) {
		if firstString(obj, "team_alias", "alias", "name") == team {
			return firstString(obj, "team_id", "id"), nil
		}
	}
	return "", nil
}

func (c *llmClient) createTeam(ctx context.Context, tier, team string) (string, error) {
	limits := tierLimits(tier) // pick RPM, TPM, budget based on tier

	payload := map[string]any{
		"team_alias":      team,
		"metadata":        map[string]any{"owner": os.Getenv("USER")},
		"rpm_limit":       limits.RPM,
		"tpm_limit":       limits.TPM,
		"max_budget":      limits.Budget,
		"budget_duration": limits.BudgetDuration, // e.g. "1d"
	}

	var out map[string]any
	if err := c.post(ctx, "/team/new", payload, &out); err != nil {
		return "", err
	}
	id := firstString(out, "team_id", "id")
	if id == "" {
		return "", errors.New("no team_id in response")
	}
	return id, nil
}

func tierLimits(tier string) tierCfg {
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case "small":
		return tierCfg{
			RPM: 60, TPM: 30_000,
			Budget: 10, BudgetDuration: "1d",
		}
	case "medium":
		return tierCfg{
			RPM: 240, TPM: 120_000,
			Budget: 100, BudgetDuration: "1d",
		}
	case "large":
		return tierCfg{
			RPM: 1200, TPM: 600_000,
			Budget: 1000, BudgetDuration: "1d",
		}
	default:
		// sane defaults if unknown tier
		return tierCfg{
			RPM: 120, TPM: 60_000,
			Budget: 25, BudgetDuration: "1d",
		}
	}
}

func (c *llmClient) generateKey(ctx context.Context, teamID, keyAlias, duration string, models []string) (*keyGenerateResp, error) {
	payload := map[string]any{
		"team_id":   teamID,
		"key_alias": keyAlias,
		"duration":  duration,
		"models":    models,
	}
	var out keyGenerateResp
	if err := c.post(ctx, "/key/generate", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *llmClient) deleteKeyByAlias(ctx context.Context, keyAlias string) error {
	payload := map[string]any{"key_aliases": []string{keyAlias}}
	return c.post(ctx, "/key/delete", payload, nil)
}

func newClient(auth string) *llmClient {
	return &llmClient{
		auth: auth,
		hc:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *llmClient) get(ctx context.Context, path string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.auth)
	res, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(out)
}

func (c *llmClient) post(ctx context.Context, path string, payload any, out any) error {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.auth)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if out == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// helpers

func normalizeToObjects(raw any) []map[string]any {
	switch v := raw.(type) {
	case []any:
		return coerceSlice(v)
	case map[string]any:
		for _, k := range []string{"teams", "data", "items"} {
			if inner, ok := v[k]; ok {
				return normalizeToObjects(inner)
			}
		}
		return []map[string]any{v}
	default:
		return nil
	}
}

func coerceSlice(in []any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, it := range in {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
