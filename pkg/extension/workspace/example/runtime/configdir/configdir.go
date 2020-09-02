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

package configdir

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	envoybootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"

	"github.com/tetratelabs/getenvoy/pkg/extension/manager"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/envoy/template"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/envoy/util"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/multierror"

	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

// NewConfigDir creates a config directory for a single example run.
func NewConfigDir(ctx *runtime.RunContext) (_ ConfigDir, errs error) {
	// create a temporary directory
	tempDir, err := ioutil.TempDir("", "getenvoy_extension_run")
	if err != nil {
		return nil, err
	}
	tempDir, err = filepath.EvalSymlinks(tempDir)
	if err != nil {
		return nil, err
	}
	dir := &configDir{tempDir: tempDir, ctx: ctx}
	defer func() {
		if errs != nil {
			if e := dir.Close(); e != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}()
	if err := dir.init(); err != nil {
		return nil, err
	}
	if err := dir.process(); err != nil {
		return nil, err
	}
	return dir, nil
}

// ConfigDir represents a config directory of a single example run.
type ConfigDir interface {
	GetDir() string
	GetBootstrapFile() string
	GetBootstrap() *envoybootstrap.Bootstrap
	io.Closer
}

// configDir represents a config directory of a single example run.
type configDir struct {
	tempDir string
	ctx     *runtime.RunContext

	bootstrapFileName string
	bootstrap         *envoybootstrap.Bootstrap
}

func (d *configDir) GetDir() string {
	return d.tempDir
}

func (d *configDir) GetBootstrapFile() string {
	return filepath.Join(d.tempDir, d.bootstrapFileName)
}

func (d *configDir) GetBootstrap() *envoybootstrap.Bootstrap {
	return d.bootstrap
}

func (d *configDir) Close() error {
	return os.RemoveAll(d.tempDir)
}

func (d *configDir) init() error {
	files := d.ctx.Opts.Example.GetFiles()
	// copy all example files
	for _, fileName := range files.GetNames() {
		if err := d.writeFile(fileName, files.Get(fileName).Content); err != nil {
			return err
		}
	}
	return nil
}

// process resolves placehoders in the Envoy bootstrap config, including
//  1) placeholders in the bootstrap file (envoy.tmpl.yaml or envoy.tmpl.json)
//  2) (optional) placeholders in a LDS file (value of `bootstrap.dynamic_resources.lds_config.path`)
//  3) (optional) placeholders in a CDS file (value of `bootstrap.dynamic_resources.cds_config.path`)
func (d *configDir) process() error {
	expandContext, err := d.newExpandContext()
	if err != nil {
		return err
	}

	// resolve placehoders in the bootstrap file
	bootstrapFileName, bootstrapFile := d.ctx.Opts.Example.GetEnvoyConfig()
	bootstrapContent, err := d.processEnvoyTemplate(bootstrapFile, expandContext)
	if err != nil {
		return err
	}
	d.bootstrapFileName = bootstrapFileName
	// overwrite the original file from the Example
	err = d.writeFile(d.bootstrapFileName, bootstrapContent)
	if err != nil {
		return err
	}

	// unmarshal bootstrap config to inspect the actual configuration
	var bootstrap envoybootstrap.Bootstrap
	err = util.Load(&model.File{Source: bootstrapFile.Source, Content: bootstrapContent}, &bootstrap)
	if err != nil {
		// if configuration is indeed invalid, let the error message come from Envoy
		return nil
	}
	d.bootstrap = &bootstrap

	// resolve placehoders in the LDS file
	if fileName := d.bootstrap.GetDynamicResources().GetLdsConfig().GetPath(); fileName != "" {
		if err := d.processEnvoyXdsFile(fileName, expandContext); err != nil {
			return err
		}
	}

	// resolve placehoders in the CDS file
	if fileName := d.bootstrap.GetDynamicResources().GetCdsConfig().GetPath(); fileName != "" {
		if err := d.processEnvoyXdsFile(fileName, expandContext); err != nil {
			return err
		}
	}
	return nil
}

func (d *configDir) newExpandContext() (*template.ExpandContext, error) {
	wasmFile, err := filepath.Abs(d.ctx.Opts.Extension.WasmFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve absolute path of a *.wasm file %q", d.ctx.Opts.Extension.WasmFile)
	}
	configuration := string(d.ctx.Opts.GetExtensionConfig().Content)

	return &template.ExpandContext{
		DefaultExtension:       manager.NewLocalExtension(d.ctx.Opts.Workspace.GetExtensionDescriptor(), wasmFile),
		DefaultExtensionConfig: configuration,
	}, nil
}

func (d *configDir) processEnvoyXdsFile(fileName string, expandContext *template.ExpandContext) error {
	files := d.ctx.Opts.Example.GetFiles()
	if !files.Has(fileName) {
		// if configuration is indeed invalid, e.g. non-existing path, let the error message come from Envoy
		return nil
	}
	file := files.Get(fileName)
	content, err := d.processEnvoyTemplate(file, expandContext)
	if err != nil {
		return err
	}
	// overwrite the original file from the Example
	return d.writeFile(fileName, content)
}

// processEnvoyTemplate resolves placehoders in an Envoy config file.
func (d *configDir) processEnvoyTemplate(configFile *model.File, expandContext *template.ExpandContext) ([]byte, error) {
	config, err := template.Expand(configFile.Content, expandContext)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to process Envoy config template coming from %q", configFile.Source)
	}
	return config, nil
}

func (d *configDir) writeFile(fileName string, data []byte) error {
	outputFile := filepath.Join(d.tempDir, fileName)
	if err := osutil.EnsureDirExists(filepath.Dir(outputFile)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(outputFile, data, 0644); err != nil {
		return errors.Wrapf(err, "failed to write config file to %q", outputFile)
	}
	return nil
}
