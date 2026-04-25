package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	ServerURL         string
	ClusterToken      string
	ClusterID         string
	HeartbeatInterval time.Duration
	SyncInterval      time.Duration
	KubeconfigPath    string
}

type Agent struct {
	cfg    *Config
	client *http.Client
}

func New(cfg *Config) *Agent {
	return &Agent{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *Agent) Run(ctx context.Context) error {
	if a.cfg.ClusterID == "" {
		id, err := a.register(ctx)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}
		a.cfg.ClusterID = id
		log.Printf("Registered with cluster ID: %s", id)
	}

	heartbeatTicker := time.NewTicker(a.cfg.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	syncTicker := time.NewTicker(a.cfg.SyncInterval)
	defer syncTicker.Stop()

	if err := a.sendHeartbeat(ctx); err != nil {
		log.Printf("Initial heartbeat failed: %v", err)
	}
	if err := a.syncDesiredState(ctx); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartbeatTicker.C:
			if err := a.sendHeartbeat(ctx); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
		case <-syncTicker.C:
			if err := a.syncDesiredState(ctx); err != nil {
				log.Printf("Sync failed: %v", err)
			}
		}
	}
}

func (a *Agent) register(ctx context.Context) (string, error) {
	body := map[string]string{
		"token": a.cfg.ClusterToken,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", a.cfg.ServerURL+"/api/v1/clusters", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cluster-Token", a.cfg.ClusterToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("registration failed: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	return "", fmt.Errorf("invalid registration response")
}

func (a *Agent) sendHeartbeat(ctx context.Context) error {
	metrics, err := a.collectMetrics()
	if err != nil {
		log.Printf("Failed to collect metrics: %v", err)
		metrics = &ClusterMetrics{}
	}

	body := map[string]interface{}{
		"token":        a.cfg.ClusterToken,
		"node_count":   metrics.NodeCount,
		"cpu_usage":    metrics.CPUUsage,
		"mem_usage":    metrics.MemUsage,
		"pod_count":    metrics.PodCount,
		"version":      metrics.Version,
		"cpu_capacity": metrics.CPUCapacity,
		"mem_capacity": metrics.MemCapacity,
	}
	jsonBody, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/api/v1/clusters/%s/heartbeat", a.cfg.ServerURL, a.cfg.ClusterID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cluster-Token", a.cfg.ClusterToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed: %s", string(respBody))
	}

	log.Println("Heartbeat sent successfully")
	return nil
}

func (a *Agent) syncDesiredState(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/clusters/%s/desired-state", a.cfg.ServerURL, a.cfg.ClusterID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Cluster-Token", a.cfg.ClusterToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get desired state failed: %s", string(respBody))
	}

	var state DesiredStateResponse
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return err
	}

	var syncResults []SyncResultItem

	for _, dep := range state.Deployments {
		result := a.reconcileDeployment(dep)
		syncResults = append(syncResults, result)
	}

	for _, removedID := range state.Removed {
		result := a.uninstallDeployment(removedID)
		syncResults = append(syncResults, result)
	}

	if len(syncResults) > 0 {
		if err := a.reportSyncResult(ctx, syncResults); err != nil {
			log.Printf("Failed to report sync result: %v", err)
		}
	}

	log.Printf("Sync completed: %d deployments processed", len(state.Deployments))
	return nil
}

func (a *Agent) reconcileDeployment(dep DeploymentDesiredState) SyncResultItem {
	result := SyncResultItem{
		DeploymentID: dep.DeploymentID,
	}

	switch dep.Action {
	case "install":
		output, err := a.helmInstall(dep)
		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("install failed: %v, output: %s", err, output)
		} else {
			result.Status = "synced"
			result.ActualVersion = dep.ChartVersion
			result.Message = "installed successfully"
		}
	case "upgrade":
		output, err := a.helmUpgrade(dep)
		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("upgrade failed: %v, output: %s", err, output)
		} else {
			result.Status = "synced"
			result.ActualVersion = dep.ChartVersion
			result.Message = "upgraded successfully"
		}
	case "sync":
		result.Status = "synced"
		result.ActualVersion = dep.ChartVersion
		result.Message = "already in sync"
	default:
		result.Status = "error"
		result.Message = fmt.Sprintf("unknown action: %s", dep.Action)
	}

	return result
}

