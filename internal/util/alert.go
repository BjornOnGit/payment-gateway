package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityError    AlertSeverity = "error"
	SeverityWarning  AlertSeverity = "warning"
	SeverityInfo     AlertSeverity = "info"
)

// Alert represents an alert message
type Alert struct {
	Service   string         `json:"service"`
	Severity  AlertSeverity  `json:"severity"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// AlertClient sends alerts to a webhook endpoint
type AlertClient struct {
	webhookURL string
	logger     *zap.Logger
	httpClient *http.Client
}

// NewAlertClient creates a new alert client
func NewAlertClient(webhookURL string, logger *zap.Logger) *AlertClient {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AlertClient{
		webhookURL: webhookURL,
		logger:     logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert sends an alert to the webhook endpoint
func (ac *AlertClient) SendAlert(ctx context.Context, service string, severity AlertSeverity, message string, details map[string]any) error {
	if ac.webhookURL == "" {
		ac.logger.Debug("alert webhook URL not configured, skipping", zap.String("service", service))
		return nil
	}

	alert := Alert{
		Service:   service,
		Severity:  severity,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UTC(),
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		ac.logger.Error("failed to marshal alert", zap.Error(err))
		return fmt.Errorf("marshal alert: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ac.webhookURL, bytes.NewReader(payload))
	if err != nil {
		ac.logger.Error("failed to create request", zap.Error(err))
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		ac.logger.Error("failed to send alert",
			zap.String("service", service),
			zap.String("severity", string(severity)),
			zap.Error(err),
		)
		return fmt.Errorf("send alert: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		ac.logger.Error("alert webhook returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("service", service),
			zap.String("severity", string(severity)),
			zap.ByteString("response", body),
		)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	ac.logger.Info("alert sent successfully",
		zap.String("service", service),
		zap.String("severity", string(severity)),
		zap.String("message", message),
	)

	return nil
}

// SendCritical sends a critical alert
func (ac *AlertClient) SendCritical(ctx context.Context, service, message string, details map[string]any) error {
	return ac.SendAlert(ctx, service, SeverityCritical, message, details)
}

// SendError sends an error alert
func (ac *AlertClient) SendError(ctx context.Context, service, message string, details map[string]any) error {
	return ac.SendAlert(ctx, service, SeverityError, message, details)
}

// SendWarning sends a warning alert
func (ac *AlertClient) SendWarning(ctx context.Context, service, message string, details map[string]any) error {
	return ac.SendAlert(ctx, service, SeverityWarning, message, details)
}

// SendInfo sends an info alert
func (ac *AlertClient) SendInfo(ctx context.Context, service, message string, details map[string]any) error {
	return ac.SendAlert(ctx, service, SeverityInfo, message, details)
}
