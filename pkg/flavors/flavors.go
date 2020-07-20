package flavor

import (
	"fmt"
)

// Interface to individual flavors.
type FlavorConfigTemplate interface {
  CreateConfig(params map[string]string) (error, string)
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
