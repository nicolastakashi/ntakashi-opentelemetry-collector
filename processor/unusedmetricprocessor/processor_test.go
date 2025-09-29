package unusedmetricprocessor

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestProcessor(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testCases := []struct {
		name string
	}{
		{
			name: "drop_unused_metric_if_present",
		},
		{
			name: "keep_metric_if_used",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			next := &consumertest.MetricsSink{}
			factory := NewFactory()
			processor, err := factory.CreateMetrics(
				ctx,
				processortest.NewNopSettings(metadata.Type),
				factory.CreateDefaultConfig(),
				next,
			)
			require.NoError(t, err)
			dir := filepath.Join("testdata", tc.name)

			md, err := golden.ReadMetrics(filepath.Join(dir, "input.yaml"))
			require.NoError(t, err)

			// Test that ConsumeMetrics works
			err = processor.ConsumeMetrics(ctx, md)
			require.NoError(t, err)

			allMetrics := next.AllMetrics()

			expectedNextData, err := golden.ReadMetrics(filepath.Join(dir, "output.yaml"))
			require.NoError(t, err)
			require.NoError(t, pmetrictest.CompareMetrics(expectedNextData, allMetrics[0]))
		})
	}
}
