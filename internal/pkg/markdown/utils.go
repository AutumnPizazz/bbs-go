package markdown

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"bbs-go/internal/pkg/html"

	"github.com/88250/lute"
	"github.com/mlogclub/simple/common/strs"
)

var (
	engine    *lute.Lute
	once      sync.Once
	htmlCache sync.Map // map[string]string: sha256(markdown) -> html
)

func getEngine() *lute.Lute {
	once.Do(func() {
		engine = lute.New(func(l *lute.Lute) {
			// ToC 由 misc_render.go 的 goquery 后处理统一生成，这里不启用引擎级 ToC
			// l.SetToC(true)

			// ========== P0-1: 关闭引擎级 XSS 过滤，统一由后端的 bluemonday 负责 ==========
			// SetSanitize(true) 会误杀 GFM 规范中的合法 HTML（如 <kbd>、<details>、<figure> 等），
			// 而后端 bluemonday 已配置为允许这些标签 + class + style 属性，更灵活且安全。
			l.SetSanitize(false)

			// GFM 任务列表
			l.SetGFMTaskListItem(true)

			// ========== P1-1: 启用 Emoji 短码解析 (:smile: → 😄) ==========
			l.SetEmoji(true)

			// ========== P0-2: CSS 类名模式代码高亮（自适应 light/dark） ==========
			// 生成 .highlight-kn / .highlight-kd / .highlight-s 等 CSS 类名，
			// 由 code-block.css 中用 CSS 变量提供 light 和 dark 两套配色。
			l.SetCodeSyntaxHighlightInlineStyle(false) // CSS 类名模式，配合自适应主题
			l.SetCodeSyntaxHighlightLineNum(false) // 行号可后续按需开启
		})
	})
	return engine
}

// cacheKey 为 markdown 原文生成缓存键
func cacheKey(markdownStr string) string {
	h := sha256.Sum256([]byte(markdownStr))
	return hex.EncodeToString(h[:])
}

// ToHTML 将 Markdown 转为 HTML，结果会被缓存。
// 同一份 Markdown 原文多次渲染时直接返回缓存结果。
func ToHTML(markdownStr string) string {
	if strs.IsBlank(markdownStr) {
		return ""
	}

	key := cacheKey(markdownStr)
	if cached, ok := htmlCache.Load(key); ok {
		return cached.(string)
	}

	result := getEngine().MarkdownStr("", markdownStr)
	htmlCache.Store(key, result)
	return result
}

// InvalidateCache 清除指定 Markdown 文本的缓存。
// 当帖子/评论内容更新时可调用。
func InvalidateCache(markdownStr string) {
	if strs.IsBlank(markdownStr) {
		return
	}
	htmlCache.Delete(cacheKey(markdownStr))
}

// InvalidateCacheBulk 批量清除缓存。
func InvalidateCacheBulk(markdownStrings ...string) {
	for _, s := range markdownStrings {
		InvalidateCache(s)
	}
}

// GetSummary 从 Markdown 提取纯文本摘要
func GetSummary(markdownStr string, summaryLen int) string {
	htmlStr := ToHTML(markdownStr)
	return html.GetSummary(htmlStr, summaryLen)
}
