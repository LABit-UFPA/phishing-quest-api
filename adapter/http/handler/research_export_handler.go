package handler

import (
	"encoding/csv"
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ResearchExportHandler struct {
	exportUseCase *usecase.ResearchExportUseCase
}

func NewResearchExportHandler(euc *usecase.ResearchExportUseCase) *ResearchExportHandler {
	return &ResearchExportHandler{exportUseCase: euc}
}

// ExportCSV gera um CSV pseudonimizado com todas as tentativas
// coletadas (GET /api/v1/research/export?format=csv). Protegido por
// RequireRole(researcher, admin) no router — nunca chega aqui sem
// essa role.
func (reh *ResearchExportHandler) ExportCSV(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	if format != "csv" {
		c.JSON(http.StatusBadRequest, response.Error("formato nao suportado, use format=csv"))
		return
	}

	rows, err := reh.exportUseCase.ExportAttempts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=research_export.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{
		"pseudo_user_id", "item_id", "channel", "is_malicious", "session_id",
		"condition", "verdict", "action", "confidence", "is_correct",
		"latency_ms", "clicked_link", "created_at",
	})

	for _, row := range rows {
		_ = writer.Write([]string{
			row.PseudoUserId,
			row.ItemId.String(),
			row.Channel,
			strconv.FormatBool(row.IsMalicious),
			row.SessionId.String(),
			row.Condition,
			formatNullableBool(row.Verdict),
			string(row.Action),
			formatNullableInt(row.Confidence),
			formatNullableBool(row.IsCorrect),
			formatNullableInt(row.LatencyMs),
			strconv.FormatBool(row.ClickedLink),
			row.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func formatNullableBool(v *bool) string {
	if v == nil {
		return ""
	}
	return strconv.FormatBool(*v)
}

func formatNullableInt(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}
