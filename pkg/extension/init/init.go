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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
)

// ProgressHandler is a sink for progress events.
type ProgressHandler interface {
	OnStart()
	OnFile(file string)
	OnComplete()
}

// ProgressFuncs dispatches progress events to individual handler functions.
type ProgressFuncs struct {
	OnStartFunc    func()
	OnFileFunc     func(file string)
	OnCompleteFunc func()
}

// OnStart is called when scaffolding is about to get started.
func (f ProgressFuncs) OnStart() {
	if f.OnStartFunc != nil {
		f.OnStartFunc()
	}
}

// OnFile is called for every generated file.
func (f ProgressFuncs) OnFile(file string) {
	if f.OnFileFunc != nil {
		f.OnFileFunc(file)
	}

}

// OnComplete is called when scaffolding has finished successfully.
func (f ProgressFuncs) OnComplete() {
	if f.OnCompleteFunc != nil {
		f.OnCompleteFunc()
	}
}

// ScaffoldOpts represents configuration options supported by Scaffold().
type ScaffoldOpts struct {
	Language     string
	Category     string
	TemplateName string

	OutputDir string

	ProgressHandler
}

// Scaffold generates the initial set of files to kick off development of a new extension.
func Scaffold(opts *ScaffoldOpts) (err error) {
	templateDir, fs, err := getTemplateSource().GetTemplateDir(opts.Language, opts.Category, opts.TemplateName)
	if err != nil {
		return fmt.Errorf("no such template: %v", err)
	}
	if opts.ProgressHandler == nil {
		opts.ProgressHandler = ProgressFuncs{}
	}
	opts.ProgressHandler.OnStart()
	walker := &scaffolder{opts: opts, sourceFS: fs}
	defer func() {
		if err == nil {
			opts.OnComplete()
		}
	}()
	return walker.walk(templateDir, "")
}

// scaffolder generates the initial set of files to kick off development of a new extension.
type scaffolder struct {
	opts      *ScaffoldOpts
	sourceFS  http.FileSystem
	transform func([]byte) ([]byte, error)
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
	if err := os.MkdirAll(filepath.Dir(outputFileName), os.ModeDir|0755); err != nil {
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
		data, err := s.transform(content)
		if err != nil {
			return err
		}
		content = data
	}
	if err := ioutil.WriteFile(outputFileName, content, sourceFileInfo.Mode()); err != nil {
		return err
	}
	s.opts.ProgressHandler.OnFile(relOutputFileName)
	return nil
}
