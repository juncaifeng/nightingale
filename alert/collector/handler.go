package collector

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/prompb"
	"github.com/toolkits/pkg/logger"
)

type AlertCollectorHandler struct {
	registrar *AlertMetricsRegistrar
	writer    WriterInterface
}

type WriterInterface interface {
	Write(ts *prompb.TimeSeries)
}

func NewAlertCollectorHandler(registrar *AlertMetricsRegistrar, writer WriterInterface) *AlertCollectorHandler {
	return &AlertCollectorHandler{
		registrar: registrar,
		writer:    writer,
	}
}

func (h *AlertCollectorHandler) CollectAlerts(c *gin.Context) {
	var events []AlertEvent

	if err := c.ShouldBindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if len(events) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no alert events provided",
		})
		return
	}

	successCount := 0
	errorCount := 0

	for _, event := range events {
		tsList := h.registrar.EventToTimeSeries(&event)

		for _, ts := range tsList {
			if ts != nil {
				h.writer.Write(ts)
				successCount++
			} else {
				errorCount++
			}
		}
	}

	logger.Infof("alert_collector: received %d events, converted to %d metrics (errors: %d)",
		len(events), successCount, errorCount)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"total":   len(events),
		"metrics": successCount,
		"errors":  errorCount,
	})
}

func (h *AlertCollectorHandler) CollectSingleAlert(c *gin.Context) {
	var event AlertEvent

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	tsList := h.registrar.EventToTimeSeries(&event)

	for _, ts := range tsList {
		if ts != nil {
			h.writer.Write(ts)
		}
	}

	logger.Infof("alert_collector: received single alert '%s' from '%s', converted to %d metrics",
		event.Name, event.Ident, len(tsList))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"metrics": len(tsList),
	})
}
