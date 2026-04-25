package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kubenexus/server/internal/model"
	"github.com/kubenexus/server/internal/store"
)

type AlertEngine struct {
	store *store.Store
}

func NewAlertEngine(s *store.Store) *AlertEngine {
	return &AlertEngine{store: s}
}

func (e *AlertEngine) Start() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			e.evaluate()
		}
	}()
}

func (e *AlertEngine) evaluate() {
	rules, err := e.store.ListAlertRules()
	if err != nil {
		return
	}

	clusters, err := e.store.ListClusters()
	if err != nil {
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		for _, cluster := range clusters {
			e.evaluateRuleForCluster(rule, cluster)
		}
	}

	e.evaluateLicenseRules(rules)
}

func (e *AlertEngine) evaluateRuleForCluster(rule model.AlertRule, cluster model.Cluster) {
	switch rule.Type {
	case "cluster_down":
		if cluster.Status == "unavailable" {
			e.fireAlert(rule, cluster.ID, "集群 "+cluster.Name+" 处于离线状态")
		} else {
			e.resolveAlert(rule.ID, cluster.ID)
		}
	case "cpu_high":
		hb, err := e.store.GetLatestHeartbeat(cluster.ID)
		if err != nil {
			return
		}
		threshold := parseThresholdFloat(rule.Condition, 80)
		if hb.CPUUsage > threshold {
			e.fireAlert(rule, cluster.ID, "集群 "+cluster.Name+" CPU使用率 %.1f%% 超过阈值 %.1f%%", hb.CPUUsage, threshold)
		} else {
			e.resolveAlert(rule.ID, cluster.ID)
		}
	case "mem_high":
		hb, err := e.store.GetLatestHeartbeat(cluster.ID)
		if err != nil {
			return
		}
		threshold := parseThresholdFloat(rule.Condition, 80)
		if hb.MemUsage > threshold {
			e.fireAlert(rule, cluster.ID, "集群 "+cluster.Name+" 内存使用率 %.1f%% 超过阈值 %.1f%%", hb.MemUsage, threshold)
		} else {
			e.resolveAlert(rule.ID, cluster.ID)
		}
	case "drift_detected":
		deployments, err := e.store.ListDeploymentsByCluster(cluster.ID)
		if err != nil {
			return
		}
		for _, d := range deployments {
			if d.Status == "drifted" {
				e.fireAlert(rule, cluster.ID, "集群 "+cluster.Name+" 部署 "+d.Name+" 检测到配置漂移")
				return
			}
		}
		e.resolveAlert(rule.ID, cluster.ID)
	}
}

func (e *AlertEngine) evaluateLicenseRules(rules []model.AlertRule) {
	for _, rule := range rules {
		if !rule.Enabled || rule.Type != "license_expiring" {
			continue
		}
		lic, err := e.store.GetLicense()
		if err != nil {
			continue
		}
		daysLeft := time.Until(lic.ExpiresAt).Hours() / 24
		threshold := parseThresholdFloat(rule.Condition, 30)
		if daysLeft < threshold {
			e.fireAlert(rule, "", "License 将在 %.0f 天后过期", daysLeft)
		} else {
			e.resolveAlert(rule.ID, "")
		}
	}
}

func (e *AlertEngine) fireAlert(rule model.AlertRule, clusterID string, format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	_, err := e.store.GetFiringAlertRecord(rule.ID, clusterID)
	if err == nil {
		return
	}

	record := &model.AlertRecord{
		ID:          uuid.New().String(),
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		ClusterID:   clusterID,
		Severity:    rule.Severity,
		Message:     msg,
		Status:      "firing",
		TriggeredAt: time.Now(),
	}
	if err := e.store.CreateAlertRecord(record); err != nil {
		log.Printf("Failed to create alert record: %v", err)
		return
	}

	rule.LastTriggered = time.Now()
	e.store.UpdateAlertRule(&rule)
}

func (e *AlertEngine) resolveAlert(ruleID string, clusterID string) {
	records, err := e.store.ListFiringAlertRecords(ruleID, clusterID)
	if err != nil {
		return
	}
	now := time.Now()
	for i := range records {
		records[i].Status = "resolved"
		records[i].ResolvedAt = now
		e.store.UpdateAlertRecord(&records[i])
	}
}

func parseThresholdFloat(condition string, defaultVal float64) float64 {
	var cond map[string]interface{}
	if err := json.Unmarshal([]byte(condition), &cond); err != nil {
		return defaultVal
	}
	if t, ok := cond["threshold"]; ok {
		switch v := t.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultVal
}
