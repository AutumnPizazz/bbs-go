package render

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/bbsurls"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/locales"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"

	"github.com/microcosm-cc/bluemonday"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/common/urls"
	"github.com/mlogclub/simple/web"

	"bbs-go/internal/models"
	"bbs-go/internal/services"
)

// calloutPattern 匹配 GFM 扩展 callout 语法：> [!TYPE]\n> content
// 支持的 TYPE: NOTE, INFO, TIP, WARNING, DANGER, SUCCESS, IMPORTANT, CAUTION
var calloutPattern = regexp.MustCompile(`(?i)^\s*\[!(NOTE|INFO|TIP|WARNING|DANGER|SUCCESS|IMPORTANT|CAUTION)\]\s*`)

// calloutTypeToCSS 将 callout 类型映射到 CSS 类名
var calloutTypeToCSS = map[string]string{
	"NOTE":      "callout-info",
	"INFO":      "callout-info",
	"TIP":       "callout-tip",
	"WARNING":   "callout-warning",
	"DANGER":    "callout-warning",
	"SUCCESS":   "callout-success",
	"IMPORTANT": "callout-tip",
	"CAUTION":   "callout-warning",
}

// calloutTypeToIcon 将 callout 类型映射到图标（emoji）
var calloutTypeToIcon = map[string]string{
	"NOTE":      "📝",
	"INFO":      "ℹ️",
	"TIP":       "💡",
	"WARNING":   "⚠️",
	"DANGER":    "🚫",
	"SUCCESS":   "✅",
	"IMPORTANT": "🔔",
	"CAUTION":   "⚡",
}

func xssProtection(htmlContent string) string {
	ugcProtection := bluemonday.UGCPolicy() // 用户生成内容模式

	// 放开 style 属性（bluemonday 内部会过滤 expression/javascript: 等危险 CSS）
	// TipTap 富文本编辑器的颜色、对齐、背景色等均依赖 inline style
	ugcProtection.AllowAttrs("style").Globally()

	// 放开 class 属性，使编辑器生成的类名样式能够保留
	// （如高亮、表格样式、代码块语言标记等）
	ugcProtection.AllowAttrs("class").OnElements(
		"span", "div", "p",
		"table", "thead", "tbody", "tfoot", "tr", "td", "th", "caption", "colgroup", "col",
		"pre", "code", "blockquote",
		"ul", "ol", "li",
		"a", "img",
		"mark", "details", "summary", "figure", "figcaption",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"b", "i", "u", "s", "em", "strong", "small", "sub", "sup",
		"input", // P0-3: 放开 input 标签以支持 GFM 任务列表 checkbox
	)

	// P0-3: 允许 GFM 任务列表 checkbox 相关属性
	ugcProtection.AllowAttrs("type", "checked", "disabled").OnElements("input")

	// P1-3: 允许原生懒加载属性
	ugcProtection.AllowAttrs("loading").OnElements("img")

	ugcProtection.AllowAttrs("start").OnElements("ol", "ul", "li")
	return ugcProtection.Sanitize(htmlContent)
}

// HandleHtmlContent 处理html内容（导出供 markdown API 使用）
func HandleHtmlContent(htmlContent string) string {
	htmlContent, _ = handleHtmlContentWithToc(htmlContent, false)
	return htmlContent
}

// handleHtmlContent 处理html内容（内部使用，保持向后兼容）
func handleHtmlContent(htmlContent string) string {
	return HandleHtmlContent(htmlContent)
}

func handleTopicHtmlContent(htmlContent string) (string, []resp.TopicTocItem) {
	return handleHtmlContentWithToc(htmlContent, true)
}

