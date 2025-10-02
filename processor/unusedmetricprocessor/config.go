// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package unusedmetricprocessor // import "github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor"

import (
	"errors"
	"time"
)

var defaultTimeout = 10 * time.Second

type Config struct {
	// prevents unkeyed literal initialization
	_ struct{}

	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	// address of the server to connect to
	Address string `mapstructure:"address"`

	// timeout for the client to connect to the server
	// default is 10 seconds
	Timeout *time.Duration `mapstructure:"timeout"`

	// tls configuration
	TlsConfig TLSConfig `mapstructure:"tls_config"`
}

type TLSConfig struct {
	// if true, the server's certificate will not be verified
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return errors.New("server address is required")
	}
	if c.Server.Timeout == nil {
		c.Server.Timeout = &defaultTimeout
	}
	return nil
}
