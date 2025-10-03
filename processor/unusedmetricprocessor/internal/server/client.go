package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	GetMetricUsage(ctx context.Context, job string, name string) (MetricUsage, error)
}

type Config struct {
	Address   string         `mapstructure:"address"`
	Timeout   *time.Duration `mapstructure:"timeout"`
	TlsConfig TLSConfig      `mapstructure:"tls_config"`
}

type TLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

type client struct {
	client *http.Client
	config *Config
}

func NewClient(config *Config) Client {
	return &client{
		config: config,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: config.TlsConfig.InsecureSkipVerify,
				},
			},
			Timeout: *config.Timeout,
		},
	}
}

type MetricUsageResponse struct {
	Data []MetricUsage `json:"data"`
}

type MetricUsage struct {
	Name    string              `json:"name"`
	Unused  bool                `json:"unused"`
	Summary *MetricUsageSummary `json:"summary"`
}

type MetricUsageSummary struct {
	AlertCount     int `json:"alert_count"`
	RecordCount    int `json:"record_count"`
	DashboardCount int `json:"dashboard_count"`
	QueryCount     int `json:"query_count"`
}

// /api/v1/metrics/unused?job=myJob&name=http_requests_total
func (c *client) GetMetricUsage(ctx context.Context, job string, name string) (MetricUsage, error) {
	url := c.config.Address + "/api/v1/metrics/unused?job=" + job + "&name=" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return MetricUsage{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return MetricUsage{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MetricUsage{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response MetricUsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return MetricUsage{}, err
	}

	if len(response.Data) == 0 {
		return MetricUsage{
			Name:   name,
			Unused: false,
		}, nil
	}

	return response.Data[0], nil
}
