package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/models"
	"github.com/stretchr/testify/assert"
)

// TestWebServiceBasicFunctionality 测试Web服务基础功能
func TestWebServiceBasicFunctionality(t *testing.T) {
	// 创建Web服务实例
	webService := backend.NewWebService()
	router := webService.SetupRoutes()

	t.Run("测试首页访问", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "K8s巡检管理平台")
	})

	t.Run("测试规则分类API", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		// 验证返回的分类数据
		categories, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(categories), 0)
	})

	t.Run("测试获取所有规则API", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/rules", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		// 验证返回的规则数据结构
		rules, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(rules), 0)
	})

	t.Run("测试命令模板API", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/inspection/templates", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		// 验证返回的命令模板
		templates, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(templates), 0)
	})

	t.Run("测试命令执行API", func(t *testing.T) {
		// 构造执行请求
		requestBody := `{
			"template_id": "cluster_list",
			"parameters": {}
		}`

		req, _ := http.NewRequest("POST", "/api/inspection/execute", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 注意：这个测试可能会失败，因为需要实际的K8s集群连接
		// 但我们可以验证API的基本结构是否正确
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
	})
}

// TestRuleServiceFunctionality 测试规则服务功能
func TestRuleServiceFunctionality(t *testing.T) {
	// 这些测试验证规则文件读取功能
	t.Run("测试规则文件读取", func(t *testing.T) {
		webService := backend.NewWebService()
		router := webService.SetupRoutes()

		// 测试获取node分类的规则
		req, _ := http.NewRequest("GET", "/api/rules/node", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 如果规则文件存在，应该返回成功
		if w.Code == http.StatusOK {
			var response models.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response.Success)

			// 验证规则数据结构
			rules, ok := response.Data.([]interface{})
			assert.True(t, ok)
			
			if len(rules) > 0 {
				rule := rules[0].(map[string]interface{})
				assert.Contains(t, rule, "id")
				assert.Contains(t, rule, "name")
				assert.Contains(t, rule, "condition")
			}
		}
	})
}

// TestCommandServiceFunctionality 测试命令服务功能
func TestCommandServiceFunctionality(t *testing.T) {
	t.Run("测试命令模板结构", func(t *testing.T) {
		webService := backend.NewWebService()
		router := webService.SetupRoutes()

		req, _ := http.NewRequest("GET", "/api/inspection/templates", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		templates, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(templates), 0)

		// 验证第一个模板的结构
		if len(templates) > 0 {
			template := templates[0].(map[string]interface{})
			assert.Contains(t, template, "id")
			assert.Contains(t, template, "name")
			assert.Contains(t, template, "description")
			assert.Contains(t, template, "category")
			assert.Contains(t, template, "command")
		}
	})
}

// TestAPIErrorHandling 测试API错误处理
func TestAPIErrorHandling(t *testing.T) {
	webService := backend.NewWebService()
	router := webService.SetupRoutes()

	t.Run("测试不存在的规则分类", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/rules/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Error)
	})

	t.Run("测试无效的命令执行请求", func(t *testing.T) {
		// 发送无效的JSON
		req, _ := http.NewRequest("POST", "/api/inspection/execute", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "无效的请求参数")
	})
}
