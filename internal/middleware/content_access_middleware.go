package middleware

import (
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/ginx"

	"github.com/gin-gonic/gin"
)

// ContentAccessMiddleware enforces the strict authenticated-content policy.
func ContentAccessMiddleware(ctx *gin.Context) {
	if common.GetCurrentUser(ctx) == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		ctx.Abort()
		return
	}
	ctx.Next()
}
