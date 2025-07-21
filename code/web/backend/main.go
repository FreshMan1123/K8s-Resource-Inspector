package backend

import (
	"embed"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/handlers"
)

//go:embed static/*
var staticFiles embed.FS

// WebService Web服务
type WebService struct {
	ruleHandler       *handlers.RuleHandler
	inspectionHandler *handlers.InspectionHandler
}

// NewWebService 创建Web服务实例
func NewWebService() *WebService {
	return &WebService{
		ruleHandler:       handlers.NewRuleHandler(),
		inspectionHandler: handlers.NewInspectionHandler(),
	}
}

// SetupRoutes 设置路由
func (ws *WebService) SetupRoutes() *gin.Engine {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// 静态文件服务
	r.StaticFS("/static", http.FS(staticFiles))

	// 首页
	r.GET("/", ws.indexHandler)

	// API路由
	api := r.Group("/api")
	{
		// 规则管理API
		rules := api.Group("/rules")
		{
			rules.GET("", ws.ruleHandler.GetAllRules)
			rules.GET("/:category", ws.ruleHandler.GetRulesByCategory)
			rules.GET("/:category/:id", ws.ruleHandler.GetRuleByID)
		}

		// 分类API
		api.GET("/categories", ws.ruleHandler.GetCategories)

		// 巡检API
		inspection := api.Group("/inspection")
		{
			inspection.GET("/templates", ws.inspectionHandler.GetCommandTemplates)
			inspection.POST("/execute", ws.inspectionHandler.ExecuteCommand)
		}
	}

	return r
}

// indexHandler 首页处理器
func (ws *WebService) indexHandler(c *gin.Context) {
	// 读取嵌入的HTML文件
	htmlContent, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "页面加载失败")
		return
	}
	
	c.Data(http.StatusOK, "text/html; charset=utf-8", htmlContent)
}

// StartWebServer 启动Web服务器
func StartWebServer(port int) {
	webService := NewWebService()
	router := webService.SetupRoutes()

	addr := ":" + strconv.Itoa(port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("启动Web服务器失败: %v", err)
	}
}

// 注释掉main函数，因为这个包现在是可导入的
// func main() {
// 	StartWebServer(8080)
// }
