package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kubenexus/server/internal/model"
	"github.com/kubenexus/server/internal/store"
	"github.com/google/uuid"
)

type ClusterService struct {
	store *store.Store
}

func NewClusterService(s *store.Store) *ClusterService {
	return &ClusterService{store: s}
}

func (svc *ClusterService) RegisterCluster(req RegisterClusterRequest) (*model.Cluster, error) {
	lic, err := svc.store.GetLicense()
	if err != nil {
		return nil, errors.New("license not found, cannot register cluster")
	}
	if !lic.IsValid || time.Now().After(lic.ExpiresAt) {
		return nil, errors.New("license expired or invalid, cannot register new cluster")
	}
	count, _ := svc.store.CountClusters()
	if count >= int64(lic.MaxClusters) {
		return nil, fmt.Errorf("cluster count reached license limit (%d)", lic.MaxClusters)
	}

	token := fmt.Sprintf("cn-%s", uuid.New().String())
	labels := make(model.StringMap)
	for k, v := range req.Labels {
		labels[k] = v
	}
	if labels == nil {
		labels = make(model.StringMap)
	}
	orgName := ""
	if req.OrgID != "" {
		org, err := svc.store.GetOrganization(req.OrgID)
		if err == nil {
			orgName = org.Name
		}
	}
	c := &model.Cluster{
		ID:           uuid.New().String(),
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Status:       "registered",
		Token:        token,
		OrgID:        req.OrgID,
		OrgName:      orgName,
		Region:       req.Region,
		Labels:       labels,
	}
	if err := svc.store.CreateCluster(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (svc *ClusterService) GetCluster(id string) (*model.Cluster, error) {
	return svc.store.GetCluster(id)
}

func (svc *ClusterService) ListClusters() ([]model.Cluster, error) {
	return svc.store.ListClusters()
}

func (svc *ClusterService) DeleteCluster(id string) error {
	return svc.store.DeleteCluster(id)
}

func (svc *ClusterService) UpdateLabels(id string, labels map[string]string) (*model.Cluster, error) {
	cluster, err := svc.store.GetCluster(id)
	if err != nil {
		return nil, err
	}
	newLabels := make(model.StringMap)
	for k, v := range labels {
		newLabels[k] = v
	}
	cluster.Labels = newLabels
	if err := svc.store.UpdateCluster(cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (svc *ClusterService) RotateToken(id string) (*model.Cluster, error) {
	cluster, err := svc.store.GetCluster(id)
	if err != nil {
		return nil, err
	}
	cluster.Token = fmt.Sprintf("cn-%s", uuid.New().String())
	if err := svc.store.UpdateCluster(cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (svc *ClusterService) HandleHeartbeat(clusterID string, req HeartbeatRequest) error {
	cluster, err := svc.store.GetCluster(clusterID)
	if err != nil {
		return err
	}

	hb := &model.Heartbeat{
		ID:         uuid.New().String(),
		ClusterID:  clusterID,
		NodeCount:  req.NodeCount,
		CPUUsage:   req.CPUUsage,
		MemUsage:   req.MemUsage,
		PodCount:   req.PodCount,
		Version:    req.Version,
		Info:       req.Info,
		ReportedAt: time.Now(),
	}
	if err := svc.store.CreateHeartbeat(hb); err != nil {
		return err
	}

	cluster.Status = "active"
	cluster.NodeCount = req.NodeCount
	cluster.Version = req.Version
	cluster.LastHeartbeat = time.Now()
	if req.CPUCapacity != "" {
		cluster.CPUCapacity = req.CPUCapacity
	}
	if req.MemCapacity != "" {
		cluster.MemCapacity = req.MemCapacity
	}
	return svc.store.UpdateCluster(cluster)
}

func (svc *ClusterService) UpdateClusterStatus(cluster *model.Cluster) error {
	return svc.store.UpdateCluster(cluster)
}

func (svc *ClusterService) GetInstallScript(id string, serverURL string) (string, error) {
	cluster, err := svc.store.GetCluster(id)
	if err != nil {
		return "", err
	}
	script := fmt.Sprintf(`#!/bin/sh
set -e

echo "=== 衡牧KubeNexusK3s多集群管理系统 Agent Installer ==="

if ! command -v k3s >/dev/null 2>&1; then
    echo "Installing K3s..."
    curl -sfL https://get.k3s.io | sh -
    echo "K3s installed successfully."
else
    echo "K3s already installed, skipping."
fi

echo "Deploying 衡牧KubeNexusK3s多集群管理系统 Agent..."
kubectl apply -f "%s/api/v1/clusters/%s/registration.yaml"

echo "Agent deployed. Waiting for registration..."
echo "You can check status with: kubectl get pods -n kubenexus-system"
`, serverURL, cluster.ID)
	return script, nil
}

func (svc *ClusterService) GetRegistrationYAML(id string, serverURL string) (string, error) {
	cluster, err := svc.store.GetCluster(id)
	if err != nil {
		return "", err
	}
	yaml := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: kubenexus-system
---
apiVersion: v1
kind: Secret
metadata:
  name: kubenexus-agent-config
  namespace: kubenexus-system
type: Opaque
stringData:
  SERVER_URL: "%s"
  CLUSTER_TOKEN: "%s"
  CLUSTER_ID: "%s"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubenexus-agent
  namespace: kubenexus-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubenexus-agent
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubenexus-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubenexus-agent
subjects:
- kind: ServiceAccount
  name: kubenexus-agent
  namespace: kubenexus-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubenexus-agent
  namespace: kubenexus-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubenexus-agent
  template:
    metadata:
      labels:
        app: kubenexus-agent
    spec:
      serviceAccountName: kubenexus-agent
      containers:
      - name: agent
        image: kubenexus/agent:latest
        env:
        - name: SERVER_URL
          valueFrom:
            secretKeyRef:
              name: kubenexus-agent-config
              key: SERVER_URL
        - name: CLUSTER_TOKEN
          valueFrom:
            secretKeyRef:
              name: kubenexus-agent-config
              key: CLUSTER_TOKEN
        - name: CLUSTER_ID
          valueFrom:
            secretKeyRef:
              name: kubenexus-agent-config
              key: CLUSTER_ID
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 50m
            memory: 64Mi
`, serverURL, cluster.Token, cluster.ID)
	return yaml, nil
}

func (svc *ClusterService) GetDesiredState(clusterID string) (*DesiredStateResponse, error) {
	cluster, err := svc.store.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	if cluster.Status != "active" && cluster.Status != "registered" {
		return nil, errors.New("cluster not available")
	}

	deployments, err := svc.store.ListDeploymentsByCluster(clusterID)
	if err != nil {
		return nil, err
	}

	var items []DeploymentDesiredState
	var removed []string

	for _, d := range deployments {
		if d.Status == "stopped" {
			removed = append(removed, d.ID)
			continue
		}
		app, err := svc.store.GetApplication(d.ApplicationID)
		if err != nil {
			continue
		}
		action := "install"
		if d.ActualStatus != "" {
			if d.Version != d.ActualVersion {
				action = "upgrade"
			} else {
				action = "sync"
			}
		}
		items = append(items, DeploymentDesiredState{
			DeploymentID: d.ID,
			Name:         d.Name,
			Namespace:    d.Namespace,
			ChartName:    app.ChartName,
			ChartRepo:    app.ChartRepo,
			ChartVersion: d.Version,
			Values:       d.Values,
			Action:       action,
		})
	}

	return &DesiredStateResponse{
		ClusterID:   clusterID,
		Deployments: items,
		Removed:     removed,
	}, nil
}

func (svc *ClusterService) HandleSyncResult(clusterID string, req SyncResultRequest) error {
	cluster, err := svc.store.GetCluster(clusterID)
	if err != nil {
		return err
	}

	for _, r := range req.Results {
		dep, err := svc.store.GetDeployment(r.DeploymentID)
		if err != nil {
			continue
		}
		dep.ActualStatus = r.Status
		dep.ActualVersion = r.ActualVersion
		dep.Message = r.Message
		dep.LastSynced = time.Now()
		if r.Status == "synced" {
			dep.Status = "synced"
		} else if r.Status == "error" {
			dep.Status = "error"
		} else if r.Status == "drifted" {
			dep.Status = "drifted"
			dep.DriftDetail = r.DriftDetail
		} else if r.Status == "syncing" {
			dep.Status = "syncing"
		}
		if err := svc.store.UpdateDeployment(dep); err != nil {
			log.Printf("Failed to update deployment %s: %v", dep.ID, err)
		}
	}

	if req.ClusterMetrics != nil {
		cluster.NodeCount = req.ClusterMetrics.NodeCount
		cluster.Version = req.ClusterMetrics.Version
		cluster.LastHeartbeat = time.Now()
		cluster.Status = "active"
		if err := svc.store.UpdateCluster(cluster); err != nil {
			log.Printf("Failed to update cluster %s after sync: %v", clusterID, err)
		}
	}

	return nil
}

func (svc *ClusterService) GetClusterMetrics(clusterID string, limit int) ([]model.Heartbeat, error) {
	return svc.store.ListHeartbeats(clusterID, limit)
}

type RegisterClusterRequest struct {
	Name        string            `json:"name" binding:"required"`
	DisplayName string            `json:"display_name"`
	OrgID       string            `json:"org_id"`
	Region      string            `json:"region"`
	Labels      map[string]string `json:"labels"`
}

type HeartbeatRequest struct {
	Token       string  `json:"token" binding:"required"`
	NodeCount   int     `json:"node_count"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemUsage    float64 `json:"mem_usage"`
	PodCount    int     `json:"pod_count"`
	Version     string  `json:"version"`
	CPUCapacity string  `json:"cpu_capacity"`
	MemCapacity string  `json:"mem_capacity"`
	Info        string  `json:"info"`
}

type DesiredStateResponse struct {
	ClusterID   string                  `json:"cluster_id"`
	Deployments []DeploymentDesiredState `json:"deployments"`
	Removed     []string                `json:"removed"`
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

type SyncResultRequest struct {
	Results        []SyncResultItem   `json:"results"`
	ClusterMetrics *ClusterMetricsDTO `json:"cluster_metrics"`
}

type SyncResultItem struct {
	DeploymentID   string `json:"deployment_id"`
	Status         string `json:"status"`
	ActualVersion  string `json:"actual_version"`
	ActualReplicas int    `json:"actual_replicas"`
	Message        string `json:"message"`
	DriftDetail    string `json:"drift_detail"`
}

type ClusterMetricsDTO struct {
	NodeCount int     `json:"node_count"`
	CPUUsage  float64 `json:"cpu_usage"`
	MemUsage  float64 `json:"mem_usage"`
	PodCount  int     `json:"pod_count"`
	Version   string  `json:"version"`
}
