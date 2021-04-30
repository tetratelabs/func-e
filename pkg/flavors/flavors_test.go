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
package flavors_test

import (
	"testing"

	"github.com/tetratelabs/getenvoy/pkg/flavors"
)

// Flavor for mocking and testing.
type TestFlavor struct {
	Test string
}

var flavor TestFlavor

const testConfig string = "This is UnitTest config"

func (f *TestFlavor) GenerateConfig(_ map[string]string) (string, error) {
	// Just return predefined value.
	// Normally a flavor would use template but
	// here using templates is not part of the test.
	return testConfig, nil
}

// Test adding and retrieving config template
func TestAdd(t *testing.T) {
	flavors.AddFlavor("test", &flavor)

	_, err := flavors.GetFlavor("test")

	if err != nil {
		t.Error("Just added template cannot be located")
	}
}

// Test retrieving non-existing template
func TestGetNonExisting(t *testing.T) {
	flavors.AddFlavor("test", &flavor)

	_, err := flavors.GetFlavor("test1")

	if err == nil {
		t.Error("Error should be returned for non-existing template")
	}
}

// Test creating config.
// Test verifies that after adding TestFlavor to the list
// of known flavors, it can create a proper config.
func TestCreateConfig(t *testing.T) {
	flavors.AddFlavor("test", &flavor)

	params := map[string]string{"Test": "UnitTest"}
	config, err := flavors.CreateConfig("test", params)

	if err != nil {
		t.Error("Creating config failed with proper parameters")
	}

	if config != testConfig {
		t.Errorf("Created config %s not as expected", config)
	}
}
