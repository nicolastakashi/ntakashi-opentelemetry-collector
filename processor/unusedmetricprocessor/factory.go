// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package unusedmetricprocessor // import "github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"

	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/metadata"
	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/server"
)

// NewFactory returns a new factory for the Span processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(
	ctx context.Context,
	params processor.Settings,
	baseCfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	cfg := baseCfg.(*Config)

	client := server.NewClient(&server.Config{
		Address: cfg.Server.Address,
		Timeout: cfg.Server.Timeout,
		TlsConfig: server.TLSConfig{
			InsecureSkipVerify: cfg.Server.TlsConfig.InsecureSkipVerify,
		},
	})

	unusedMetricProcessor, err := newUnusedMetricProcessor(ctx,
		params,
		cfg,
		nextConsumer,
		client,
	)
	if err != nil {
		return nil, err
	}

	return unusedMetricProcessor, nil
}
