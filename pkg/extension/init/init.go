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

package init

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"github.com/tetratelabs/multierror"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"

	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
	scaffoldutil "github.com/tetratelabs/getenvoy/pkg/util/scaffold"
)

// ScaffoldOpts represents configuration options supported by Scaffold().
type ScaffoldOpts struct {
	Extension    *extension.Descriptor
	TemplateName string

	OutputDir string

	ProgressSink scaffoldutil.ProgressSink
}

// Scaffold generates the initial set of files to kick off development of a new extension.
func Scaffold(opts *ScaffoldOpts) (err error) {
	templateDir, fs, err := getTemplateSource().GetTemplateDir(opts.Extension.Language, opts.Extension.Category, opts.TemplateName)
	if err != nil {
		return fmt.Errorf("no such template: %v", err)
	}
	if opts.ProgressSink == nil {
		opts.ProgressSink = scaffoldutil.NoOpProgressSink()
	}
	opts.ProgressSink.OnStart()
	defer func() {
		if err == nil {
			opts.ProgressSink.OnComplete()
		}
	}()
	if err := generateWorkspace(opts); err != nil {
		return err
	}
	walker := &scaffolder{opts: opts, sourceFS: fs, transform: interpolate(opts.Extension)}
	return walker.walk(templateDir, "")
}

// scaffolder generates the initial set of files to kick off development of a new extension.
type scaffolder struct {
	opts      *ScaffoldOpts
	sourceFS  http.FileSystem
	transform func(string, []byte) ([]byte, error)
}

func (s *scaffolder) walk(sourceDirName, destinationDirName string) (errs error) {
	sourceDir, err := s.sourceFS.Open(sourceDirName)
	if err != nil {
		return err
	}
	defer func() {
		if e := sourceDir.Close(); e != nil {
			errs = multierror.Append(errs, e)
		}
	}()
	sourceFiles, err := sourceDir.Readdir(-1)
	if err != nil && err != io.EOF {
		return err
	}
	for _, sourceFile := range sourceFiles {
		if sourceFile.IsDir() {
			if err := s.walk(path.Join(sourceDirName, sourceFile.Name()), filepath.Join(destinationDirName, sourceFile.Name())); err != nil {
				return err
			}
		} else {
			if err := s.visit(sourceDirName, destinationDirName, sourceFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *scaffolder) visit(sourceDirName, destinationDirName string, sourceFileInfo os.FileInfo) (errs error) {
	relOutputFileName := filepath.Join(destinationDirName, sourceFileInfo.Name())
	outputFileName := filepath.Join(s.opts.OutputDir, relOutputFileName)
	if err := osutil.EnsureDirExists(filepath.Dir(outputFileName)); err != nil {
		return err
	}
	sourceFile, err := s.sourceFS.Open(path.Join(sourceDirName, sourceFileInfo.Name()))
	if err != nil {
		return err
	}
	defer func() {
		if e := sourceFile.Close(); e != nil {
			errs = multierror.Append(errs, e)
		}
	}()
	content, err := ioutil.ReadAll(sourceFile)
	if err != nil {
		return err
	}
	if s.transform != nil {
		data, err := s.transform(relOutputFileName, content)
		if err != nil {
			return err
		}
		content = data
	}
	if err := ioutil.WriteFile(outputFileName, content, sourceFileInfo.Mode()); err != nil {
		return err
	}
	s.opts.ProgressSink.OnFile(relOutputFileName)
	return nil
}

// interpolateData represents the data object a source code template will be applied to.
type interpolateData struct {
	Extension *extension.Descriptor
}

// interpolate resolves placeholders in a given source file.
func interpolate(descriptor *extension.Descriptor) func(string, []byte) ([]byte, error) {
	data := &interpolateData{Extension: descriptor}
	return func(file string, content []byte) ([]byte, error) {
		tmpl, err := template.New(file).Parse(string(content))
		if err != nil {
			// must be caught by unit tests
			panic(err)
		}
		var out bytes.Buffer
		err = tmpl.Execute(&out, data)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render %q", file)
		}
		return out.Bytes(), nil
	}
}
