package main

import (
	"log"
	"os"

	"github.com/kubenexus/server/internal/api"
	"github.com/kubenexus/server/internal/middleware"
	"github.com/kubenexus/server/internal/service"
	"github.com/kubenexus/server/internal/store"
	"github.com/kubenexus/server/internal/tunnel"
	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "kubenexus.db"
	}
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	s, err := store.New(dsn)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	middleware.InitAuth(s)

	clusterSvc := service.NewClusterService(s)
	appSvc := service.NewApplicationService(s)
	deploySvc := service.NewDeploymentService(s)
	orgSvc := service.NewOrganizationService(s)
	licenseSvc := service.NewLicenseService(s)
	alertSvc := service.NewAlertService(s)
	configSvc := service.NewConfigService(s)
	auditSvc := service.NewAuditService(s)
	userSvc := service.NewUserService(s)
	authSvc := service.NewAuthService(s)
	dashboardSvc := service.NewDashboardService(s)

	tunnelMgr := tunnel.NewManager(
		func(clusterID string) {
			cluster, err := clusterSvc.GetCluster(clusterID)
			if err == nil {
				cluster.WsConnected = true
				clusterSvc.UpdateClusterStatus(cluster)
			}
		},
		func(clusterID string) {
			cluster, err := clusterSvc.GetCluster(clusterID)
			if err == nil {
				cluster.WsConnected = false
				clusterSvc.UpdateClusterStatus(cluster)
			}
		},
	)

	handler := api.NewHandler(
		clusterSvc, appSvc, deploySvc, orgSvc,
		licenseSvc, alertSvc, configSvc, auditSvc,
		userSvc, authSvc, dashboardSvc,
		tunnelMgr, serverURL, s,
	)

	r := gin.Default()
	r.Use(corsMiddleware())
	handler.RegisterRoutes(r)
	r.Static("/assets", "./web/dist/assets")
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	api.StartClusterStatusChecker(clusterSvc)

	alertEngine := service.NewAlertEngine(s)
	alertEngine.Start()

	go service.StartHeartbeatCleanup(s)

	log.Printf("衡牧KubeNexusK3s多集群管理系统 Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigins := []string{"http://localhost:3000", "http://localhost:3001"}
		allowOrigin := ""
		for _, o := range allowedOrigins {
			if origin == o {
				allowOrigin = o
				break
			}
		}
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Cluster-Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
