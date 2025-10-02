package unusedmetricprocessor // import "github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor"

import (
	"context"

	"github.com/nicolastakashi/ntakashi-opentelemetry-collector/processor/unusedmetricprocessor/internal/server"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
)

type unusedMetricProcessor struct {
	config *Config
	client server.Client
	component.StartFunc
	component.ShutdownFunc
	logger *zap.Logger
}

func newUnusedMetricProcessor(
	ctx context.Context,
	settings processor.Settings,
	cfg *Config,
	nextConsumer consumer.Metrics,
	client server.Client,
) (processor.Metrics, error) {

	sp := &unusedMetricProcessor{
		config: cfg,
		client: client,
		logger: settings.Logger.With(zap.String("component", "unusedmetricprocessor")),
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
	response, err := sp.client.GetMetricUsage(ctx, job, metricName)
	if err != nil {
		sp.logger.Error("error getting metric usage", zap.Error(err))
		return false
	}
	if response.Unused {
		sp.logger.Debug("metric is unused", zap.String("job", job), zap.String("metricName", metricName))
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
						if job, ok := dp.Attributes().Get("job"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						if job, ok := dp.Attributes().Get("service.name"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						if job, ok := dp.Attributes().Get("scrape_job"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						return sp.shouldRemoveDatapoint(ctx, "", metricName)
					})
					sp.logger.Info("Removing gauge data points", zap.Int("gauge_data_points_len", m.Gauge().DataPoints().Len()))
					return m.Gauge().DataPoints().Len() == 0
				case pmetric.MetricTypeSum:
					m.Sum().DataPoints().RemoveIf(func(dp pmetric.NumberDataPoint) bool {
						if job, ok := dp.Attributes().Get("job"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						if job, ok := dp.Attributes().Get("service.name"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						return sp.shouldRemoveDatapoint(ctx, "", metricName)
					})
					sp.logger.Info("Removing sum data points", zap.Int("sum_data_points_len", m.Sum().DataPoints().Len()))
					return m.Sum().DataPoints().Len() == 0
				case pmetric.MetricTypeExponentialHistogram:
					m.ExponentialHistogram().DataPoints().RemoveIf(func(dp pmetric.ExponentialHistogramDataPoint) bool {
						if job, ok := dp.Attributes().Get("job"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						if job, ok := dp.Attributes().Get("service.name"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						return sp.shouldRemoveDatapoint(ctx, "", metricName)
					})
					sp.logger.Info("Removing exponential histogram data points", zap.Int("exponential_histogram_data_points_len", m.ExponentialHistogram().DataPoints().Len()))
					return m.ExponentialHistogram().DataPoints().Len() == 0
				case pmetric.MetricTypeHistogram:
					m.Histogram().DataPoints().RemoveIf(func(dp pmetric.HistogramDataPoint) bool {
						if job, ok := dp.Attributes().Get("job"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						if job, ok := dp.Attributes().Get("service.name"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						return sp.shouldRemoveDatapoint(ctx, "", metricName)
					})
					sp.logger.Info("Removing histogram data points", zap.Int("histogram_data_points_len", m.Histogram().DataPoints().Len()))
					return m.Histogram().DataPoints().Len() == 0
				case pmetric.MetricTypeSummary:
					m.Summary().DataPoints().RemoveIf(func(dp pmetric.SummaryDataPoint) bool {
						if job, ok := dp.Attributes().Get("job"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						if job, ok := dp.Attributes().Get("service.name"); ok {
							return sp.shouldRemoveDatapoint(ctx, job.AsString(), metricName)
						}
						return sp.shouldRemoveDatapoint(ctx, "", metricName)
					})
					sp.logger.Info("Removing summary data points", zap.Int("summary_data_points_len", m.Summary().DataPoints().Len()))
					return m.Summary().DataPoints().Len() == 0
				}
				return false
			})
			sp.logger.Info("Removing scope metrics", zap.Int("scope_metrics_len", sm.Metrics().Len()))
			return sm.Metrics().Len() == 0
		})
		sp.logger.Info("Removing resource metrics", zap.Int("resource_metrics_len", rm.ScopeMetrics().Len()))
		return rm.ScopeMetrics().Len() == 0
	})

	return md, nil
}