func handleHtmlContentWithToc(htmlContent string, buildToc bool) (string, []resp.TopicTocItem) {
	htmlContent = xssProtection(htmlContent)

	// ========== P2-1: Callout 后处理——在 goquery 解析前先转换 ==========
	// 将 GFM 扩展 callout 块引用（> [!TYPE]）转为带 CSS 类的 div
	htmlContent = convertCallouts(htmlContent)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent, nil
	}

	doc.Find("a").Each(func(_ int, selection *goquery.Selection) {
		href := selection.AttrOr("href", "")

		if strs.IsBlank(href) {
			return
		}

		// 不是内部链接
		if !bbsurls.IsInternalUrl(href) {
			selection.SetAttr("target", "_blank")
			selection.SetAttr("rel", "external nofollow") // 标记站外链接，搜索引擎爬虫不传递权重值

			if services.SysConfigService.IsUrlRedirect() { // 开启非内部链接跳转
				newHref := urls.ParseUrl(bbsurls.AbsUrl("/redirect")).AddQuery("url", href).BuildStr()
				selection.SetAttr("href", newHref)
			}
		}

		// 如果a标签没有title，那么设置title
		title := selection.AttrOr("title", "")
		if len(title) == 0 {
			selection.SetAttr("title", selection.Text())
		}
	})

	// 处理图片
	doc.Find("img").Each(func(_ int, selection *goquery.Selection) {
		src := selection.AttrOr("src", "")

		// 处理第三方图片
		if strings.Contains(src, "qpic.cn") {
			src = urls.ParseUrl("/api/img/proxy").AddQuery("url", src).BuildStr()
		}

		// 处理图片样式
		src = HandleOssImageStyleDetail(src)

		selection.SetAttr("src", src)

		// ========== P1-3: 原生懒加载 ==========
		// 使用浏览器原生 loading="lazy"，无需 JS
		if _, has := selection.Attr("loading"); !has {
			selection.SetAttr("loading", "lazy")
		}
	})

	// ========== 代码块增强：包裹容器 + 语言标签（服务端渲染，保证可靠） ==========
	doc.Find("pre").Each(func(_ int, pre *goquery.Selection) {
		code := pre.Find("code")
		if code.Length() == 0 {
			return
		}

		// 提取语言标签
		cls := code.AttrOr("class", "")
		lang := ""
		if idx := strings.Index(cls, "language-"); idx >= 0 {
			rest := cls[idx+len("language-"):]
			if space := strings.IndexAny(rest, " \t"); space >= 0 {
				lang = rest[:space]
			} else {
				lang = rest
			}
		}

		// 获取 pre 的属性字符串（保留 data-* 等）
		preAttrs := ""
		if len(pre.Nodes) > 0 {
			for _, attr := range pre.Nodes[0].Attr {
				preAttrs += fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val)
			}
		}
		preInner, err := pre.Html()
		if err != nil || preInner == "" {
			preInner = code.Text() // 回退到纯文本
		}

		wrapper := fmt.Sprintf(
			`<div class="code-block-wrapper"><div class="code-block-header"><span class="code-block-lang">%s</span><button class="code-block-copy-btn" type="button" onclick="__bbsCopyCode(this)"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg><span>复制</span></button></div><pre%s>%s</pre></div>`,
			lang, preAttrs, preInner,
		)
		pre.ReplaceWithHtml(wrapper)
	})

	// ========== P0-3: 任务列表 checkbox 设为 disabled（只读展示） ==========
	doc.Find("input[type='checkbox']").Each(func(_ int, selection *goquery.Selection) {
		selection.SetAttr("disabled", "disabled")
	})

	// ========== 代码块：移除 Chroma 内联 background-color（避免 dark 模式下白色背景） ==========
	doc.Find("pre").Each(func(_ int, selection *goquery.Selection) {
		// Chroma 内联样式会给 <pre> 设置 background-color:#ffffff，
		// 必须移除才能让 CSS 变量 background 在 dark 模式下正常生效。
		style := selection.AttrOr("style", "")
		style = strings.ReplaceAll(style, "background-color: #ffffff", "")
		style = strings.ReplaceAll(style, "background-color:#ffffff", "")
		style = strings.ReplaceAll(style, "background-color: #fff", "")
		style = strings.ReplaceAll(style, "background-color:#fff", "")
		style = strings.ReplaceAll(style, "background-color:#f8f8f8", "")
		style = strings.TrimSpace(strings.ReplaceAll(style, ";;", ";"))
		if style == "" {
			selection.RemoveAttr("style")
		} else {
			selection.SetAttr("style", style)
		}
	})

	var toc []resp.TopicTocItem
	if buildToc {
		toc = buildTopicToc(doc)
	}

	// ========== P1-2: 给标题注入 hover 可见的锚点链接 ==========
	injectHeadingAnchors(doc)

	if htmlStr, err := doc.Find("body").Html(); err == nil {
		return htmlStr, toc
	}
	return htmlContent, toc
}

