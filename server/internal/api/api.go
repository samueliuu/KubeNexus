package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kubenexus/server/internal/middleware"
	"github.com/kubenexus/server/internal/service"
	"github.com/kubenexus/server/internal/store"
	"github.com/kubenexus/server/internal/tunnel"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	ClusterSvc     *service.ClusterService
	ApplicationSvc *service.ApplicationService
	DeploymentSvc  *service.DeploymentService
	OrganizationSvc *service.OrganizationService
	LicenseSvc     *service.LicenseService
	AlertSvc       *service.AlertService
	ConfigSvc      *service.ConfigService
	AuditSvc       *service.AuditService
	UserSvc        *service.UserService
	AuthSvc        *service.AuthService
	DashboardSvc   *service.DashboardService
	TunnelMgr      *tunnel.Manager
	ServerURL      string
	store          *store.Store
}

func NewHandler(
	clusterSvc *service.ClusterService,
	appSvc *service.ApplicationService,
	deploySvc *service.DeploymentService,
	orgSvc *service.OrganizationService,
	licenseSvc *service.LicenseService,
	alertSvc *service.AlertService,
	configSvc *service.ConfigService,
	auditSvc *service.AuditService,
	userSvc *service.UserService,
	authSvc *service.AuthService,
	dashboardSvc *service.DashboardService,
	tunnelMgr *tunnel.Manager,
	serverURL string,
	s *store.Store,
) *Handler {
	return &Handler{
		ClusterSvc:     clusterSvc,
		ApplicationSvc: appSvc,
		DeploymentSvc:  deploySvc,
		OrganizationSvc: orgSvc,
		LicenseSvc:     licenseSvc,
		AlertSvc:       alertSvc,
		ConfigSvc:      configSvc,
		AuditSvc:       auditSvc,
		UserSvc:        userSvc,
		AuthSvc:        authSvc,
		DashboardSvc:   dashboardSvc,
		TunnelMgr:      tunnelMgr,
		ServerURL:      serverURL,
		store:          s,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.Login)
			auth.POST("/token", middleware.AuthMiddleware(), h.RefreshToken)
			auth.GET("/me", middleware.AuthMiddleware(), h.GetCurrentUser)
		}

		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware())
		{
			dashboard.GET("/stats", h.GetDashboardStats)
		}

		clusters := api.Group("/clusters")
		{
			clusters.POST("", middleware.AuthMiddleware(), h.RegisterCluster)
			clusters.GET("", middleware.AuthMiddleware(), h.ListClusters)
			clusters.GET("/:id", middleware.AuthMiddleware(), h.GetCluster)
			clusters.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), h.DeleteCluster)
			clusters.PUT("/:id/labels", middleware.AuthMiddleware(), h.UpdateClusterLabels)
			clusters.POST("/:id/token/rotate", middleware.AuthMiddleware(), middleware.AdminMiddleware(), h.RotateClusterToken)
			clusters.GET("/:id/install-script", middleware.AuthMiddleware(), h.GetInstallScript)
			clusters.GET("/:id/registration.yaml", middleware.AuthMiddleware(), h.GetRegistrationYAML)
			clusters.POST("/:id/heartbeat", middleware.AgentAuthMiddleware(h.store), h.Heartbeat)
			clusters.GET("/:id/desired-state", middleware.AgentAuthMiddleware(h.store), h.GetDesiredState)
			clusters.POST("/:id/sync-result", middleware.AgentAuthMiddleware(h.store), h.SyncResult)
			clusters.GET("/:id/tunnel", middleware.AgentAuthMiddleware(h.store), h.HandleTunnel)
			clusters.GET("/:id/metrics", middleware.AuthMiddleware(), h.GetClusterMetrics)
			clusters.POST("/:id/proxy", middleware.AuthMiddleware(), h.ProxyK8sAPI)
		}

		applications := api.Group("/applications")
		applications.Use(middleware.AuthMiddleware())
		{
			applications.POST("", middleware.AdminMiddleware(), h.CreateApplication)
			applications.GET("", h.ListApplications)
			applications.GET("/:id", h.GetApplication)
			applications.PUT("/:id", middleware.AdminMiddleware(), h.UpdateApplication)
			applications.DELETE("/:id", middleware.AdminMiddleware(), h.DeleteApplication)
		}

		deployments := api.Group("/deployments")
		deployments.Use(middleware.AuthMiddleware())
		{
			deployments.POST("", h.Deploy)
			deployments.POST("/batch", middleware.AdminMiddleware(), h.BatchDeploy)
			deployments.GET("", h.ListDeployments)
			deployments.GET("/:id", h.GetDeployment)
			deployments.PUT("/:id", h.UpdateDeployment)
			deployments.DELETE("/:id", h.DeleteDeployment)
		}

		organizations := api.Group("/organizations")
		organizations.Use(middleware.AuthMiddleware())
		{
			organizations.POST("", middleware.AdminMiddleware(), h.CreateOrganization)
			organizations.GET("", h.ListOrganizations)
			organizations.GET("/:id", h.GetOrganization)
			organizations.PUT("/:id", middleware.AdminMiddleware(), h.UpdateOrganization)
			organizations.DELETE("/:id", middleware.AdminMiddleware(), h.DeleteOrganization)
		}

		license := api.Group("/license")
		license.Use(middleware.AuthMiddleware())
		{
			license.GET("", h.GetLicense)
			license.POST("/activate", middleware.AdminMiddleware(), h.ActivateLicense)
			license.GET("/quota", h.GetLicenseQuota)
		}

		alerts := api.Group("/alerts")
		alerts.Use(middleware.AuthMiddleware())
		{
			alerts.GET("/rules", h.ListAlertRules)
			alerts.POST("/rules", middleware.AdminMiddleware(), h.CreateAlertRule)
			alerts.PUT("/rules/:id", middleware.AdminMiddleware(), h.UpdateAlertRule)
			alerts.DELETE("/rules/:id", middleware.AdminMiddleware(), h.DeleteAlertRule)
			alerts.GET("/records", h.ListAlertRecords)
		}

		configs := api.Group("/configs")
		configs.Use(middleware.AuthMiddleware())
		{
			configs.GET("", h.ListConfigTemplates)
			configs.POST("", h.CreateConfigTemplate)
			configs.PUT("/:id", h.UpdateConfigTemplate)
			configs.DELETE("/:id", h.DeleteConfigTemplate)
		}

		auditLogs := api.Group("/audit-logs")
		auditLogs.Use(middleware.AuthMiddleware())
		{
			auditLogs.GET("", h.ListAuditLogs)
		}

		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			users.POST("", h.CreateUser)
			users.GET("", h.ListUsers)
			users.PUT("/:id", h.UpdateUser)
			users.DELETE("/:id", h.DeleteUser)
		}
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.AuthSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	h.AuditSvc.Log(user.ID, user.Username, "login", "user", user.ID, user.Username, "", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{"id": user.ID, "username": user.Username, "role": user.Role},
	})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	role := c.GetString("role")
	token, err := middleware.GenerateToken(userID, username, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.AuthSvc.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "username": user.Username, "role": user.Role})
}

