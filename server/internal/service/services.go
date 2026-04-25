package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/kubenexus/server/internal/model"
	"github.com/kubenexus/server/internal/store"
	"github.com/google/uuid"
)

type OrganizationService struct {
	store *store.Store
}

func NewOrganizationService(s *store.Store) *OrganizationService {
	return &OrganizationService{store: s}
}

func (svc *OrganizationService) CreateOrganization(req CreateOrganizationRequest) (*model.Organization, error) {
	o := &model.Organization{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Code:        req.Code,
		Contact:     req.Contact,
		Phone:       req.Phone,
		Email:       req.Email,
		Type:        req.Type,
		Description: req.Description,
	}
	if o.Type == "" {
		o.Type = "department"
	}
	if err := svc.store.CreateOrganization(o); err != nil {
		return nil, err
	}
	return o, nil
}

func (svc *OrganizationService) GetOrganization(id string) (*model.Organization, error) {
	return svc.store.GetOrganization(id)
}

func (svc *OrganizationService) ListOrganizations() ([]model.Organization, error) {
	return svc.store.ListOrganizations()
}

func (svc *OrganizationService) UpdateOrganization(id string, req UpdateOrganizationRequest) (*model.Organization, error) {
	org, err := svc.store.GetOrganization(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Code != "" {
		org.Code = req.Code
	}
	if req.Contact != "" {
		org.Contact = req.Contact
	}
	if req.Phone != "" {
		org.Phone = req.Phone
	}
	if req.Email != "" {
		org.Email = req.Email
	}
	if req.Type != "" {
		org.Type = req.Type
	}
	if req.Description != "" {
		org.Description = req.Description
	}
	if err := svc.store.UpdateOrganization(org); err != nil {
		return nil, err
	}
	return org, nil
}

func (svc *OrganizationService) DeleteOrganization(id string) error {
	return svc.store.DeleteOrganization(id)
}

type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Contact     string `json:"contact"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type UpdateOrganizationRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Contact     string `json:"contact"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type LicenseService struct {
	store *store.Store
}

func NewLicenseService(s *store.Store) *LicenseService {
	return &LicenseService{store: s}
}

func (svc *LicenseService) GetLicense() (*model.License, error) {
	lic, err := svc.store.GetLicense()
	if err != nil {
		return nil, err
	}
	return lic, nil
}

func (svc *LicenseService) ActivateLicense(req ActivateLicenseRequest) (*model.License, error) {
	lic, err := svc.store.GetLicense()
	if err != nil {
		lic = &model.License{ID: "license-default"}
	}
	lic.Key = req.Key
	lic.Product = req.Product
	lic.CustomerName = req.CustomerName
	lic.MaxClusters = req.MaxClusters
	lic.MaxDeployments = req.MaxDeployments
	lic.Features = req.Features
	lic.ExpiresAt = req.ExpiresAt
	lic.IssuedAt = req.IssuedAt
	lic.IsValid = true
	if err := svc.store.UpdateLicense(lic); err != nil {
		return nil, err
	}
	return lic, nil
}

func (svc *LicenseService) CheckQuota(resource string) (int64, int64, error) {
	lic, err := svc.store.GetLicense()
	if err != nil {
		return 0, 0, err
	}
	switch resource {
	case "clusters":
		current, _ := svc.store.CountClusters()
		return current, int64(lic.MaxClusters), nil
	case "deployments":
		current, _ := svc.store.CountDeployments()
		return current, int64(lic.MaxDeployments), nil
	}
	return 0, 0, nil
}

type ActivateLicenseRequest struct {
	Key            string `json:"key" binding:"required"`
	Product        string `json:"product"`
	CustomerName   string `json:"customer_name"`
	MaxClusters    int    `json:"max_clusters"`
	MaxDeployments int    `json:"max_deployments"`
	Features       string `json:"features"`
	ExpiresAt      time.Time `json:"expires_at"`
	IssuedAt       time.Time `json:"issued_at"`
}

type AlertService struct {
	store *store.Store
}

func NewAlertService(s *store.Store) *AlertService {
	return &AlertService{store: s}
}

func (svc *AlertService) CreateAlertRule(req CreateAlertRuleRequest) (*model.AlertRule, error) {
	r := &model.AlertRule{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Type:           req.Type,
		Condition:      req.Condition,
		Severity:       req.Severity,
		Enabled:        req.Enabled,
		NotifyChannels: req.NotifyChannels,
	}
	if r.Severity == "" {
		r.Severity = "warning"
	}
	if err := svc.store.CreateAlertRule(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (svc *AlertService) GetAlertRule(id string) (*model.AlertRule, error) {
	return svc.store.GetAlertRule(id)
}

func (svc *AlertService) ListAlertRules() ([]model.AlertRule, error) {
	return svc.store.ListAlertRules()
}

func (svc *AlertService) UpdateAlertRule(id string, req UpdateAlertRuleRequest) (*model.AlertRule, error) {
	rule, err := svc.store.GetAlertRule(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Type != "" {
		rule.Type = req.Type
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	rule.Enabled = req.Enabled
	if req.NotifyChannels != "" {
		rule.NotifyChannels = req.NotifyChannels
	}
	if err := svc.store.UpdateAlertRule(rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (svc *AlertService) DeleteAlertRule(id string) error {
	return svc.store.DeleteAlertRule(id)
}

func (svc *AlertService) ListAlertRecords(clusterID string, status string, limit int) ([]model.AlertRecord, error) {
	return svc.store.ListAlertRecords(clusterID, status, limit)
}

func (svc *AlertService) AcknowledgeAlertRecord(id string) error {
	record, err := svc.store.GetAlertRecordByID(id)
	if err != nil {
		return fmt.Errorf("alert record not found")
	}
	if record.Status == "firing" {
		record.Status = "resolved"
		now := time.Now()
		record.ResolvedAt = now
		return svc.store.UpdateAlertRecord(record)
	}
	return nil
}

type CreateAlertRuleRequest struct {
	Name           string `json:"name" binding:"required"`
	Type           string `json:"type" binding:"required"`
	Condition      string `json:"condition" binding:"required"`
	Severity       string `json:"severity"`
	Enabled        bool   `json:"enabled"`
	NotifyChannels string `json:"notify_channels"`
}

type UpdateAlertRuleRequest struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Condition      string `json:"condition"`
	Severity       string `json:"severity"`
	Enabled        bool   `json:"enabled"`
	NotifyChannels string `json:"notify_channels"`
}

type ConfigService struct {
	store *store.Store
}

func NewConfigService(s *store.Store) *ConfigService {
	return &ConfigService{store: s}
}

func (svc *ConfigService) CreateConfigTemplate(req CreateConfigTemplateRequest) (*model.ConfigTemplate, error) {
	c := &model.ConfigTemplate{
		ID:            uuid.New().String(),
		Name:          req.Name,
		OrgID:         req.OrgID,
		ApplicationID: req.ApplicationID,
		Values:        req.Values,
		Description:   req.Description,
	}
	if err := svc.store.CreateConfigTemplate(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (svc *ConfigService) GetConfigTemplate(id string) (*model.ConfigTemplate, error) {
	return svc.store.GetConfigTemplate(id)
}

func (svc *ConfigService) ListConfigTemplates(orgID string) ([]model.ConfigTemplate, error) {
	return svc.store.ListConfigTemplates(orgID)
}

func (svc *ConfigService) UpdateConfigTemplate(id string, req UpdateConfigTemplateRequest) (*model.ConfigTemplate, error) {
	tpl, err := svc.store.GetConfigTemplate(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		tpl.Name = req.Name
	}
	if req.Values != "" {
		tpl.Values = req.Values
	}
	if req.Description != "" {
		tpl.Description = req.Description
	}
	if err := svc.store.UpdateConfigTemplate(tpl); err != nil {
		return nil, err
	}
	return tpl, nil
}

func (svc *ConfigService) DeleteConfigTemplate(id string) error {
	return svc.store.DeleteConfigTemplate(id)
}

type CreateConfigTemplateRequest struct {
	Name          string `json:"name" binding:"required"`
	OrgID         string `json:"org_id"`
	ApplicationID string `json:"application_id"`
	Values        string `json:"values" binding:"required"`
	Description   string `json:"description"`
}

type UpdateConfigTemplateRequest struct {
	Name        string `json:"name"`
	Values      string `json:"values"`
	Description string `json:"description"`
}

type AuditService struct {
	store *store.Store
}

func NewAuditService(s *store.Store) *AuditService {
	return &AuditService{store: s}
}

func (svc *AuditService) Log(userID, username, action, resourceType, resourceID, resourceName, detail, ip string) error {
	a := &model.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Detail:       detail,
		IP:           ip,
	}
	return svc.store.CreateAuditLog(a)
}

func (svc *AuditService) ListAuditLogs(resourceType string, username string, action string, limit int) ([]model.AuditLog, error) {
	return svc.store.ListAuditLogs(resourceType, username, action, limit)
}

type UserService struct {
	store *store.Store
}

func NewUserService(s *store.Store) *UserService {
	return &UserService{store: s}
}

func (svc *UserService) CreateUser(req CreateUserRequest) (*model.User, error) {
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Password: hash,
		Role:     req.Role,
	}
	if u.Role == "" {
		u.Role = "viewer"
	}
	if err := svc.store.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (svc *UserService) ListUsers() ([]model.User, error) {
	return svc.store.ListUsers()
}

func (svc *UserService) UpdateUser(id string, req UpdateUserRequest) (*model.User, error) {
	u, err := svc.store.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if req.Role != "" {
		u.Role = req.Role
	}
	if req.Password != "" {
		hash, err := HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		u.Password = hash
	}
	if err := svc.store.UpdateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (svc *UserService) DeleteUser(id string) error {
	return svc.store.DeleteUser(id)
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Role     string `json:"role"`
	Password string `json:"password"`
}

type AuthService struct {
	store *store.Store
}

func NewAuthService(s *store.Store) *AuthService {
	return &AuthService{store: s}
}

func (svc *AuthService) Login(username, password string) (*model.User, error) {
	user, err := svc.store.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !CheckPassword(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func (svc *AuthService) GetUserByID(id string) (*model.User, error) {
	return svc.store.GetUserByID(id)
}

type DashboardService struct {
	store *store.Store
}

func NewDashboardService(s *store.Store) *DashboardService {
	return &DashboardService{store: s}
}

func (svc *DashboardService) GetStats() (*DashboardStats, error) {
	clusters, err := svc.store.ListClusters()
	if err != nil {
		return nil, err
	}
	apps, err := svc.store.ListApplications()
	if err != nil {
		return nil, err
	}
	orgs, err := svc.store.ListOrganizations()
	if err != nil {
		return nil, err
	}

	var activeClusters, unavailableClusters int
	for _, c := range clusters {
		if c.Status == "active" {
			activeClusters++
		} else if c.Status == "unavailable" {
			unavailableClusters++
		}
	}

	totalDeployments, _ := svc.store.CountDeployments()

	alertRecords, _ := svc.store.ListAlertRecords("", "", 5)

	return &DashboardStats{
		TotalClusters:       len(clusters),
		ActiveClusters:      activeClusters,
		UnavailableClusters: unavailableClusters,
		TotalApplications:   len(apps),
		TotalDeployments:    int(totalDeployments),
		TotalOrganizations:  len(orgs),
		RecentAlerts:        len(alertRecords),
	}, nil
}

type DashboardStats struct {
	TotalClusters       int `json:"total_clusters"`
	ActiveClusters      int `json:"active_clusters"`
	UnavailableClusters int `json:"unavailable_clusters"`
	TotalApplications   int `json:"total_applications"`
	TotalDeployments    int `json:"total_deployments"`
	TotalOrganizations  int `json:"total_organizations"`
	RecentAlerts        int `json:"recent_alerts"`
}

func StartHeartbeatCleanup(s *store.Store) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		clusters, err := s.ListClusters()
		if err != nil {
			continue
		}
		cutoff := time.Now().Add(-7 * 24 * time.Hour)
		for _, c := range clusters {
			s.CleanupOldHeartbeats(c.ID, cutoff)
		}
	}
}
