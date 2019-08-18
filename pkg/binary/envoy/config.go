package envoy

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"istio.io/pkg/log"
)

type Mode string

const (
	Sidecar Mode = "sidecar"
	Router  Mode = "router"
)

var ValidModes = []string{string(Sidecar), string(Router)}

func ParseMode(s string) (Mode, error) {
	switch {
	case Mode(s) == Sidecar:
		log.Warnf("sidecar mode not yet supported, using router instead")
		return Router, nil
	case Mode(s) == Router:
		return Router, nil
	case s == "":
		return "", nil
	default:
		return "", fmt.Errorf("unable to parse mode %v, must be one of %v", s, ValidModes)
	}
}

func NewConfig(options ...func(*Config)) *Config {
	cfg := &Config{
		AdminPort:      15000,
		StatNameLength: 189,
		DrainDuration:  types.DurationProto(30 * time.Second),
		ConnectTimeout: types.DurationProto(5 * time.Second),
	}
	for _, o := range options {
		o(cfg)
	}
	return cfg
}

type Config struct {
	XDSAddress     string
	Mode           Mode
	IPAddresses    []string
	ALSAddresss    string
	DrainDuration  *types.Duration
	ConnectTimeout *types.Duration
	AdminPort      int32
	StatNameLength int32
}
