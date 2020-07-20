package postgres

import (
	"testing"
        "github.com/tetratelabs/getenvoy/pkg/flavors"
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

  params := map[string]string {
	  "blah": "bleh",
  }
  var testFlavor PostgresFlavor

  err, _ := testFlavor.CreateConfig(params)

  if err == nil {
	  t.Error("Not specifying mandatory template args does not trigger error")
  }
}

// Verify that passing all required params does not trigger any arror.
func TestCreateConfig(t *testing.T) {
  params := map[string]string {
		"Endpoint": "127.0.0.1",
	}
  var testFlavor PostgresFlavor

  err, _ := testFlavor.CreateConfig(params)

  if err != nil {
	  t.Errorf("Creating config failed: %s", err)
  }
}

// Verify that as long as required params are included template processing 
// is successful
func TestExtraParams(t * testing.T) {
  params := map[string]string {
		"Endpoint": "127.0.0.1",
		"blah": "blah",
	}
  var testFlavor PostgresFlavor

  err, _ := testFlavor.CreateConfig(params)

  if err != nil {
	  t.Errorf("Creating config failed: %s", err)
  }
}
