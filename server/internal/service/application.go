package service

import (
	"errors"
	"time"

	"github.com/kubenexus/server/internal/model"
	"github.com/kubenexus/server/internal/store"
	"github.com/google/uuid"
)

type ApplicationService struct {
	store *store.Store
}

func NewApplicationService(s *store.Store) *ApplicationService {
	return &ApplicationService{store: s}
}

func (svc *ApplicationService) CreateApplication(req CreateApplicationRequest) (*model.Application, error) {
	a := &model.Application{
		ID:            uuid.New().String(),
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		Icon:          req.Icon,
		ChartName:     req.ChartName,
		ChartRepo:     req.ChartRepo,
		ChartVersion:  req.ChartVersion,
		Category:      req.Category,
		IsSaaS:        req.IsSaaS,
		DefaultValues: req.DefaultValues,
	}
	if err := svc.store.CreateApplication(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (svc *ApplicationService) GetApplication(id string) (*model.Application, error) {
	return svc.store.GetApplication(id)
}

func (svc *ApplicationService) ListApplications() ([]model.Application, error) {
	return svc.store.ListApplications()
}

func (svc *ApplicationService) UpdateApplication(id string, req UpdateApplicationRequest) (*model.Application, error) {
	app, err := svc.store.GetApplication(id)
	if err != nil {
		return nil, err
	}
	if req.DisplayName != "" {
		app.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		app.Description = req.Description
	}
	if req.Icon != "" {
		app.Icon = req.Icon
	}
	if req.ChartName != "" {
		app.ChartName = req.ChartName
	}
	if req.ChartRepo != "" {
		app.ChartRepo = req.ChartRepo
	}
	if req.ChartVersion != "" {
		app.ChartVersion = req.ChartVersion
	}
	if req.Category != "" {
		app.Category = req.Category
	}
	if req.DefaultValues != "" {
		app.DefaultValues = req.DefaultValues
	}
	if err := svc.store.UpdateApplication(app); err != nil {
		return nil, err
	}
	return app, nil
}

func (svc *ApplicationService) DeleteApplication(id string) error {
	return svc.store.DeleteApplication(id)
}

type CreateApplicationRequest struct {
	Name          string `json:"name" binding:"required"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description"`
	Icon          string `json:"icon"`
	ChartName     string `json:"chart_name" binding:"required"`
	ChartRepo     string `json:"chart_repo"`
	ChartVersion  string `json:"chart_version"`
	Category      string `json:"category"`
	IsSaaS        bool   `json:"is_saas"`
	DefaultValues string `json:"default_values"`
}

type UpdateApplicationRequest struct {
	DisplayName   string `json:"display_name"`
	Description   string `json:"description"`
	Icon          string `json:"icon"`
	ChartName     string `json:"chart_name"`
	ChartRepo     string `json:"chart_repo"`
	ChartVersion  string `json:"chart_version"`
	Category      string `json:"category"`
	DefaultValues string `json:"default_values"`
}

type DeploymentService struct {
	store *store.Store
}

func NewDeploymentService(s *store.Store) *DeploymentService {
	return &DeploymentService{store: s}
}

func (svc *DeploymentService) Deploy(req DeployRequest) (*model.Deployment, error) {
	lic, err := svc.store.GetLicense()
	if err != nil {
		return nil, errors.New("license not found, cannot create deployment")
	}
	if !lic.IsValid || time.Now().After(lic.ExpiresAt) {
		return nil, errors.New("license expired or invalid, cannot create new deployment")
	}
	count, _ := svc.store.CountDeployments()
	if count >= int64(lic.MaxDeployments) {
		return nil, errors.New("deployment count reached license limit")
	}

	_, err = svc.store.GetCluster(req.ClusterID)
	if err != nil {
		return nil, errors.New("cluster not found")
	}
	_, err = svc.store.GetApplication(req.ApplicationID)
	if err != nil {
		return nil, errors.New("application not found")
	}

	d := &model.Deployment{
		ID:            uuid.New().String(),
		ClusterID:     req.ClusterID,
		ApplicationID: req.ApplicationID,
		Name:          req.Name,
		Namespace:     req.Namespace,
		Values:        req.Values,
		Status:        "pending",
		Replicas:      req.Replicas,
		Version:       req.Version,
	}
	if d.Namespace == "" {
		d.Namespace = "default"
	}
	if d.Replicas == 0 {
		d.Replicas = 1
	}
	if err := svc.store.CreateDeployment(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (svc *DeploymentService) BatchDeploy(req BatchDeployRequest) ([]model.Deployment, error) {
	app, err := svc.store.GetApplication(req.ApplicationID)
	if err != nil {
		return nil, errors.New("application not found")
	}

	var clusters []model.Cluster
	if len(req.ClusterSelector.Labels) > 0 {
		clusters, err = svc.store.ListClustersByLabels(req.ClusterSelector.Labels)
		if err != nil {
			return nil, err
		}
	}
	if len(req.ClusterIDs) > 0 {
		seen := make(map[string]bool)
		for _, c := range clusters {
			seen[c.ID] = true
		}
		for _, cid := range req.ClusterIDs {
			if seen[cid] {
				continue
			}
			c, err := svc.store.GetCluster(cid)
			if err != nil {
				continue
			}
			clusters = append(clusters, *c)
		}
	}

	var results []model.Deployment
	for _, cluster := range clusters {
		values := app.DefaultValues
		if req.ValuesOverrides != "" {
			values = req.ValuesOverrides
		}
		d := &model.Deployment{
			ID:            uuid.New().String(),
			ClusterID:     cluster.ID,
			ApplicationID: req.ApplicationID,
			Name:          req.Name,
			Namespace:     req.Namespace,
			Values:        values,
			Status:        "pending",
			Replicas:      req.Replicas,
			Version:       app.ChartVersion,
		}
		if d.Namespace == "" {
			d.Namespace = "default"
		}
		if d.Replicas == 0 {
			d.Replicas = 1
		}
		if err := svc.store.CreateDeployment(d); err != nil {
			continue
		}
		results = append(results, *d)
	}
	return results, nil
}

func (svc *DeploymentService) GetDeployment(id string) (*model.Deployment, error) {
	return svc.store.GetDeployment(id)
}

func (svc *DeploymentService) ListDeployments(clusterID string) ([]model.Deployment, error) {
	return svc.store.ListDeployments(clusterID)
}

func (svc *DeploymentService) UpdateDeployment(id string, req UpdateDeploymentRequest) (*model.Deployment, error) {
	dep, err := svc.store.GetDeployment(id)
	if err != nil {
		return nil, err
	}
	if req.Values != "" {
		dep.Values = req.Values
	}
	if req.Replicas > 0 {
		dep.Replicas = req.Replicas
	}
	if req.Version != "" {
		dep.Version = req.Version
	}
	dep.Status = "pending"
	if err := svc.store.UpdateDeployment(dep); err != nil {
		return nil, err
	}
	return dep, nil
}

func (svc *DeploymentService) UpdateStatus(id string, status string, message string) error {
	d, err := svc.store.GetDeployment(id)
	if err != nil {
		return err
	}
	d.Status = status
	d.Message = message
	return svc.store.UpdateDeployment(d)
}

func (svc *DeploymentService) DeleteDeployment(id string) error {
	return svc.store.DeleteDeployment(id)
}

type DeployRequest struct {
	ClusterID     string `json:"cluster_id" binding:"required"`
	ApplicationID string `json:"application_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Namespace     string `json:"namespace"`
	Values        string `json:"values"`
	Replicas      int    `json:"replicas"`
	Version       string `json:"version"`
}

type BatchDeployRequest struct {
	ApplicationID   string           `json:"application_id" binding:"required"`
	Name            string           `json:"name" binding:"required"`
	Namespace       string           `json:"namespace"`
	ClusterIDs      []string         `json:"cluster_ids"`
	ClusterSelector ClusterSelector  `json:"cluster_selector"`
	ValuesOverrides string           `json:"values_overrides"`
	Replicas        int              `json:"replicas"`
}

type ClusterSelector struct {
	Labels map[string]string `json:"labels"`
}

type UpdateDeploymentRequest struct {
	Values   string `json:"values"`
	Replicas int    `json:"replicas"`
	Version  string `json:"version"`
}
