package flavor

import (
	"testing"
)

type TestFlavor struct {
  Test	string
}

var flavor TestFlavor


func (TestFlavor) CreateConfig(params map[string]string) (error, string) {
  return nil, "CreateTestConfig"
}

func (TestFlavor)  CheckParams(params map[string]string) (error, interface{}) {
	// Just set the 
	flavor.Test = "UnitTest"
	return nil, flavor
}

func  (TestFlavor) GetTemplate() string {
	return "This is {{ .Test }} template"
}
// Test adding and retrieving config template
func TestAdd(t *testing.T) {
  AddTemplate("test", flavor)

  err, out := GetTemplate("test")

  if err != nil {
	  t.Error("Just added template cannot be located")
  }

  if (flavor != out) {
	  t.Error("Added and retrieved templates are different")
  }
}

// Test retrieving non-existing template
func TestGetNonExisting(t *testing.T) {
  AddTemplate("test", flavor)

  err, _ := GetTemplate("test1")

  if err == nil {
	  t.Error("Error should be returned for non-existing template")
  }
}

// Test creating config.
// Test verifies that after adding TestFlavor to the list
// of known flavors, it can create a proper config.
func TestCreateConfig(t *testing.T) {
  AddTemplate("test", flavor)

  params := map[string]string {"Test": "UnitTest"}
  err, config := CreateConfig("test", params)

  if err != nil {
    t.Error("Creating config failed with proper parameters")
  }

  if config != "This is UnitTest template" {
    t.Errorf("Created config %s not as expected", config)
  }
}
