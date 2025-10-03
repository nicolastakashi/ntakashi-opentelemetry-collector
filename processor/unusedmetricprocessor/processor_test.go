package unusedmetricprocessor

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/metadata"
	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/server"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
)

type fakeClient struct {
	// map[job][metricName] => unused
	decisions map[string]map[string]bool
	// optional error injection
	errFor map[string]map[string]error
}

func (f *fakeClient) GetMetricUsage(ctx context.Context, job string, name string) (server.MetricUsage, error) {
	if jobMap, ok := f.errFor[job]; ok {
		if err, ok2 := jobMap[name]; ok2 && err != nil {
			return server.MetricUsage{}, err
		}
	}
	unused := false
	if jobMap, ok := f.decisions[job]; ok {
		if val, ok2 := jobMap[name]; ok2 {
			unused = val
		}
	}
	return server.MetricUsage{Unused: unused, Name: name}, nil
}

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

			// Build cfg with defaults
			cfg := NewFactory().CreateDefaultConfig().(*Config)
			cfg.Server.Address = "http://localhost:0"
			// Ensure timeout non-nil for any code paths that might use it
			cfg.Server.Timeout = func() *time.Duration { d := 100 * time.Millisecond; return &d }()
			require.NoError(t, cfg.Validate())

			// Prepare fake client behavior per test case
			f := &fakeClient{decisions: map[string]map[string]bool{}}
			switch tc.name {
			case "drop_unused_metric_if_present":
				f.decisions["myJob"] = map[string]bool{
					"unused_metric": true, // should be dropped
				}
			case "keep_metric_if_used":
				f.decisions["myJob"] = map[string]bool{
					"used_metric": false, // should be kept
				}
			}

			processor, err := newUnusedMetricProcessor(
				ctx,
				processortest.NewNopSettings(metadata.Type),
				cfg,
				next,
				f,
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
