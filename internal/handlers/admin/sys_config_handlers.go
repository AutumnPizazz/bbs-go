package admin

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web"

	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/services"
)

func SysConfigDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.SysConfigService.Get(id)
	if t == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, redactSysConfig(t))

}

func SysConfigList(ctx *gin.Context) {
	list, paging := services.SysConfigService.FindPageByParams(params.NewQueryParams(ctx).PageByReq().Desc("id"))
	for i := range list {
		list[i] = *redactSysConfig(&list[i])
	}
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})

}

func SysConfigConfigs(ctx *gin.Context) {

	resp := &dto.SysConfigAdminResponse{
		SiteTitle:          cache.SysConfigCache.GetStr(constants.SysConfigSiteTitle),
		SiteDescription:    cache.SysConfigCache.GetStr(constants.SysConfigSiteDescription),
		BaseURL:            services.SysConfigService.GetBaseURL(),
		SiteKeywords:       cache.SysConfigCache.GetStrArr(constants.SysConfigSiteKeywords),
		SiteLogo:           cache.SysConfigCache.GetStr(constants.SysConfigSiteLogo),
		SiteNavs:           services.SysConfigService.GetSiteNavs(),
		SiteNotification:   cache.SysConfigCache.GetStr(constants.SysConfigSiteNotification),
		AboutPageConfig:    services.SysConfigService.GetAboutPageConfig(),
		FooterLinks:        services.SysConfigService.GetFooterLinks(),
		RecommendTags:      cache.SysConfigCache.GetStrArr(constants.SysConfigRecommendTags),
		UrlRedirect:        services.SysConfigService.IsUrlRedirect(),
		DefaultCategoryId:  services.SysConfigService.GetDefaultCategoryId(),
		TopicListStyle:     services.SysConfigService.GetTopicListStyle(),
		TopicCaptcha:       services.SysConfigService.IsTopicCaptcha(),
		UserObserveSeconds: cache.SysConfigCache.GetInt(constants.SysConfigUserObserveSeconds),
		TokenExpireDays:    services.SysConfigService.GetTokenExpireDays(),
		EnableHideContent:  services.SysConfigService.IsEnableHideContent(),
		Modules:            services.SysConfigService.GetModules(),
		NotificationTypes:  services.SysConfigService.GetNotificationTypes(),
		LoginConfig:        services.SysConfigService.GetLoginConfigAdmin(),
		UploadConfig:       services.SysConfigService.GetUploadConfigAdmin(),
		AttachmentConfig:   services.SysConfigService.GetAttachmentConfig(),
		ScriptInjections:   services.SysConfigService.GetScriptInjections(),
	}
	if strs.IsBlank(resp.SiteLogo) {
		resp.SiteLogo = "/res/images/logo.png"
	}
	ginx.WriteJSON(ctx, resp)

}

func SysConfigSave(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.SysConfigService.SetAll(string(body)); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntitySysConfig, 0,
			"更新普通系统设置：keys="+strings.Join(configKeys(body), ","), ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)

}

func SysConfigSaveSensitive(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.SysConfigService.SetSensitive(string(body)); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntitySysConfig, 0,
			"更新敏感系统设置（不记录配置值）：keys="+strings.Join(configKeys(body), ","), ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)
}

func redactSysConfig(value *models.SysConfig) *models.SysConfig {
	if value == nil {
		return nil
	}
	copy := *value
	if services.SysConfigService.IsSensitiveKey(copy.Key) {
		if strings.TrimSpace(copy.Value) == "" {
			copy.Value = ""
		} else {
			copy.Value = "[configured]"
		}
	}
	return &copy
}

func configKeys(body []byte) []string {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(body, &values); err != nil {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if key != "clearSensitive" {
			keys = append(keys, key)
		}
	}
	return keys
}
