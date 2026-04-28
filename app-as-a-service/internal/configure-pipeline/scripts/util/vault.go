package util

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func VaultDeleteIgnoreMissing(vaultAddr, vaultToken, path string) error {
	url := fmt.Sprintf("%s/v1/%s", strings.TrimRight(vaultAddr, "/"), path)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", vaultToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("vault delete failed for %s with HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(body)))
}

func VaultRevokePrefixIgnoreMissing(vaultAddr, vaultToken, payload string) error {
	url := fmt.Sprintf("%s/v1/sys/leases/revoke-prefix", strings.TrimRight(vaultAddr, "/"))
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(payload))
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", vaultToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("vault lease revoke-prefix failed with HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}
