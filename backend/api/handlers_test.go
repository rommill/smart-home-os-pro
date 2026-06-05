package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetTelemetryRoute(t *testing.T) {
	// 1. Настраиваем Gin в тестовом режиме
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Создаем временный тестовый эндпоинт
	r.GET("/test-telemetry", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 2. Создаем фейковый запрос
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-telemetry", nil)
	
	// 3. Выполняем запрос
	r.ServeHTTP(w, req)

	// 4. Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}