func (h *Handler) GetDashboardStats(c *gin.Context) {
	stats, err := h.DashboardSvc.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) RegisterCluster(c *gin.Context) {
	var req service.RegisterClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cluster, err := h.ClusterSvc.RegisterCluster(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "create", "cluster", cluster.ID, cluster.Name)
	c.JSON(http.StatusCreated, cluster)
}

func (h *Handler) GetCluster(c *gin.Context) {
	id := c.Param("id")
	cluster, err := h.ClusterSvc.GetCluster(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}
	c.JSON(http.StatusOK, cluster)
}

func (h *Handler) ListClusters(c *gin.Context) {
	clusters, err := h.ClusterSvc.ListClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": clusters})
}

func (h *Handler) DeleteCluster(c *gin.Context) {
	id := c.Param("id")
	h.auditLog(c, "delete", "cluster", id, "")
	if err := h.ClusterSvc.DeleteCluster(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) UpdateClusterLabels(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Labels map[string]string `json:"labels" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cluster, err := h.ClusterSvc.UpdateLabels(id, req.Labels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "update", "cluster", id, cluster.Name)
	c.JSON(http.StatusOK, cluster)
}

func (h *Handler) RotateClusterToken(c *gin.Context) {
	id := c.Param("id")
	cluster, err := h.ClusterSvc.RotateToken(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "rotate_token", "cluster", id, cluster.Name)
	c.JSON(http.StatusOK, gin.H{"message": "token rotated", "token": cluster.Token})
}

func (h *Handler) GetInstallScript(c *gin.Context) {
	id := c.Param("id")
	script, err := h.ClusterSvc.GetInstallScript(id, h.ServerURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}
	c.Header("Content-Type", "text/x-shellscript")
	c.String(http.StatusOK, script)
}

func (h *Handler) GetRegistrationYAML(c *gin.Context) {
	id := c.Param("id")
	yaml, err := h.ClusterSvc.GetRegistrationYAML(id, h.ServerURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}
	c.Header("Content-Type", "application/yaml")
	c.String(http.StatusOK, yaml)
}

func (h *Handler) Heartbeat(c *gin.Context) {
	id := c.Param("id")
	var req service.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.ClusterSvc.HandleHeartbeat(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) GetDesiredState(c *gin.Context) {
	id := c.Param("id")
	state, err := h.ClusterSvc.GetDesiredState(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *Handler) SyncResult(c *gin.Context) {
	id := c.Param("id")
	var req service.SyncResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.ClusterSvc.HandleSyncResult(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) HandleTunnel(c *gin.Context) {
	h.TunnelMgr.HandleWebSocket(c)
}

func (h *Handler) GetClusterMetrics(c *gin.Context) {
	id := c.Param("id")
	metrics, err := h.ClusterSvc.GetClusterMetrics(id, 60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": metrics})
}

func (h *Handler) ProxyK8sAPI(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Method  string            `json:"method" binding:"required"`
		Path    string            `json:"path" binding:"required"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	blockedPrefixes := []string{"/api/v1/secrets", "/api/v1/configmaps"}
	for _, prefix := range blockedPrefixes {
		if strings.HasPrefix(req.Path, prefix) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access to this resource type is restricted"})
			return
		}
	}
	resp, err := h.TunnelMgr.ProxyRequest(id, &tunnel.TunnelPayload{
		Method:  req.Method,
		Path:    req.Path,
		Headers: req.Headers,
		Body:    req.Body,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateApplication(c *gin.Context) {
	var req service.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := h.ApplicationSvc.CreateApplication(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "create", "application", app.ID, app.Name)
	c.JSON(http.StatusCreated, app)
}

func (h *Handler) GetApplication(c *gin.Context) {
	id := c.Param("id")
	app, err := h.ApplicationSvc.GetApplication(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) ListApplications(c *gin.Context) {
	apps, err := h.ApplicationSvc.ListApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": apps})
}

func (h *Handler) UpdateApplication(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := h.ApplicationSvc.UpdateApplication(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "update", "application", id, app.Name)
	c.JSON(http.StatusOK, app)
}

func (h *Handler) DeleteApplication(c *gin.Context) {
	id := c.Param("id")
	h.auditLog(c, "delete", "application", id, "")
	if err := h.ApplicationSvc.DeleteApplication(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) Deploy(c *gin.Context) {
	var req service.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dep, err := h.DeploymentSvc.Deploy(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "deploy", "deployment", dep.ID, dep.Name)
	c.JSON(http.StatusCreated, dep)
}

func (h *Handler) BatchDeploy(c *gin.Context) {
	var req service.BatchDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	deps, err := h.DeploymentSvc.BatchDeploy(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "batch_deploy", "deployment", "", req.Name)
	c.JSON(http.StatusCreated, gin.H{"items": deps})
}

func (h *Handler) GetDeployment(c *gin.Context) {
	id := c.Param("id")
	dep, err := h.DeploymentSvc.GetDeployment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}
	c.JSON(http.StatusOK, dep)
}

func (h *Handler) ListDeployments(c *gin.Context) {
	clusterID := c.Query("cluster_id")
	deps, err := h.DeploymentSvc.ListDeployments(clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": deps})
}

func (h *Handler) UpdateDeployment(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dep, err := h.DeploymentSvc.UpdateDeployment(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "update", "deployment", id, dep.Name)
	c.JSON(http.StatusOK, dep)
}

func (h *Handler) DeleteDeployment(c *gin.Context) {
	id := c.Param("id")
	h.auditLog(c, "delete", "deployment", id, "")
	if err := h.DeploymentSvc.DeleteDeployment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) CreateOrganization(c *gin.Context) {
	var req service.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org, err := h.OrganizationSvc.CreateOrganization(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "create", "organization", org.ID, org.Name)
	c.JSON(http.StatusCreated, org)
}

func (h *Handler) GetOrganization(c *gin.Context) {
	id := c.Param("id")
	org, err := h.OrganizationSvc.GetOrganization(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	c.JSON(http.StatusOK, org)
}

func (h *Handler) ListOrganizations(c *gin.Context) {
	orgs, err := h.OrganizationSvc.ListOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": orgs})
}

func (h *Handler) UpdateOrganization(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org, err := h.OrganizationSvc.UpdateOrganization(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "update", "organization", id, org.Name)
	c.JSON(http.StatusOK, org)
}

func (h *Handler) DeleteOrganization(c *gin.Context) {
	id := c.Param("id")
	h.auditLog(c, "delete", "organization", id, "")
	if err := h.OrganizationSvc.DeleteOrganization(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) GetLicense(c *gin.Context) {
	lic, err := h.LicenseSvc.GetLicense()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "license not found"})
		return
	}
	c.JSON(http.StatusOK, lic)
}

func (h *Handler) ActivateLicense(c *gin.Context) {
	var req service.ActivateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lic, err := h.LicenseSvc.ActivateLicense(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "activate", "license", lic.ID, "")
	c.JSON(http.StatusOK, lic)
}

func (h *Handler) GetLicenseQuota(c *gin.Context) {
	clusterCurrent, clusterMax, _ := h.LicenseSvc.CheckQuota("clusters")
	deployCurrent, deployMax, _ := h.LicenseSvc.CheckQuota("deployments")
	c.JSON(http.StatusOK, gin.H{
		"clusters":   gin.H{"current": clusterCurrent, "max": clusterMax},
		"deployments": gin.H{"current": deployCurrent, "max": deployMax},
	})
}

func (h *Handler) ListAlertRules(c *gin.Context) {
	rules, err := h.AlertSvc.ListAlertRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rules})
}

func (h *Handler) CreateAlertRule(c *gin.Context) {
	var req service.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.AlertSvc.CreateAlertRule(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "create", "alert_rule", rule.ID, rule.Name)
	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) UpdateAlertRule(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.AlertSvc.UpdateAlertRule(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *Handler) DeleteAlertRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.AlertSvc.DeleteAlertRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) ListAlertRecords(c *gin.Context) {
	clusterID := c.Query("cluster_id")
	records, err := h.AlertSvc.ListAlertRecords(clusterID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": records})
}

func (h *Handler) ListConfigTemplates(c *gin.Context) {
	orgID := c.Query("org_id")
	configs, err := h.ConfigSvc.ListConfigTemplates(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": configs})
}

func (h *Handler) CreateConfigTemplate(c *gin.Context) {
	var req service.CreateConfigTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tpl, err := h.ConfigSvc.CreateConfigTemplate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "create", "config_template", tpl.ID, tpl.Name)
	c.JSON(http.StatusCreated, tpl)
}

func (h *Handler) UpdateConfigTemplate(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateConfigTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tpl, err := h.ConfigSvc.UpdateConfigTemplate(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

func (h *Handler) DeleteConfigTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := h.ConfigSvc.DeleteConfigTemplate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	resourceType := c.Query("resource_type")
	logs, err := h.AuditSvc.ListAuditLogs(resourceType, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": logs})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.UserSvc.CreateUser(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "create", "user", user.ID, user.Username)
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.UserSvc.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.UserSvc.UpdateUser(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.auditLog(c, "update", "user", id, user.Username)
	c.JSON(http.StatusOK, user)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	h.auditLog(c, "delete", "user", id, "")
	if err := h.UserSvc.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) auditLog(c *gin.Context, action, resourceType, resourceID, resourceName string) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	clientIP := c.ClientIP()
	go h.AuditSvc.Log(userID, username, action, resourceType, resourceID, resourceName, "", clientIP)
}

func StartClusterStatusChecker(svc *service.ClusterService) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			clusters, err := svc.ListClusters()
			if err != nil {
				continue
			}
			for i := range clusters {
				if clusters[i].Status == "active" && time.Since(clusters[i].LastHeartbeat) > 90*time.Second {
					clusters[i].Status = "unavailable"
					if err := svc.UpdateClusterStatus(&clusters[i]); err != nil {
						log.Printf("Failed to update cluster %s status: %v", clusters[i].ID, err)
					}
				}
			}
		}
	}()
}
