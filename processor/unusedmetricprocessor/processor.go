package unusedmetricprocessor // import "github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor"

import (
	"context"
	"time"

	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/metadata"
	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/server"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type unusedMetricProcessor struct {
	config *Config
	client server.Client
	component.StartFunc
	component.ShutdownFunc
	logger    *zap.Logger
	telemetry *metadata.TelemetryBuilder
}

func newUnusedMetricProcessor(
	ctx context.Context,
	settings processor.Settings,
	cfg *Config,
	nextConsumer consumer.Metrics,
	client server.Client,
) (processor.Metrics, error) {

	telemetry, err := metadata.NewTelemetryBuilder(settings.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	sp := &unusedMetricProcessor{
		config:    cfg,
		client:    client,
		logger:    settings.Logger.With(zap.String("component", "unusedmetricprocessor")),
		telemetry: telemetry,
	}

	return processorhelper.NewMetrics(ctx,
		settings,
		cfg,
		nextConsumer,
		sp.processMetrics,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}))
}

func (sp *unusedMetricProcessor) shouldRemoveDatapoint(
	ctx context.Context,
	job string,
	metricName string) bool {

	now := time.Now()
	response, err := sp.client.GetMetricUsage(ctx, job, metricName)
	duration := time.Since(now)
	sp.telemetry.OtelcolProcessorUnusedmetricBackendDuration.Record(
		ctx,
		int64(duration.Seconds()),
	)

	if err != nil {
		sp.logger.Error("error getting metric usage",
			zap.String("job", job),
			zap.String("metric", metricName),
			zap.Error(err),
		)
		sp.telemetry.OtelcolProcessorUnusedmetricError.Add(
			ctx,
			1,
			metric.WithAttributes(attribute.String("job", job)),
		)
		return false
	}
	if response.Unused {
		sp.telemetry.OtelcolProcessorUnusedmetricDropped.Add(
			ctx,
			1,
			metric.WithAttributes(attribute.String("job", job)),
		)
		sp.logger.Debug("metric is unused",
			zap.String("job", job),
			zap.String("metric", metricName),
		)
		return true
	}
	sp.telemetry.OtelcolProcessorUnusedmetricKept.Add(
		ctx,
		1,
		metric.WithAttributes(attribute.String("job", job)),
	)
	return false
}

func (sp *unusedMetricProcessor) processNumberDataPoint(
	ctx context.Context,
	dp pmetric.NumberDataPoint,
	dps pmetric.NumberDataPointSlice,
	metricName string,
) bool {
	job := ""
	if j, ok := dp.Attributes().Get("job"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("service.name"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("scrape_job"); ok {
		job = j.AsString()
	}

	if sp.shouldRemoveDatapoint(ctx, job, metricName) {
		sp.telemetry.OtelcolProcessorUnusedmetricDroppedDatapoints.Add(
			ctx,
			int64(dps.Len()),
			metric.WithAttributes(attribute.String("job", job)),
		)
		return true
	}
	return false
}

func (sp *unusedMetricProcessor) processExponentialHistogramDataPoint(
	ctx context.Context,
	dp pmetric.ExponentialHistogramDataPoint,
	dps pmetric.ExponentialHistogramDataPointSlice,
	metricName string,
) bool {
	job := ""
	if j, ok := dp.Attributes().Get("job"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("service.name"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("scrape_job"); ok {
		job = j.AsString()
	}

	if sp.shouldRemoveDatapoint(ctx, job, metricName) {
		sp.telemetry.OtelcolProcessorUnusedmetricDroppedDatapoints.Add(
			ctx,
			int64(dps.Len()),
			metric.WithAttributes(attribute.String("job", job)),
		)
		return true
	}
	return false
}

func (sp *unusedMetricProcessor) processHistogramDataPoint(
	ctx context.Context,
	dp pmetric.HistogramDataPoint,
	dps pmetric.HistogramDataPointSlice,
	metricName string,
) bool {
	job := ""
	if j, ok := dp.Attributes().Get("job"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("service.name"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("scrape_job"); ok {
		job = j.AsString()
	}

	if sp.shouldRemoveDatapoint(ctx, job, metricName) {
		sp.telemetry.OtelcolProcessorUnusedmetricDroppedDatapoints.Add(
			ctx,
			int64(dps.Len()),
			metric.WithAttributes(attribute.String("job", job)),
		)
		return true
	}
	return false
}

func (sp *unusedMetricProcessor) processSummaryDataPoint(
	ctx context.Context,
	dp pmetric.SummaryDataPoint,
	dps pmetric.SummaryDataPointSlice,
	metricName string,
) bool {
	job := ""
	if j, ok := dp.Attributes().Get("job"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("service.name"); ok {
		job = j.AsString()
	}
	if j, ok := dp.Attributes().Get("scrape_job"); ok {
		job = j.AsString()
	}

	if sp.shouldRemoveDatapoint(ctx, job, metricName) {
		sp.telemetry.OtelcolProcessorUnusedmetricDroppedDatapoints.Add(
			ctx,
			int64(dps.Len()),
			metric.WithAttributes(attribute.String("job", job)),
		)
		return true
	}
	return false
}

func (sp *unusedMetricProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	md.ResourceMetrics().RemoveIf(func(rm pmetric.ResourceMetrics) bool {
		rm.ScopeMetrics().RemoveIf(func(sm pmetric.ScopeMetrics) bool {
			sm.Metrics().RemoveIf(func(m pmetric.Metric) bool {
				metricName := m.Name()
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					m.Gauge().DataPoints().RemoveIf(func(dp pmetric.NumberDataPoint) bool {
						return sp.processNumberDataPoint(ctx, dp, m.Gauge().DataPoints(), metricName)
					})
					return m.Gauge().DataPoints().Len() == 0
				case pmetric.MetricTypeSum:
					m.Sum().DataPoints().RemoveIf(func(dp pmetric.NumberDataPoint) bool {
						return sp.processNumberDataPoint(ctx, dp, m.Sum().DataPoints(), metricName)
					})
					return m.Sum().DataPoints().Len() == 0
				case pmetric.MetricTypeExponentialHistogram:
					m.ExponentialHistogram().DataPoints().RemoveIf(func(dp pmetric.ExponentialHistogramDataPoint) bool {
						return sp.processExponentialHistogramDataPoint(ctx, dp, m.ExponentialHistogram().DataPoints(), metricName)
					})
					return m.ExponentialHistogram().DataPoints().Len() == 0
				case pmetric.MetricTypeHistogram:
					m.Histogram().DataPoints().RemoveIf(func(dp pmetric.HistogramDataPoint) bool {
						return sp.processHistogramDataPoint(ctx, dp, m.Histogram().DataPoints(), metricName)
					})
					return m.Histogram().DataPoints().Len() == 0
				case pmetric.MetricTypeSummary:
					m.Summary().DataPoints().RemoveIf(func(dp pmetric.SummaryDataPoint) bool {
						return sp.processSummaryDataPoint(ctx, dp, m.Summary().DataPoints(), metricName)
					})
					return m.Summary().DataPoints().Len() == 0
				}
				return false
			})
			return sm.Metrics().Len() == 0
		})
		return rm.ScopeMetrics().Len() == 0
	})

	return md, nil
}
