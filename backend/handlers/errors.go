package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UnableToParseRequestBody(ctx *gin.Context, err error) {
	slog.Error("unable to parse request body", "err", err)
	ctx.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse request body"})
}

func HandleError(ctx *gin.Context, httpCode int, errorMsg string, err error) {
	slog.Error(errorMsg, "err", err)
	ctx.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
}
