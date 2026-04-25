package store

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/kubenexus/server/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var labelKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-./]+$`)

type Store struct {
	DB *gorm.DB
}

func New(dsn string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Cluster{},
		&model.Application{},
		&model.Deployment{},
		&model.Organization{},
		&model.License{},
		&model.AlertRule{},
		&model.AlertRecord{},
		&model.ConfigTemplate{},
		&model.AuditLog{},
		&model.Heartbeat{},
		&model.User{},
	); err != nil {
		return nil, err
	}

	s := &Store{DB: db}
	s.ensureDefaultAdmin()
	s.ensureDefaultLicense()
	s.ensureDefaultAlertRules()
	return s, nil
}

func (s *Store) ensureDefaultAdmin() {
	var count int64
	s.DB.Model(&model.User{}).Count(&count)
	if count == 0 {
		rawPassword := uuid.New().String()[:12]
		hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("failed to hash default admin password: %v", err)
		}
		s.DB.Create(&model.User{
			ID:       "user-admin",
			Username: "admin",
			Password: string(hashed),
			Role:     "admin",
		})
		log.Printf("========================================")
		log.Printf("默认管理员账号已创建")
		log.Printf("用户名: admin")
		log.Printf("密码: %s", rawPassword)
		log.Printf("请立即登录并修改密码！")
		log.Printf("========================================")
	}
}

func (s *Store) ensureDefaultLicense() {
	var count int64
	s.DB.Model(&model.License{}).Count(&count)
	if count == 0 {
		s.DB.Create(&model.License{
			ID:             "license-default",
			Product:        "kubenexus",
			CustomerName:   "默认",
			MaxClusters:    10,
			MaxDeployments: 100,
			Features:       `{"monitoring":true,"tunnel":true,"config_center":true}`,
			IsValid:        true,
			ExpiresAt:      time.Now().Add(365 * 24 * time.Hour),
			IssuedAt:       time.Now(),
		})
	}
}

func (s *Store) ensureDefaultAlertRules() {
	var count int64
	s.DB.Model(&model.AlertRule{}).Count(&count)
	if count == 0 {
		rules := []model.AlertRule{
			{ID: "rule-cluster-down", Name: "集群离线", Type: "cluster_down", Condition: `{"metric":"status","operator":"==","threshold":"unavailable","duration":"90s"}`, Severity: "critical", Enabled: true},
			{ID: "rule-cpu-high", Name: "CPU使用率过高", Type: "cpu_high", Condition: `{"metric":"cpu_usage","operator":">","threshold":80,"duration":"5m"}`, Severity: "warning", Enabled: true},
			{ID: "rule-mem-high", Name: "内存使用率过高", Type: "mem_high", Condition: `{"metric":"mem_usage","operator":">","threshold":80,"duration":"5m"}`, Severity: "warning", Enabled: true},
			{ID: "rule-drift", Name: "配置漂移检测", Type: "drift_detected", Condition: `{"metric":"drift","operator":"==","threshold":"true"}`, Severity: "warning", Enabled: true},
			{ID: "rule-license", Name: "License即将过期", Type: "license_expiring", Condition: `{"metric":"license_days","operator":"<","threshold":30}`, Severity: "warning", Enabled: true},
		}
		for _, r := range rules {
			s.DB.Create(&r)
		}
	}
}

func (s *Store) CreateCluster(c *model.Cluster) error {
	return s.DB.Create(c).Error
}

func (s *Store) GetCluster(id string) (*model.Cluster, error) {
	var c model.Cluster
	if err := s.DB.First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) GetClusterByToken(token string) (*model.Cluster, error) {
	var c model.Cluster
	if err := s.DB.First(&c, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListClusters() ([]model.Cluster, error) {
	var clusters []model.Cluster
	if err := s.DB.Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}

func (s *Store) ListClustersByLabels(labels map[string]string) ([]model.Cluster, error) {
	var clusters []model.Cluster
	q := s.DB
	for k, v := range labels {
		if !labelKeyPattern.MatchString(k) {
			return nil, fmt.Errorf("invalid label key: %s", k)
		}
		jsonPath := "$.\"" + k + "\""
		q = q.Where("json_extract(labels, ?) = ?", jsonPath, v)
	}
	if err := q.Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}

func (s *Store) UpdateCluster(c *model.Cluster) error {
	return s.DB.Save(c).Error
}

func (s *Store) DeleteCluster(id string) error {
	return s.DB.Delete(&model.Cluster{}, "id = ?", id).Error
}

func (s *Store) CountClusters() (int64, error) {
	var count int64
	err := s.DB.Model(&model.Cluster{}).Count(&count).Error
	return count, err
}

func (s *Store) CreateApplication(a *model.Application) error {
	return s.DB.Create(a).Error
}

func (s *Store) GetApplication(id string) (*model.Application, error) {
	var a model.Application
	if err := s.DB.First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) ListApplications() ([]model.Application, error) {
	var apps []model.Application
	if err := s.DB.Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (s *Store) UpdateApplication(a *model.Application) error {
	return s.DB.Save(a).Error
}

func (s *Store) DeleteApplication(id string) error {
	return s.DB.Delete(&model.Application{}, "id = ?", id).Error
}

func (s *Store) CreateDeployment(d *model.Deployment) error {
	return s.DB.Create(d).Error
}

func (s *Store) GetDeployment(id string) (*model.Deployment, error) {
	var d model.Deployment
	if err := s.DB.First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) ListDeployments(clusterID string) ([]model.Deployment, error) {
	var deps []model.Deployment
	q := s.DB
	if clusterID != "" {
		q = q.Where("cluster_id = ?", clusterID)
	}
	if err := q.Find(&deps).Error; err != nil {
		return nil, err
	}
	return deps, nil
}

func (s *Store) ListDeploymentsByCluster(clusterID string) ([]model.Deployment, error) {
	var deps []model.Deployment
	if err := s.DB.Where("cluster_id = ? AND status != 'stopped'", clusterID).Find(&deps).Error; err != nil {
		return nil, err
	}
	return deps, nil
}

func (s *Store) UpdateDeployment(d *model.Deployment) error {
	return s.DB.Save(d).Error
}

func (s *Store) DeleteDeployment(id string) error {
	return s.DB.Delete(&model.Deployment{}, "id = ?", id).Error
}

func (s *Store) CountDeployments() (int64, error) {
	var count int64
	err := s.DB.Model(&model.Deployment{}).Where("status != 'stopped'").Count(&count).Error
	return count, err
}

func (s *Store) CreateOrganization(o *model.Organization) error {
	return s.DB.Create(o).Error
}

func (s *Store) GetOrganization(id string) (*model.Organization, error) {
	var o model.Organization
	if err := s.DB.First(&o, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) ListOrganizations() ([]model.Organization, error) {
	var orgs []model.Organization
	if err := s.DB.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

func (s *Store) UpdateOrganization(o *model.Organization) error {
	return s.DB.Save(o).Error
}

func (s *Store) DeleteOrganization(id string) error {
	return s.DB.Delete(&model.Organization{}, "id = ?", id).Error
}

func (s *Store) GetLicense() (*model.License, error) {
	var l model.License
	if err := s.DB.First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *Store) UpdateLicense(l *model.License) error {
	return s.DB.Save(l).Error
}

func (s *Store) CreateAlertRule(r *model.AlertRule) error {
	return s.DB.Create(r).Error
}

func (s *Store) GetAlertRule(id string) (*model.AlertRule, error) {
	var r model.AlertRule
	if err := s.DB.First(&r, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) ListAlertRules() ([]model.AlertRule, error) {
	var rules []model.AlertRule
	if err := s.DB.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *Store) UpdateAlertRule(r *model.AlertRule) error {
	return s.DB.Save(r).Error
}

func (s *Store) DeleteAlertRule(id string) error {
	return s.DB.Delete(&model.AlertRule{}, "id = ?", id).Error
}

func (s *Store) CreateAlertRecord(r *model.AlertRecord) error {
	return s.DB.Create(r).Error
}

func (s *Store) ListAlertRecords(clusterID string, status string, limit int) ([]model.AlertRecord, error) {
	var records []model.AlertRecord
	q := s.DB.Order("triggered_at DESC")
	if clusterID != "" {
		q = q.Where("cluster_id = ?", clusterID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (s *Store) UpdateAlertRecord(r *model.AlertRecord) error {
	return s.DB.Save(r).Error
}

func (s *Store) GetFiringAlertRecord(ruleID string, clusterID string) (*model.AlertRecord, error) {
	var r model.AlertRecord
	if err := s.DB.Where("rule_id = ? AND cluster_id = ? AND status = ?", ruleID, clusterID, "firing").First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) ListFiringAlertRecords(ruleID string, clusterID string) ([]model.AlertRecord, error) {
	var records []model.AlertRecord
	q := s.DB.Where("rule_id = ? AND cluster_id = ? AND status = ?", ruleID, clusterID, "firing")
	if err := q.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (s *Store) GetAlertRecordByID(id string) (*model.AlertRecord, error) {
	var r model.AlertRecord
	if err := s.DB.First(&r, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) CreateConfigTemplate(c *model.ConfigTemplate) error {
	return s.DB.Create(c).Error
}

func (s *Store) GetConfigTemplate(id string) (*model.ConfigTemplate, error) {
	var c model.ConfigTemplate
	if err := s.DB.First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListConfigTemplates(orgID string) ([]model.ConfigTemplate, error) {
	var configs []model.ConfigTemplate
	q := s.DB
	if orgID != "" {
		q = q.Where("org_id = ? OR org_id = ''", orgID)
	}
	if err := q.Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *Store) UpdateConfigTemplate(c *model.ConfigTemplate) error {
	return s.DB.Save(c).Error
}

func (s *Store) DeleteConfigTemplate(id string) error {
	return s.DB.Delete(&model.ConfigTemplate{}, "id = ?", id).Error
}

func (s *Store) CreateAuditLog(a *model.AuditLog) error {
	return s.DB.Create(a).Error
}

func (s *Store) ListAuditLogs(resourceType string, username string, action string, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	q := s.DB.Order("created_at DESC")
	if resourceType != "" {
		q = q.Where("resource_type = ?", resourceType)
	}
	if username != "" {
		q = q.Where("username = ?", username)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *Store) CreateHeartbeat(h *model.Heartbeat) error {
	return s.DB.Create(h).Error
}

func (s *Store) GetLatestHeartbeat(clusterID string) (*model.Heartbeat, error) {
	var h model.Heartbeat
	if err := s.DB.Where("cluster_id = ?", clusterID).Order("reported_at DESC").First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *Store) ListHeartbeats(clusterID string, limit int) ([]model.Heartbeat, error) {
	var hbs []model.Heartbeat
	q := s.DB.Where("cluster_id = ?", clusterID).Order("reported_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&hbs).Error; err != nil {
		return nil, err
	}
	return hbs, nil
}

func (s *Store) CleanupOldHeartbeats(clusterID string, before time.Time) error {
	return s.DB.Where("cluster_id = ? AND reported_at < ?", clusterID, before).Delete(&model.Heartbeat{}).Error
}

func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	if err := s.DB.First(&u, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByID(id string) (*model.User, error) {
	var u model.User
	if err := s.DB.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) CreateUser(u *model.User) error {
	return s.DB.Create(u).Error
}

func (s *Store) ListUsers() ([]model.User, error) {
	var users []model.User
	if err := s.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) UpdateUser(u *model.User) error {
	return s.DB.Save(u).Error
}

func (s *Store) DeleteUser(id string) error {
	return s.DB.Delete(&model.User{}, "id = ?", id).Error
}
