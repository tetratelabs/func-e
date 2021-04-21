// Copyright 2019 Tetrate
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

package pii

import (
	"crypto/sha256"
	"fmt"

	"bitbucket.org/creachadair/shell"
	l "github.com/tetratelabs/log"
)

var logger = l.RegisterScope("pkg/pii", "filters log lines for PII data", 0)

var (
	istioFormat = `[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"` +
		` %RESPONSE_CODE% %RESPONSE_FLAGS% "%DYNAMIC_METADATA(istio.mixer:status)%" "%REQ(USER-AGENT)%"`
	defaultPII = map[string]bool{
		// keys without variables
		`[%START_TIME%]`:                           true,
		`[%BYTES_RECEIVED%]`:                       true,
		`%PROTOCOL%`:                               true,
		`%RESPONSE_CODE%`:                          true,
		`%RESPONSE_CODE_DETAILS%`:                  true,
		`%BYTES_SENT%`:                             true,
		`%DURATION%`:                               true,
		`%RESPONSE_DURATION%`:                      true,
		`%RESPONSE_FLAGS%`:                         true,
		`%RESPONSE_TX_DURATION%`:                   true,
		`%ROUTE_NAME%`:                             true,
		`%UPSTREAM_HOST%`:                          true,
		`%UPSTREAM_CLUSTER%`:                       true,
		`%UPSTREAM_LOCAL_ADDRESS%`:                 true,
		`%UPSTREAM_TRANSPORT_FAILURE_REASON%`:      true,
		`%DOWNSTREAM_REMOTE_ADDRESS%`:              true,
		`%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%`: true,
		`%DOWNSTREAM_LOCAL_ADDRESS%`:               true,
		`%DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT%`:  true,
		`%REQUESTED_SERVER_NAME%`:                  true,
		`%DOWNSTREAM_LOCAL_URI_SAN%`:               true,
		`%DOWNSTREAM_PEER_URI_SAN%`:                true,
		`%DOWNSTREAM_LOCAL_SUBJECT%`:               true,
		`%DOWNSTREAM_PEER_SUBJECT%`:                true,
		`%DOWNSTREAM_PEER_ISSUER%`:                 true,
		`%DOWNSTREAM_TLS_SESSION_ID%`:              true,
		`%DOWNSTREAM_TLS_CIPHER%`:                  true,
		`%DOWNSTREAM_TLS_VERSION%`:                 true,
		`%DOWNSTREAM_PEER_FINGERPRINT_256%`:        true,
		`%DOWNSTREAM_PEER_SERIAL%`:                 true,
		`%DOWNSTREAM_PEER_CERT%`:                   true,
		`%DOWNSTREAM_PEER_CERT_V_START%`:           true,
		`%DOWNSTREAM_PEER_CERT_V_END%`:             true,
		// keys with variables
		`%REQ(X?Y):Z%`:                         true,
		`%RESP(X?Y):Z%`:                        true,
		`%TRAILER(X?Y):Z%`:                     true,
		`%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%`: true,
	}
)

// Filter filters log fields using pii and modify all PII fields using f
type Filter struct {
	f      func(string) string
	pii    map[string]bool
	format []string
}

// NewFilter constructs a custom filter object
func NewFilter(formatStr string, hash func(string) string, pii map[string]bool) (Filter, error) {
	// splitting formats and handle error
	format, ok := shell.Split(formatStr)
	if !ok {
		return Filter{}, fmt.Errorf("error in splitting format string: %s", format)
	}
	return Filter{f: hash, pii: pii, format: format}, nil
}

// Default creates a filter with default fields for istio
func Default() (Filter, error) {
	return NewFilter(istioFormat, defaultHash, defaultPII)
}

// Process logs and hash the filter, assumes that filter has valid fields
func (f Filter) Process(logs []string) []string {
	out := make([]string, 0, len(logs))
	for _, log := range logs {
		fieldValues, ok := shell.Split(log)
		if !ok {
			if logger.DebugEnabled() {
				logger.Debugf("error splitting log, skipping: %s", log)
			} else {
				logger.Info("error splitting log, skipping")
			}
			continue
		}

		if len(fieldValues) == len(f.format) {
			// pick the PII fields and Hash the fields
			for j, name := range f.format {
				// TODO: may require further parsing name to verify if it matches as a key in the map with variables
				if f.pii[name] {
					fieldValues[j] = f.f(fieldValues[j])
				}
			}
			out = append(out, shell.Join(fieldValues))
		}
	}
	return out
}

// defaultHash returns the hashed value of s using sha256 defaultHash function
// TODO: salt the Hash
func defaultHash(s string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(s))
	return string(h.Sum(nil))
}
