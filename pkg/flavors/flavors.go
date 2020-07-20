package flavor

import (
	"bytes"
	"fmt"
	"text/template"
)

// Interface to individual flavors.
type FlavorConfigTemplate interface {
  CreateConfig(params map[string]string) (error, string)
  CheckParams(params map[string]string) (error, interface{})
  GetTemplate() string
}


// Main repo for templates.

type TemplateStore struct {
  // This is a map indexed by flavor pointing to the individual 
  // implementaions of each flavor.
  templates map[string]FlavorConfigTemplate
}

var store TemplateStore = TemplateStore{templates: make(map[string]FlavorConfigTemplate)}

//func New(path string) *TemplateStore {
//	return &TemplateStore{path: path};
//}

func (store *TemplateStore) GetTemplate(flavor string) (error, FlavorConfigTemplate) {
  template, ok := store.templates[flavor]
  if !ok {
	  return fmt.Errorf("Cannot find template for flavor %s", flavor), nil
  }

  return nil, template
}

func AddTemplate(flavor string,  configTemplate FlavorConfigTemplate) {
  store.templates[flavor] = configTemplate
}

func GetTemplate(flavor string) (error, FlavorConfigTemplate) {
  template, ok := store.templates[flavor]
  if !ok {
	  return fmt.Errorf("Cannot find template for flavor %s", flavor), nil
  }

  return nil, template
}

// Function checks flavor specific paramaters, get flavor's template and
// create a config.
func CreateConfig(flavor string, params map[string]string) (error, string) {
  err, flavorData := GetTemplate(flavor)

  if err != nil {
	  return err, ""
  }

  err, data := flavorData.CheckParams(params)
  if err != nil  {
	return err, ""
  }

  // NOw run the template substitution
  tmpl := template.New(flavor)
  tmpl, err = tmpl.Parse(flavorData.GetTemplate())
  if err != nil {
    // Template is not supplied by a user, but is compiled-in, so this error should
    // happen only during development time.
    return fmt.Errorf("Supplied template for flavor %s is incorrect.", flavor), ""
  }
  var buf bytes.Buffer
  tmpl.Execute(&buf, data)
  return nil, buf.String()
}