func (a *Agent) uninstallDeployment(deploymentID string) SyncResultItem {
	return SyncResultItem{
		DeploymentID: deploymentID,
		Status:       "synced",
		Message:      "uninstalled",
	}
}

func (a *Agent) helmInstall(dep DeploymentDesiredState) (string, error) {
	args := []string{"install", dep.Name, dep.ChartName,
		"--namespace", dep.Namespace,
		"--version", dep.ChartVersion,
		"--create-namespace",
	}

	if dep.ChartRepo != "" {
		args = append(args, "--repo", dep.ChartRepo)
	}

	if dep.Values != "" {
		valuesFile, err := a.writeTempValuesFile(dep.Values)
		if err == nil {
			defer os.Remove(valuesFile)
			args = append(args, "-f", valuesFile)
		}
	}

	return a.runCommand("helm", args...)
}

func (a *Agent) helmUpgrade(dep DeploymentDesiredState) (string, error) {
	args := []string{"upgrade", dep.Name, dep.ChartName,
		"--namespace", dep.Namespace,
		"--version", dep.ChartVersion,
		"--create-namespace",
	}

	if dep.ChartRepo != "" {
		args = append(args, "--repo", dep.ChartRepo)
	}

	if dep.Values != "" {
		valuesFile, err := a.writeTempValuesFile(dep.Values)
		if err == nil {
			defer os.Remove(valuesFile)
			args = append(args, "-f", valuesFile)
		}
	}

	return a.runCommand("helm", args...)
}

func (a *Agent) collectMetrics() (*ClusterMetrics, error) {
	metrics := &ClusterMetrics{}

	output, err := a.runCommand("kubectl", "get", "nodes", "--no-headers")
	if err == nil && output != "" {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		metrics.NodeCount = len(lines)
	}

	output, err = a.runCommand("kubectl", "get", "pods", "--all-namespaces", "--no-headers")
	if err == nil && output != "" {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		metrics.PodCount = len(lines)
	}

	output, err = a.runCommand("kubectl", "version", "--short", "-o", "json")
	if err == nil {
		var versionInfo map[string]interface{}
		if json.Unmarshal([]byte(output), &versionInfo) == nil {
			if sv, ok := versionInfo["serverVersion"].(map[string]interface{}); ok {
				if gitVer, ok := sv["gitVersion"].(string); ok {
					metrics.Version = gitVer
				}
			}
		}
	}

	return metrics, nil
}

func (a *Agent) reportSyncResult(ctx context.Context, results []SyncResultItem) error {
	metrics, _ := a.collectMetrics()

	body := map[string]interface{}{
		"results":         results,
		"cluster_metrics": metrics,
	}
	jsonBody, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/api/v1/clusters/%s/sync-result", a.cfg.ServerURL, a.cfg.ClusterID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cluster-Token", a.cfg.ClusterToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (a *Agent) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if a.cfg.KubeconfigPath != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+a.cfg.KubeconfigPath)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}

func (a *Agent) writeTempValuesFile(content string) (string, error) {
	tmpDir := os.TempDir()
	f, err := os.CreateTemp(tmpDir, "kubenexus-values-*.yaml")
	if err != nil {
		return "", err
	}
	filename := f.Name()
	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		os.Remove(filename)
		return "", err
	}
	f.Close()
	return filename, nil
}

type ClusterMetrics struct {
	NodeCount   int     `json:"node_count"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemUsage    float64 `json:"mem_usage"`
	PodCount    int     `json:"pod_count"`
	Version     string  `json:"version"`
	CPUCapacity string  `json:"cpu_capacity"`
	MemCapacity string  `json:"mem_capacity"`
}

type DesiredStateResponse struct {
	ClusterID   string                   `json:"cluster_id"`
	Deployments []DeploymentDesiredState `json:"deployments"`
	Removed     []string                 `json:"removed"`
}

type DeploymentDesiredState struct {
	DeploymentID string `json:"deployment_id"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ChartName    string `json:"chart_name"`
	ChartRepo    string `json:"chart_repo"`
	ChartVersion string `json:"chart_version"`
	Values       string `json:"values"`
	Action       string `json:"action"`
}

type SyncResultItem struct {
	DeploymentID   string `json:"deployment_id"`
	Status         string `json:"status"`
	ActualVersion  string `json:"actual_version"`
	ActualReplicas int    `json:"actual_replicas"`
	Message        string `json:"message"`
	DriftDetail    string `json:"drift_detail"`
}
