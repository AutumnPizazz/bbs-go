package api

import (
	"bbs-go/internal/handlers/render"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/markdown"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/strs"
)

// MarkdownRender 将 Markdown 文本渲染为 HTML（供编辑器预览使用）。
// POST /api/markdown/render
// Body: { "content": "markdown text" }
// Response: { "html": "rendered HTML" }
func MarkdownRender(ctx *gin.Context) {
	content := ctx.PostForm("content")
	if strs.IsBlank(content) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("内容不能为空"))
		return
	}

	// 使用与正式发布完全相同的渲染管线：markdown → HTML → 后处理
	html := markdown.ToHTML(content)
	html = render.HandleHtmlContent(html)

	ginx.WriteJSON(ctx, gin.H{
		"html": html,
	})
}
