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
package flavors

import (
	"testing"
)

// Flavor for mocking and testing.
type TestFlavor struct {
	Test string
}

var flavor TestFlavor

func (TestFlavor) CreateConfig(params map[string]string) (string, error) {
	return "CreateTestConfig", nil
}

func (TestFlavor) CheckParams(params map[string]string) (interface{}, error) {
	// Just set the
	flavor.Test = "UnitTest"
	return flavor, nil
}

func (TestFlavor) GetTemplate() string {
	return "This is {{ .Test }} template"
}

// Test adding and retrieving config template
func TestAdd(t *testing.T) {
	AddTemplate("test", flavor)

	out, err := GetTemplate("test")

	if err != nil {
		t.Error("Just added template cannot be located")
	}

	if flavor != out {
		t.Error("Added and retrieved templates are different")
	}
}

// Test retrieving non-existing template
func TestGetNonExisting(t *testing.T) {
	AddTemplate("test", flavor)

	_, err := GetTemplate("test1")

	if err == nil {
		t.Error("Error should be returned for non-existing template")
	}
}

// Test creating config.
// Test verifies that after adding TestFlavor to the list
// of known flavors, it can create a proper config.
func TestCreateConfig(t *testing.T) {
	AddTemplate("test", flavor)

	params := map[string]string{"Test": "UnitTest"}
	config, err := CreateConfig("test", params)

	if err != nil {
		t.Error("Creating config failed with proper parameters")
	}

	if config != "This is UnitTest template" {
		t.Errorf("Created config %s not as expected", config)
	}
}
