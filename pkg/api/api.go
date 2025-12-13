package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/konpure/Kon-Agent-export/pkg/storage"
)

// APIServer HTTP API服务器
type APIServer struct {
	storage storage.Storage
	server  *http.Server
}

// NewAPIServer 创建API服务器实例
func NewAPIServer(storage storage.Storage) *APIServer {
	return &APIServer{
		storage: storage,
	}
}

// Start 启动API服务器
func (s *APIServer) Start(addr string, readTimeout, writeTimeout time.Duration) error {
	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 定义API路由
	api := r.Group("/api/v1")
	{
		api.GET("/metrics", s.getAllMetrics)
		api.GET("/metrics/:agent_id", s.getMetricsByAgentID)
		api.GET("/metrics/type/:metric_type", s.getMetricsByType)
		api.GET("/metrics/latest", s.getLatestMetrics)
		api.GET("/metrics/range", s.getMetricsByTimeRange)
	}

	// 定义HTTP服务器
	s.server = &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	log.Printf("HTTP API server starting on %s", addr)
	return s.server.ListenAndServe()
}

// getAllMetrics 获取所有监控数据
func (s *APIServer) getAllMetrics(c *gin.Context) {
	// 获取查询参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// 调用存储层获取最新数据
	metrics, err := s.storage.GetLatestMetrics(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// getMetricsByAgentID 按Agent ID获取监控数据
func (s *APIServer) getMetricsByAgentID(c *gin.Context) {
	// 获取路径参数
	agentID := c.Param("agent_id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id is required"})
		return
	}

	// 获取查询参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// 调用存储层获取数据
	metrics, err := s.storage.GetMetricsByAgentID(agentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// getMetricsByType 按指标类型获取监控数据
func (s *APIServer) getMetricsByType(c *gin.Context) {
	// 获取路径参数
	metricType := c.Param("metric_type")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric_type is required"})
		return
	}

	// 获取查询参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// 调用存储层获取数据
	metrics, err := s.storage.GetMetricsByType(metricType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// getLatestMetrics 获取最新监控数据
func (s *APIServer) getLatestMetrics(c *gin.Context) {
	// 获取查询参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 调用存储层获取最新数据
	metrics, err := s.storage.GetLatestMetrics(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// getMetricsByTimeRange 按时间范围获取监控数据
func (s *APIServer) getMetricsByTimeRange(c *gin.Context) {
	// 获取查询参数
	startStr := c.DefaultQuery("start", "0")
	endStr := c.DefaultQuery("end", strconv.FormatInt(time.Now().UnixMilli(), 10))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// 解析时间戳
	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start timestamp"})
		return
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end timestamp"})
		return
	}

	// 转换为time.Time
	startTime := time.UnixMilli(start)
	endTime := time.UnixMilli(end)

	// 调用存储层获取数据
	metrics, err := s.storage.GetMetricsByTimeRange(startTime, endTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Stop 停止API服务器
func (s *APIServer) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(nil)
	}
	return nil
}
