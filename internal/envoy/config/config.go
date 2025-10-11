// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Admin           *admin           `yaml:"admin"`
	StaticResources *staticResources `yaml:"static_resources"`
}

type admin struct {
	Address Address `yaml:"address"`
}

type staticResources struct {
	Listeners []listener `yaml:"listeners"`
}

type listener struct {
	Name            string        `yaml:"name"`
	Address         Address       `yaml:"address"`
	FilterChains    []filterChain `yaml:"filter_chains"`
	ListenerFilters []filter      `yaml:"listener_filters"`
}

type filterChain struct {
	Filters []filter `yaml:"filters"`
}

type filter struct {
	Name        string    `yaml:"name"`
	TypedConfig yaml.Node `yaml:"typed_config"`
}

type Address struct {
	SocketAddress socketAddress `yaml:"socket_address"`
}

type socketAddress struct {
	Address   string `yaml:"address"`
	PortValue int    `yaml:"port_value"`
	Protocol  string `yaml:"protocol"`
}

type filterInfo struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Config string `yaml:"config"`
}

type Listener struct {
	Name     string
	Address  string // host:port format
	Protocol string
	Filters  []filterInfo
}

type Config struct {
	Admin           string // host:port format, empty if no admin
	StaticListeners []Listener
}

// ParseListeners parses the admin address (if any) and all static listeners from config sources.
//
// This mimics Envoy's config merging behavior from source/server/server.cc:
//   - configPath is loaded first (if non-empty)
//   - configYaml is merged on top via protobuf MergeFrom (if non-empty)
//   - configYaml always wins for conflicting fields, regardless of which was specified first on CLI
//
// For listeners with the same name, the later config wins (protobuf MergeFrom behavior).
func ParseListeners(configPath, configYaml string) (*Config, error) {
	listenerMap := make(map[string]Listener)
	var adminAddr string

	// Load config-path first
	if configPath != "" {
		yamlBytes, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
		}
		admin, listeners, err := parseListenersFromYAML(string(yamlBytes))
		if err != nil {
			return nil, err
		}
		if admin != "" {
			adminAddr = admin
		}
		for _, l := range listeners {
			listenerMap[l.Name] = l
		}
	}

	// Merge config-yaml on top (always wins)
	if configYaml != "" {
		admin, listeners, err := parseListenersFromYAML(configYaml)
		if err != nil {
			return nil, err
		}
		if admin != "" {
			adminAddr = admin
		}
		for _, l := range listeners {
			listenerMap[l.Name] = l
		}
	}

	allListeners := make([]Listener, 0, len(listenerMap))
	for _, l := range listenerMap {
		allListeners = append(allListeners, l)
	}

	return &Config{
		Admin:           adminAddr,
		StaticListeners: allListeners,
	}, nil
}

// FindAdminAddress parses the admin address from config sources.
func FindAdminAddress(configPath, configYaml string) (string, error) {
	result, err := ParseListeners(configPath, configYaml)
	if err != nil {
		return "", err
	}
	return result.Admin, nil
}

// FindAdminAddressFromArgs extracts config sources from args and returns the admin address.
func FindAdminAddressFromArgs(args []string) (string, error) {
	var configPath, configYaml string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "--config-path":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--config-yaml":
			if i+1 < len(args) {
				configYaml = args[i+1]
				i++
			}
		}
	}
	return FindAdminAddress(configPath, configYaml)
}

func parseListenersFromYAML(yamlString string) (admin string, listeners []Listener, err error) {
	config := config{StaticResources: &staticResources{}} // prevent nils
	err = yaml.Unmarshal([]byte(yamlString), &config)
	if err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Parse admin if present
	if config.Admin != nil {
		sa := config.Admin.Address.SocketAddress
		if sa.Address != "" && sa.PortValue >= 0 {
			admin = formatAddr(sa)
		}
	}

	// Parse static listeners if present
	for _, listener := range config.StaticResources.Listeners {
		var filters []filterInfo

		for _, chain := range listener.FilterChains {
			for _, filter := range chain.Filters {
				filters = append(filters, extractFilterInfo(filter))
			}
		}

		for _, filter := range listener.ListenerFilters {
			filters = append(filters, extractFilterInfo(filter))
		}

		staticListener := Listener{
			Name:     listener.Name,
			Address:  formatAddr(listener.Address.SocketAddress),
			Protocol: listener.Address.SocketAddress.Protocol,
			Filters:  filters,
		}
		listeners = append(listeners, staticListener)
	}

	return admin, listeners, nil
}

func extractFilterInfo(f filter) filterInfo {
	fi := filterInfo{Name: f.Name}
	if f.TypedConfig.Kind == yaml.MappingNode {
		var typedMap map[string]interface{}
		if err := f.TypedConfig.Decode(&typedMap); err == nil {
			if t, ok := typedMap["@type"].(string); ok {
				fi.Type = t
			}
		}
		if raw, err := yaml.Marshal(&f.TypedConfig); err == nil {
			fi.Config = string(raw)
		}
	}
	return fi
}

func formatAddr(sa socketAddress) string {
	return fmt.Sprintf("%s:%d", sa.Address, sa.PortValue)
}
