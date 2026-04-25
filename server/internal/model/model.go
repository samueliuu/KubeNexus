package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type StringMap map[string]string

func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(StringMap)
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("unsupported type for StringMap: %T", value)
	}
	return json.Unmarshal(bytes, m)
}

type Cluster struct {
	ID            string         `gorm:"primaryKey;size:64" json:"id"`
	Name          string         `gorm:"uniqueIndex;size:128" json:"name"`
	DisplayName   string         `gorm:"size:256" json:"display_name"`
	Status        string         `gorm:"size:32;default:registered" json:"status"`
	Token         string         `gorm:"size:256" json:"-"`
	Endpoint      string         `gorm:"size:512" json:"endpoint"`
	Version       string         `gorm:"size:64" json:"version"`
	NodeCount     int            `gorm:"default:0" json:"node_count"`
	CPUCapacity   string         `gorm:"size:32" json:"cpu_capacity"`
	MemCapacity   string         `gorm:"size:32" json:"mem_capacity"`
	Labels        StringMap      `gorm:"type:text" json:"labels"`
	Region        string         `gorm:"size:64" json:"region"`
	OrgID         string         `gorm:"index;size:64" json:"org_id"`
	OrgName       string         `gorm:"size:256" json:"org_name"`
	LastHeartbeat time.Time      `json:"last_heartbeat"`
	WsConnected   bool           `gorm:"default:false" json:"ws_connected"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Application struct {
	ID            string         `gorm:"primaryKey;size:64" json:"id"`
	Name          string         `gorm:"size:128" json:"name"`
	DisplayName   string         `gorm:"size:256" json:"display_name"`
	Description   string         `gorm:"size:1024" json:"description"`
	Icon          string         `gorm:"size:512" json:"icon"`
	ChartName     string         `gorm:"size:128" json:"chart_name"`
	ChartRepo     string         `gorm:"size:512" json:"chart_repo"`
	ChartVersion  string         `gorm:"size:64" json:"chart_version"`
	Category      string         `gorm:"size:64" json:"category"`
	IsSaaS        bool           `gorm:"default:1" json:"is_saas"`
	DefaultValues string         `gorm:"type:text" json:"default_values"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Deployment struct {
	ID            string         `gorm:"primaryKey;size:64" json:"id"`
	ClusterID     string         `gorm:"index;size:64" json:"cluster_id"`
	ApplicationID string         `gorm:"index;size:64" json:"application_id"`
	Name          string         `gorm:"size:128" json:"name"`
	Namespace     string         `gorm:"size:64;default:default" json:"namespace"`
	Values        string         `gorm:"type:text" json:"values"`
	Status        string         `gorm:"size:32;default:pending" json:"status"`
	ActualStatus  string         `gorm:"size:32" json:"actual_status"`
	Replicas      int            `gorm:"default:1" json:"replicas"`
	Version       string         `gorm:"size:64" json:"version"`
	ActualVersion string         `gorm:"size:64" json:"actual_version"`
	DriftDetail   string         `gorm:"type:text" json:"drift_detail"`
	Message       string         `gorm:"size:1024" json:"message"`
	LastSynced    time.Time      `json:"last_synced"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Organization struct {
	ID          string         `gorm:"primaryKey;size:64" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:256" json:"name"`
	Code        string         `gorm:"uniqueIndex;size:64" json:"code"`
	Contact     string         `gorm:"size:128" json:"contact"`
	Phone       string         `gorm:"size:32" json:"phone"`
	Email       string         `gorm:"size:128" json:"email"`
	Type        string         `gorm:"size:32;default:department" json:"type"`
	Description string         `gorm:"size:1024" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type License struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	Key            string    `gorm:"size:1024" json:"-"`
	Product        string    `gorm:"size:128" json:"product"`
	CustomerName   string    `gorm:"size:256" json:"customer_name"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	MaxClusters    int       `gorm:"default:5" json:"max_clusters"`
	MaxDeployments int       `gorm:"default:50" json:"max_deployments"`
	Features       string    `gorm:"type:text" json:"features"`
	IsValid        bool      `gorm:"default:1" json:"is_valid"`
	CreatedAt      time.Time `json:"created_at"`
}

type AlertRule struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	Name           string    `gorm:"size:128" json:"name"`
	Type           string    `gorm:"size:64" json:"type"`
	Condition      string    `gorm:"type:text" json:"condition"`
	Severity       string    `gorm:"size:32;default:warning" json:"severity"`
	Enabled        bool      `gorm:"default:1" json:"enabled"`
	NotifyChannels string    `gorm:"type:text" json:"notify_channels"`
	LastTriggered  time.Time `json:"last_triggered"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AlertRecord struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	RuleID      string    `gorm:"index;size:64" json:"rule_id"`
	RuleName    string    `gorm:"size:128" json:"rule_name"`
	ClusterID   string    `gorm:"index;size:64" json:"cluster_id"`
	Severity    string    `gorm:"size:32" json:"severity"`
	Message     string    `gorm:"size:1024" json:"message"`
	Status      string    `gorm:"size:32;default:firing" json:"status"`
	TriggeredAt time.Time `json:"triggered_at"`
	ResolvedAt  time.Time `json:"resolved_at"`
}

type ConfigTemplate struct {
	ID            string    `gorm:"primaryKey;size:64" json:"id"`
	Name          string    `gorm:"size:128" json:"name"`
	OrgID         string    `gorm:"index;size:64" json:"org_id"`
	ApplicationID string    `gorm:"index;size:64" json:"application_id"`
	Values        string    `gorm:"type:text" json:"values"`
	Description   string    `gorm:"size:1024" json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID           string    `gorm:"primaryKey;size:64" json:"id"`
	UserID       string    `gorm:"index;size:64" json:"user_id"`
	Username     string    `gorm:"size:128" json:"username"`
	Action       string    `gorm:"size:64" json:"action"`
	ResourceType string    `gorm:"size:64" json:"resource_type"`
	ResourceID   string    `gorm:"size:64" json:"resource_id"`
	ResourceName string    `gorm:"size:256" json:"resource_name"`
	Detail       string    `gorm:"type:text" json:"detail"`
	IP           string    `gorm:"size:64" json:"ip"`
	CreatedAt    time.Time `json:"created_at"`
}

type Heartbeat struct {
	ID         string    `gorm:"primaryKey;size:64" json:"id"`
	ClusterID  string    `gorm:"index;size:64" json:"cluster_id"`
	NodeCount  int       `json:"node_count"`
	CPUUsage   float64   `json:"cpu_usage"`
	MemUsage   float64   `json:"mem_usage"`
	PodCount   int       `json:"pod_count"`
	Version    string    `gorm:"size:64" json:"version"`
	Info       string    `gorm:"type:text" json:"info"`
	ReportedAt time.Time `json:"reported_at"`
}

type User struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:128" json:"username"`
	Password  string         `gorm:"size:256" json:"-"`
	Role      string         `gorm:"size:32;default:viewer" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
