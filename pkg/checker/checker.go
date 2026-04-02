package checker

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/arcane-bear/agent-probe/pkg/config"
)

type Result struct {
	Status    string                 `json:"status"`
	Checks   map[string]CheckResult `json:"checks"`
	Timestamp string                `json:"timestamp,omitempty"`
}

type CheckResult struct {
	Status     string                 `json:"status"`
	LatencyMs  int64                  `json:"latency_ms,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Details    map[string]CheckResult `json:"details,omitempty"`
	RemainingPct *float64             `json:"remaining_pct,omitempty"`
}

type Checker struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Checker {
	return &Checker{cfg: cfg}
}

func (c *Checker) RunAll() *Result {
	result := &Result{
		Status: "healthy",
		Checks: make(map[string]CheckResult),
	}

	if c.cfg.Checks.LLM.Enabled {
		cr := c.checkLLM()
		result.Checks["llm"] = cr
		downgrade(result, cr.Status)
	}

	if c.cfg.Checks.TokenBudget.Enabled {
		cr := c.checkTokenBudget()
		result.Checks["token_budget"] = cr
		downgrade(result, cr.Status)
	}

	if c.cfg.Checks.Tools.Enabled {
		cr := c.checkTools()
		result.Checks["tools"] = cr
		downgrade(result, cr.Status)
	}

	if c.cfg.Checks.ResponseQuality.Enabled {
		cr := c.checkResponseQuality()
		result.Checks["response_quality"] = cr
		downgrade(result, cr.Status)
	}

	return result
}

func (c *Checker) RunReadiness() *Result {
	result := &Result{
		Status: "healthy",
		Checks: make(map[string]CheckResult),
	}

	if c.cfg.Checks.LLM.Enabled {
		cr := c.checkLLM()
		result.Checks["llm"] = cr
		downgrade(result, cr.Status)
	}

	if c.cfg.Checks.Tools.Enabled {
		cr := c.checkTools()
		result.Checks["tools"] = cr
		downgrade(result, cr.Status)
	}

	return result
}

func (c *Checker) checkLLM() CheckResult {
	cfg := c.cfg.Checks.LLM
	apiKey := os.Getenv(cfg.APIKeyEnv)
	if apiKey == "" {
		return CheckResult{Status: "unhealthy", Error: fmt.Sprintf("env var %s not set", cfg.APIKeyEnv)}
	}

	start := time.Now()
	client := &http.Client{Timeout: cfg.Timeout}

	req, err := http.NewRequest("GET", cfg.Endpoint, nil)
	if err != nil {
		return CheckResult{Status: "unhealthy", Error: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return CheckResult{Status: "unhealthy", Error: err.Error(), LatencyMs: latency}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return CheckResult{Status: "unhealthy", Error: "authentication failed", LatencyMs: latency}
	}

	return CheckResult{Status: "healthy", LatencyMs: latency}
}

func (c *Checker) checkTokenBudget() CheckResult {
	cfg := c.cfg.Checks.TokenBudget
	apiKey := os.Getenv(cfg.APIKeyEnv)
	if apiKey == "" {
		return CheckResult{Status: "unhealthy", Error: fmt.Sprintf("env var %s not set", cfg.APIKeyEnv)}
	}

	// Placeholder: in a real implementation, this would query the provider's
	// usage API and calculate remaining budget percentage.
	remaining := 100.0
	result := CheckResult{Status: "healthy", RemainingPct: &remaining}

	if remaining <= cfg.CriticalThreshold {
		result.Status = "unhealthy"
	} else if remaining <= cfg.WarnThreshold {
		result.Status = "warning"
	}

	return result
}

func (c *Checker) checkTools() CheckResult {
	cfg := c.cfg.Checks.Tools
	details := make(map[string]CheckResult)
	overallStatus := "healthy"

	for _, tool := range cfg.Endpoints {
		start := time.Now()
		client := &http.Client{Timeout: tool.Timeout}

		resp, err := client.Get(tool.URL)
		latency := time.Since(start).Milliseconds()

		if err != nil {
			details[tool.Name] = CheckResult{Status: "unhealthy", Error: err.Error(), LatencyMs: latency}
			overallStatus = "unhealthy"
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			details[tool.Name] = CheckResult{Status: "unhealthy", Error: fmt.Sprintf("HTTP %d", resp.StatusCode), LatencyMs: latency}
			overallStatus = "unhealthy"
			continue
		}

		details[tool.Name] = CheckResult{Status: "healthy", LatencyMs: latency}
	}

	return CheckResult{Status: overallStatus, Details: details}
}

func (c *Checker) checkResponseQuality() CheckResult {
	cfg := c.cfg.Checks.ResponseQuality
	apiKey := os.Getenv(cfg.APIKeyEnv)
	if apiKey == "" {
		return CheckResult{Status: "unhealthy", Error: fmt.Sprintf("env var %s not set", cfg.APIKeyEnv)}
	}

	// Placeholder: in a real implementation, this would send a test prompt
	// to the LLM and verify the response contains the expected substring.
	start := time.Now()
	latency := time.Since(start).Milliseconds()

	return CheckResult{Status: "healthy", LatencyMs: latency}
}

func downgrade(result *Result, status string) {
	if status == "unhealthy" {
		result.Status = "unhealthy"
	} else if status == "warning" && result.Status == "healthy" {
		result.Status = "degraded"
	}
}
