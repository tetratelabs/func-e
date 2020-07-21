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
	"github.com/tetratelabs/getenvoy/pkg/flavors"
	"testing"
)

// Test verifies that postgres flavor is registered when module is loaded.
func TestInit(t *testing.T) {
	err, _ := flavor.GetTemplate("postgres")

	if err != nil {
		t.Error("postgres flavor should be registered at init phase.")
	}
}

// Create set of template argumments which do not include
// required one called "endpoint"
func TestMissingParam(t *testing.T) {

	params := map[string]string{
		"blah": "bleh",
	}
	var testFlavor PostgresFlavor

	err, _ := testFlavor.CheckParams(params)

	if err == nil {
		t.Error("Not specifying mandatory template args does not trigger error")
	}
}

// Verify that passing all required params does not trigger any arror.
func TestAllParams(t *testing.T) {
	params := map[string]string{
		"Endpoint": "127.0.0.1",
	}
	var testFlavor PostgresFlavor

	err, data := testFlavor.CheckParams(params)

	if err != nil {
		t.Errorf("All required params were passed but check failed: %s", err)
	}

	// Type assertion from generic interface{} to PostgresFlavor.
	postgresData := data.(PostgresFlavor)
	if postgresData.Endpoint != "127.0.0.1" {
		t.Errorf("Parsing template params does not create proper structure")
	}
}

// Verify that as long as required params are included template processing
// is successful
func TestExtraParams(t *testing.T) {
	params := map[string]string{
		"Endpoint": "127.0.0.1",
		"blah":     "blah",
	}
	var testFlavor PostgresFlavor

	err, data := testFlavor.CheckParams(params)

	if err != nil {
		t.Errorf("All required params were passed but check failed: %s", err)
	}

	// Type assertion from generic interface{} to PostgresFlavor.
	postgresData := data.(PostgresFlavor)
	if postgresData.Endpoint != "127.0.0.1" {
		t.Errorf("Parsing template params does not create proper structure")
	}
}

// Make sure the GetTemplate returns the correct config.
func TestGetTemplate(t *testing.T) {
	var testFlavor PostgresFlavor
	if configTemplate != testFlavor.GetTemplate() {
		t.Errorf("Wrong config template returned.")
	}
}
