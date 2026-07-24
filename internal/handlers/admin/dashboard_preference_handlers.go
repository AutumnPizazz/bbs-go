package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/services"
)

const maxDashboardViewKeyLength = 256

type dashboardViewPreferenceForm struct {
	Key   string                     `json:"key"`
	Views map[string]json.RawMessage `json:"views"`
}

func DashboardViewPreferences(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	key, err := validateDashboardViewKey(ctx.Query("key"))
	if err != nil {
		ginx.WriteHttpStatusJSON(ctx, http.StatusBadRequest, err)
		return
	}
	views := map[string]json.RawMessage{}
	if preference := services.DashboardViewPreferenceService.Get(operator.Id, key); preference != nil {
		if err := json.Unmarshal([]byte(preference.Views), &views); err != nil {
			views = map[string]json.RawMessage{}
		}
	}
	ginx.WriteJSON(ctx, map[string]interface{}{"key": key, "views": views})
}

func SaveDashboardViewPreferences(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	form := &dashboardViewPreferenceForm{}
	if err := ginx.BindJSON(ctx, form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	key, err := validateDashboardViewKey(form.Key)
	if err != nil {
		ginx.WriteHttpStatusJSON(ctx, http.StatusBadRequest, err)
		return
	}
	if form.Views == nil {
		form.Views = map[string]json.RawMessage{}
	}
	encoded, err := json.Marshal(form.Views)
	if err != nil || len(encoded) > services.MaxDashboardViewJSONLength() {
		ginx.WriteHttpStatusJSON(ctx, http.StatusBadRequest, errors.New("invalid dashboard view preference"))
		return
	}
	if err := services.DashboardViewPreferenceService.Save(operator.Id, key, string(encoded)); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "dashboardView", 0, "保存后台筛选视图", ctx.Request)
	ginx.WriteJSON(ctx, map[string]interface{}{"key": key, "views": form.Views})
}

func validateDashboardViewKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" || len(key) > maxDashboardViewKeyLength || !strings.HasPrefix(key, "/api/admin/") {
		return "", errors.New("invalid dashboard view key")
	}
	return key, nil
}
