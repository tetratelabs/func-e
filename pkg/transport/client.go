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

package transport

import (
	"fmt"
	"net/http"

	"github.com/tetratelabs/getenvoy/pkg/version"
)

var (
	cliUserAgent  = fmt.Sprintf("GetEnvoy/%s", version.Version)
	defaultClient = NewClient(AddUserAgent(cliUserAgent))
)

// Options represents an argument of NewClient
type Option func(http.RoundTripper) http.RoundTripper

// NewClient returns HTTP client for use of GetEnvoy CLI.
func NewClient(opts ...Option) *http.Client {
	tr := http.DefaultTransport
	for _, opt := range opts {
		tr = opt(tr)
	}
	client := &http.Client{Transport: tr}
	return client
}

// AddUserAgent returns Option that adds passed user-agent to every requests.
// It should be passed as an argument of NewClient.
func AddUserAgent(ua string) Option {
	return func(tr http.RoundTripper) http.RoundTripper {
		return &funcTripper{roundTrip: func(r *http.Request) (*http.Response, error) {
			r.Header.Add("User-Agent", ua)
			return tr.RoundTrip(r)
		}}
	}
}

type funcTripper struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (f funcTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f.roundTrip(r)
}

// Get is thin wrapper of net/http.Get.
func Get(url string) (*http.Response, error) {
	return defaultClient.Get(url)
}
