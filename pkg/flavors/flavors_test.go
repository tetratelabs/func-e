package flavor

import (
	"testing"
)

type TestFlavor struct {
}

func (TestFlavor) CreateConfig(params map[string]string) (error, string) {
  return nil, "CreateTestConfig"
}
// Test adding and retrieving config template
func TestAdd(t *testing.T) {
  var flavor TestFlavor

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
  var flavor TestFlavor

  AddTemplate("test", flavor)

  err, _ := GetTemplate("test1")

  if err == nil {
	  t.Error("Error should be returned for non-existing template")
  }

}
