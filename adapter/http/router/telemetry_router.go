package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

func SetupTelemetryRoutes(router *gin.Engine, telemetryHandler *handler.TelemetryHandler) {
	router.POST("api/v1/telemetry-events", telemetryHandler.IngestBatch)
}
