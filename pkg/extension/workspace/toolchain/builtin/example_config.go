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

package builtin

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
	extensionconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var (
	// exampleConfigTemplate represents an example configuration
	// that will be advertised to developers.
	exampleConfigTemplate = `#
# Configuration for the built-in toolchain.
#
kind: BuiltinToolchain

#
# Default configuration for a build container.
#

## container:
##   # Builder image.
##   image: {{ .BuildImage }}
##   # Docker cli options.
##   options: []

#
# Configuration for the 'build' command.
#
# If omitted, the default configuration for a build container will be used instead.
#

## build:
##   container:
##     # Builder image.
##     image: {{ .BuildImage }}
##     # Docker cli options.
##     options: []

#
# Configuration for the 'test' command.
#
# If omitted, the default configuration for a build container will be used instead.
#

## test:
##   container:
##     # Builder image.
##     image: {{ .BuildImage }}
##     # Docker cli options.
##     options: []
`
)

type exampleConfigTemplateArgs struct {
	BuildImage string
}

// ExampleConfig returns an example toolchain config for a given extension.
func ExampleConfig(extension *extensionconfig.Descriptor) []byte {
	args := exampleConfigTemplateArgs{
		BuildImage: defaultBuildImageFor(extension.Language),
	}
	return renderExampleConfigTemplate(args)
}

func renderExampleConfigTemplate(args exampleConfigTemplateArgs) []byte {
	tmpl, err := template.New("").Parse(exampleConfigTemplate)
	if err != nil {
		// must be caught by unit tests
		panic(err)
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, args)
	if err != nil {
		// must be caught by unit tests
		panic(errors.Wrap(err, "failed to render example configuration template"))
	}
	return out.Bytes()
}
