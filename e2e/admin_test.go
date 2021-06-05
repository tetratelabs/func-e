// Copyright 2020 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

// newAdminClient returns a new client for Envoy Admin API.
func newAdminClient(address string) (*adminClient, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	return &adminClient{baseURL: fmt.Sprintf("http://%s:%s", host, port)}, nil
}

type adminClient struct {
	baseURL string
}

func (c *adminClient) isReady(ctx context.Context) bool {
	_, err := httpGet(ctx, c.baseURL+"/ready")
	return err == nil
}

func (c *adminClient) getMainListenerURL(ctx context.Context) (string, error) {
	var s map[string]interface{}
	if err := c.getJSON(ctx, "/listeners", &s); err != nil {
		return "", err
	}

	// The json structure is deep, so parsing instead of many nested structs
	for _, s := range s["listener_statuses"].([]interface{}) {
		l := s.(map[string]interface{})
		if l["name"] != "main" {
			continue
		}
		port := l["local_address"].(map[string]interface{})["socket_address"].(map[string]interface{})["port_value"]
		return fmt.Sprintf("http://127.0.0.1:%d", int(port.(float64))), nil
	}
	return "", fmt.Errorf("didn't find main port in %+v", s)
}

func (c *adminClient) getJSON(ctx context.Context, path string, v interface{}) error {
	body, err := httpGet(ctx, c.baseURL+path+"?format=json")
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