func buildTopicToc(doc *goquery.Document) []resp.TopicTocItem {
	var toc []resp.TopicTocItem
	usedIds := make(map[string]int)
	doc.Find("h1,h2,h3,h4,h5,h6").Each(func(_ int, selection *goquery.Selection) {
		title := strings.TrimSpace(selection.Text())
		if strs.IsBlank(title) {
			return
		}

		level := headingLevel(selection)
		id := uniqueHeadingId(slugHeading(title), usedIds)
		selection.SetAttr("id", id)
		toc = append(toc, resp.TopicTocItem{
			Id:    id,
			Title: title,
			Level: level,
		})
	})
	return toc
}

// ========== P1-2: 给每个标题注入 hover 可见的 # 锚点链接 ==========
func injectHeadingAnchors(doc *goquery.Document) {
	doc.Find("h1[id],h2[id],h3[id],h4[id],h5[id],h6[id]").Each(func(_ int, selection *goquery.Selection) {
		id, exists := selection.Attr("id")
		if !exists || id == "" {
			return
		}
		// 在标题内部末尾追加锚点链接
		selection.AppendHtml(fmt.Sprintf(
			`<a class="heading-anchor" href="#%s" aria-label="锚点链接">#</a>`, id,
		))
	})
}

// ========== P2-1: 将 GFM callout 块引用转为 callout div ==========
// convertCallouts 在 HTML 文本中查找 <blockquote> 内以 [TYPE] 开头的段落，
// 并将其整体替换为带 CSS 类的 callout 面板。
// 此函数在 goquery 解析之前运行，因为 goquery 需要能解析最终结构。
func convertCallouts(htmlContent string) string {
	// 匹配 <blockquote> 块，检测第一个 <p> 是否含 [TYPE] 标记
	// 使用简单的字符串替换策略，因为 goquery 之前不能可靠解析
	type calloutMatch struct {
		fullBlock  string // 完整原始 <blockquote>...</blockquote>
		innerHTML  string // blockquote 内部 HTML
		calloutType string // NOTE/INFO/TIP/WARNING/DANGER/SUCCESS/IMPORTANT/CAUTION
	}

	var result strings.Builder
	remaining := htmlContent

	for {
		// 查找下一个 <blockquote
		startIdx := strings.Index(remaining, "<blockquote")
		if startIdx < 0 {
			result.WriteString(remaining)
			break
		}

		// 写入 blockquote 之前的内容
		result.WriteString(remaining[:startIdx])
		remaining = remaining[startIdx:]

		// 查找 </blockquote> 结束标签
		endIdx := findBlockquoteEnd(remaining)
		if endIdx < 0 {
			result.WriteString(remaining)
			break
		}

		blockHTML := remaining[:endIdx]
		remaining = remaining[endIdx:]

		// 尝试解析此 blockquote 是否为 callout
		converted := tryConvertSingleCallout(blockHTML)
		result.WriteString(converted)
	}

	return result.String()
}

// findBlockquoteEnd 找到 </blockquote> 结束标签位置
func findBlockquoteEnd(html string) int {
	depth := 0
	for i := 0; i < len(html); i++ {
		if strings.HasPrefix(html[i:], "<blockquote") {
			depth++
		} else if strings.HasPrefix(html[i:], "</blockquote>") {
			depth--
			if depth == 0 {
				return i + len("</blockquote>")
			}
		}
	}
	return -1
}

