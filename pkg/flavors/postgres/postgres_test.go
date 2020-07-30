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
package postgres

import (
	"strings"
	"testing"

	"github.com/tetratelabs/getenvoy/pkg/flavors"
)

// Test verifies that postgres flavor is registered when module is loaded.
func TestInit(t *testing.T) {
	_, err := flavors.GetFlavor("postgres")

	if err != nil {
		t.Error("postgres flavor should be registered at init phase.")
	}
}

type testInputEndpoint struct {
	endpoints   string
	parseResult clusterEndpoint
	result      bool
}

// Test parsing a single endpoint, which is in one of the following forms:
// IP:port
// IP
// name:Port
// name
//
// If port is not specified in the endpoint it is assumed to be 5432
func TestParseSingleEndpoint(t *testing.T) {
	input := []testInputEndpoint{
		{"127.0.0.1:12:34", clusterEndpoint{}, false},
		{"127.0.0.1:blah", clusterEndpoint{}, false},
		{"127.0.0.1:12", clusterEndpoint{"127.0.0.1", "12", true}, true},
		{"127.0.0.1", clusterEndpoint{"127.0.0.1", "5432", true}, true},
		{"127.0", clusterEndpoint{"127.0", "5432", false}, true},
		{"127.0.0.0.1", clusterEndpoint{"127.0.0.0.1", "5432", false}, true},
		{"127.0.", clusterEndpoint{"127.0.", "5432", false}, true},
		{"postgres.com:34:12", clusterEndpoint{}, false},
		{"postgres.com:12", clusterEndpoint{"postgres.com", "12", false}, true},
		{"postgres", clusterEndpoint{"postgres", "5432", false}, true},
	}

	for _, testCase := range input {
		parsed, err := parseSingleEndpoint(testCase.endpoints)
		if testCase.result != (err == nil) {
			t.Errorf("Parsing result for %s not as expected", testCase.endpoints)
		}
		if err != nil {
			continue
		}
		if *parsed != testCase.parseResult {
			t.Errorf("Parsed structure %v for %s is not as expected", *parsed, testCase.endpoints)
		}
	}
}

// Structure is used for parameterized test
// testing parsing comma delimited list of single endpoints.
type testInputEndpointSet struct {
	// entry value - list of endpoints
	endpointset string
	// Expected parsing result. True: success.
	result bool
}

// Endpoints are passed from command line as comma delimited string of individual endpoints.
// This test verifies that the string is tokenized and parsed correctly.
func TestParseEndpointSet(t *testing.T) {
	input := []testInputEndpointSet{
		{"127.0.0.1:3456", true},
		{"127.0.0.1:blah", false},
		{"127.0.0.1:3456, 127.0.0.1:5555", true},
		{"127.0.0.1:3456127.0.0.1:5555", false},
		{"127.0:5555", true},
		{"postgres,127.0.0.1:3456,127.0.0.1:5555", true},
		{"postgres:3456127.0.0.1:5555", false},
		{"postgres:3456,postgres:5555", true},
		{"127.0:3456,127.0.0.1.1.25555", true},
		// IP address and host name should not be mixed
		{"127.0.0.1:3456,postgres:2555", true},
	}

	for _, testCase := range input {
		_, err := parseEndpointSet(testCase.endpointset)

		if testCase.result != (err == nil) {
			t.Errorf("Parsing result of endpointset %s no as expected: %s", testCase.endpointset, err)
		}
	}
}

// Structure is used for parameterized testing of input params parsing
// end verification
type testInputCmdParams struct {
	params map[string]string
	result bool
}

// Test verifies that parsing input parameters should fail
// when endpoint is not specified and when specified port has wrong format.
func TestInputCmdParams(t *testing.T) {
	var testFlavor Flavor
	input := []testInputCmdParams{
		{map[string]string{"endpoints": "127.0.0.1:3456"}, true},
		{map[string]string{"endpoints1": "127.0.0.1:3456"}, false},
		{map[string]string{"endpoints1": "127.0.0.1:3456", "endpoints": "128.0.0.1"}, true},
		{map[string]string{"endpoints1": "127.0.0.1:3456", "inport": "128.0.0.1"}, false},
		{map[string]string{"endpoints": "127.0.0.1:3456", "inport": "128.0.0.1"}, false},
		{map[string]string{"endpoints": "127.0.0.1:3456", "inport": "5432"}, true},
		{map[string]string{"endpoints": "127.0.0.1:blah", "inport": "5432"}, false},
		{map[string]string{"endpoints": "127.0.0.1:3456,postgres:1234"}, false},
	}

	for _, testCase := range input {
		testFlavor.endpoints = testFlavor.endpoints[:0]
		err := testFlavor.parseInputParams(testCase.params)

		if testCase.result != (err == nil) {
			t.Errorf("Parsing input params %v not as expected", testCase.params)
		}
	}
}

// Structure is used for parameterized testing of creating
// endpoints part of Envoy config
type testEndpointSetConfig struct {
	// Input command line params
	params map[string]string
	// What must be found after processing template
	output []string
}

// Test creating set of endpoint.
// Test only verifies that template substitution happens.
// Syntax and yaml formatting is not checked.
func TestCreateEndpointsConfig(t *testing.T) {
	var testFlavor Flavor

	input := []testEndpointSetConfig{
		{map[string]string{"endpoints": "127.0.0.1:3456"}, []string{"127.0.0.1", "3456"}},
		{map[string]string{"endpoints": "127.0.0.1:3456,128.0.0.1"}, []string{"127.0.0.1", "3456", "128.0.0.1", "5432"}},
		{map[string]string{"endpoints": "postgres:3456"}, []string{"postgres", "3456"}},
		{map[string]string{"endpoints": "postgres1:3456,postgres2"}, []string{"postgres1", "3456", "postgres2", "5432"}},
	}

	for index, testCase := range input {
		testFlavor.endpoints = testFlavor.endpoints[:0]
		testFlavor.parseInputParams(testCase.params)

		endpointsConfig, err := testFlavor.generateEndpointSetConfig()

		if err != nil {
			t.Errorf("Error creating config for testcase %d: %s", index, err)
			continue
		}

		// Scan created config for input params
		for _, item := range testCase.output {
			if !strings.Contains(endpointsConfig, item) {
				t.Errorf("Created config %s\n does not contain %s", endpointsConfig, item)
			}
		}
	}
}

// Test verifies that correct cluster is created based on passed endpoint types
// All IP addresses will create STATIC cluster, all hostnames will create STRICT DNS
// cluster.
func TestCreateMainConfig(t *testing.T) {
	var testFlavor Flavor

	input := []testEndpointSetConfig{
		{map[string]string{"endpoints": "127.0.0.1:3456"}, []string{"static"}},
		{map[string]string{"endpoints": "127.0.0.1:3456,128.0.0.1"}, []string{"static"}},
		{map[string]string{"endpoints": "postgres:3456"}, []string{"strict_dns"}},
		{map[string]string{"endpoints": "postgres1:3456,postgres2"}, []string{"strict_dns"}},
	}

	for index, testCase := range input {
		testFlavor.endpoints = testFlavor.endpoints[:0]
		testFlavor.parseInputParams(testCase.params)

		endpointsConfig, err := testFlavor.generateMainConfig()

		if err != nil {
			t.Errorf("Error creating config for testcase %d: %s", index, err)
			continue
		}

		// Scan created config for input params
		for _, item := range testCase.output {
			if !strings.Contains(endpointsConfig, item) {
				t.Errorf("Created config %s\n does not contain %s", endpointsConfig, item)
			}
		}
	}
}
