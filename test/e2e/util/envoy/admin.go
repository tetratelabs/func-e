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

package envoy

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

// AdminAPI represents Envoy Admin API.
type AdminAPI interface {
	IsReady() (bool, error)
	GetStats() (*Stats, error)
}

// Stats represents Envoy response to `/stats?format=json` endpoint.
type Stats struct {
	Metrics []Metric `json:"stats"`
}

// Metric represents recorded value of a single metric.
type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// GetMetric returns a metric by name.
func (s *Stats) GetMetric(name string) *Metric {
	for i := range s.Metrics {
		if s.Metrics[i].Name == name {
			return &s.Metrics[i]
		}
	}
	return nil
}

// NewClient returns a new client for Envoy Admin API.
func NewClient(address string) (AdminAPI, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	return &client{baseURL: fmt.Sprintf("http://%s:%s", host, port)}, nil
}

type client struct {
	baseURL string
}

func (c *client) IsReady() (bool, error) {
	resp, err := http.Get(c.baseURL + "/ready")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() //nolint:errcheck
	return resp.StatusCode == http.StatusOK, nil
}

func (c *client) GetStats() (*Stats, error) {
	resp, err := http.Get(c.baseURL + "/stats?format=json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var stats Stats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