// tryConvertSingleCallout 尝试将单个 <blockquote> 转为 callout div
func tryConvertSingleCallout(blockHTML string) string {
	// 解析出 blockquote 内部 HTML
	// <blockquote>\n<p>[!NOTE] ...</p>\n</blockquote>
	innerStart := strings.Index(blockHTML, ">")
	if innerStart < 0 {
		return blockHTML
	}
	innerStart++ // 跳过 >
	innerHTML := blockHTML[innerStart : len(blockHTML)-len("</blockquote>")]

	// 查找第一个 <p> 的内容
	pStart := strings.Index(innerHTML, "<p>")
	if pStart < 0 {
		return blockHTML
	}
	pEnd := strings.Index(innerHTML, "</p>")
	if pEnd < 0 || pEnd <= pStart {
		return blockHTML
	}
	pContent := innerHTML[pStart+3 : pEnd]

	// 检查是否匹配 [!TYPE]
	matches := calloutPattern.FindStringSubmatch(pContent)
	if matches == nil {
		return blockHTML
	}

	calloutType := strings.ToUpper(matches[1])
	cssClass := calloutTypeToCSS[calloutType]
	icon := calloutTypeToIcon[calloutType]

	// 替换掉 <p> 中的 [!TYPE] 标记，取剩余内容作为标题/描述
	cleanContent := calloutPattern.ReplaceAllString(pContent, "")
	cleanContent = strings.TrimSpace(cleanContent)

	// 构建新的 callout innerHTML：图标 + 第一个 <p>（清理后）+ 后续内容
	var newInner strings.Builder
	newInner.WriteString(fmt.Sprintf(`<span class="callout-icon">%s</span>`, icon))
	newInner.WriteString(`<span class="callout-content">`)

	if cleanContent != "" {
		newInner.WriteString("<p>")
		newInner.WriteString(cleanContent)
		newInner.WriteString("</p>")
	}

	// 保留第一个 </p> 之后的其他内容
	restContent := innerHTML[pEnd+4:]
	newInner.WriteString(restContent)
	newInner.WriteString(`</span>`)

	return fmt.Sprintf(`<div class="callout %s">%s</div>`, cssClass, newInner.String())
}

func headingLevel(selection *goquery.Selection) int {
	switch goquery.NodeName(selection) {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 0
	}
}

func slugHeading(title string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}

	slug := strings.Trim(builder.String(), "-")
	if strs.IsBlank(slug) {
		return "section"
	}
	return "topic-heading-" + slug
}

func uniqueHeadingId(base string, usedIds map[string]int) string {
	count := usedIds[base]
	usedIds[base] = count + 1
	if count == 0 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, count+1)
}

/*
BuildLoginSuccess 处理登录成功后的返回数据

Parameter:

	user - login user
	redirect - 登录来源地址，需要控制登录成功之后跳转到该地址
*/
func BuildLoginSuccess(ctx *gin.Context, user *models.User, redirect string) *web.JsonResult {
	if user == nil || user.Status != constants.StatusOk {
		return web.JsonErrorMsg(locales.Get("errors.user_not_found_or_disabled"))
	}
	token, err := services.UserTokenService.Generate(user.Id)
	if err != nil {
		return web.JsonError(err)
	}
	ginx.SetCookieKV(ctx, constants.CookieTokenKey, token, ginx.CookieHTTPOnly(true), ginx.CookieExpires(365*24*time.Hour))
	ginx.SetCookieKV(ctx, constants.CookieCSRFTokenKey, strs.UUID(), ginx.CookieExpires(365*24*time.Hour))
	return web.NewEmptyRspBuilder().
		Put("token", token).
		Put("user", BuildUserProfile(user)).
		Put("redirect", redirect).JsonResult()
}
