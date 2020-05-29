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
	"github.com/pkg/errors"

	"github.com/docker/distribution/reference"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	"github.com/tetratelabs/multierror"
)

// NewToolchainConfig returns a new built-in toolchain config.
func NewToolchainConfig() *ToolchainConfig {
	return &ToolchainConfig{
		Meta: config.Meta{
			Kind: Kind,
		},
	}
}

// DefaultTo sets default values according to a given config.
func (c *ToolchainConfig) DefaultTo(defaultConfig *ToolchainConfig) {
	if c.Container == nil && defaultConfig.Container != nil {
		c.Container = new(ContainerConfig)
	}
	c.Container.DefaultTo(defaultConfig.Container)
	if c.Build == nil && defaultConfig.Build != nil {
		c.Build = new(BuildConfig)
	}
	c.Build.DefaultTo(defaultConfig.Build)
	if c.Test == nil && defaultConfig.Test != nil {
		c.Test = new(TestConfig)
	}
	c.Test.DefaultTo(defaultConfig.Test)
	if c.Clean == nil && defaultConfig.Clean != nil {
		c.Clean = new(CleanConfig)
	}
	c.Clean.DefaultTo(defaultConfig.Clean)
}

// DefaultTo sets default values according to a given config.
func (c *BuildConfig) DefaultTo(defaultConfig *BuildConfig) {
	if defaultConfig == nil {
		return
	}
	if c.Container == nil && defaultConfig.Container != nil {
		c.Container = new(ContainerConfig)
	}
	c.Container.DefaultTo(defaultConfig.Container)
}

// DefaultTo sets default values according to a given config.
func (c *TestConfig) DefaultTo(defaultConfig *TestConfig) {
	if defaultConfig == nil {
		return
	}
	if c.Container == nil && defaultConfig.Container != nil {
		c.Container = new(ContainerConfig)
	}
	c.Container.DefaultTo(defaultConfig.Container)
}

// DefaultTo sets default values according to a given config.
func (c *CleanConfig) DefaultTo(defaultConfig *CleanConfig) {
	if defaultConfig == nil {
		return
	}
	if c.Container == nil && defaultConfig.Container != nil {
		c.Container = new(ContainerConfig)
	}
	c.Container.DefaultTo(defaultConfig.Container)
}

// DefaultTo sets default values according to a given config.
func (c *ContainerConfig) DefaultTo(defaultConfig *ContainerConfig) {
	if defaultConfig == nil {
		return
	}
	if c.Image == "" {
		c.Image = defaultConfig.Image
	}
	if c.Options == nil {
		c.Options = defaultConfig.Options
	}
}

// Validate returns an error if ToolchainConfig is not valid.
func (c *ToolchainConfig) Validate() (errs error) {
	if c.Container == nil {
		errs = multierror.Append(errs, errors.New("configuration of the default build container cannot be empty"))
	}
	if c.Container != nil {
		if err := c.Container.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "configuration of the default build container is not valid"))
		}
	}
	if c.Build != nil {
		if err := c.Build.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "configuration of 'build' tool is not valid"))
		}
	}
	if c.Test != nil {
		if err := c.Test.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "configuration of 'test' tool is not valid"))
		}
	}
	if c.Clean != nil {
		if err := c.Clean.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "configuration of 'clean' tool is not valid"))
		}
	}
	return
}

// Validate returns an error if BuildConfig is not valid.
func (c *BuildConfig) Validate() (errs error) {
	if c.Container != nil {
		if err := c.Container.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "container configuration is not valid"))
		}
	}
	return
}

// Validate returns an error if TestConfig is not valid.
func (c *TestConfig) Validate() (errs error) {
	if c.Container != nil {
		if err := c.Container.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "container configuration is not valid"))
		}
	}
	return
}

// Validate returns an error if CleanConfig is not valid.
func (c *CleanConfig) Validate() (errs error) {
	if c.Container != nil {
		if err := c.Container.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "container configuration is not valid"))
		}
	}
	return
}

// Validate returns an error if ContainerConfig is not valid.
func (c *ContainerConfig) Validate() (errs error) {
	if c.Image == "" {
		errs = multierror.Append(errs, errors.New("image name cannot be empty"))
	}
	if c.Image != "" {
		if _, err := reference.Parse(c.Image); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "%q is not a valid image name", c.Image))
		}
	}
	return
}

// GetBuildContainer returns effective configuration of a container used by 'build' tool.
func (c *ToolchainConfig) GetBuildContainer() *ContainerConfig {
	if c.Build != nil && c.Build.Container != nil {
		return c.Build.Container
	}
	return c.Container
}

// GetTestContainer returns effective configuration of a container used by 'test' tool.
func (c *ToolchainConfig) GetTestContainer() *ContainerConfig {
	if c.Test != nil && c.Test.Container != nil {
		return c.Test.Container
	}
	return c.Container
}

// GetCleanContainer returns effective configuration of a container used by 'clean' tool.
func (c *ToolchainConfig) GetCleanContainer() *ContainerConfig {
	if c.Clean != nil && c.Clean.Container != nil {
		return c.Clean.Container
	}
	return c.Container
}
