package params

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"bbs-go/internal/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/spf13/cast"
)

func paramError(name string) error {
	return fmt.Errorf("unable to find param value '%s'", name)
}

func Get(ctx *gin.Context, name string) (string, bool) {
	str := QueryValue(ctx, name)
	return str, str != ""
}

func QueryValue(ctx *gin.Context, name string) string {
	return ginx.FormValue(ctx, name)
}

func GetInt64(c *gin.Context, name string) (int64, bool) {
	str, ok := Get(c, name)
	if !ok {
		return 0, false
	}
	value, err := cast.ToInt64E(str)
	if err != nil {
		return 0, false
	}
	return value, true
}

func GetInt(c *gin.Context, name string) (int, bool) {
	str, ok := Get(c, name)
	if !ok {
		return 0, false
	}
	value, err := cast.ToIntE(str)
	if err != nil {
		return 0, false
	}
	return value, true
}

func GetInt64Arr(c *gin.Context, name string) []int64 {
	str, ok := Get(c, name)
	if !ok {
		return nil
	}
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		var ret []int64
		if err := jsons.Parse(str, &ret); err != nil {
			slog.Error(err.Error())
		}
		return ret
	}
	return StrSplitToInt64Arr(str)
}

func StrSplitToInt64Arr(str string) (ret []int64) {
	if strs.IsNotBlank(str) {
		for s := range strings.SplitSeq(str, ",") {
			if i, err := cast.ToInt64E(strings.TrimSpace(s)); err == nil {
				ret = append(ret, i)
			}
		}
	}
	return
}

func FormValue(ctx *gin.Context, name string) string {
	return ginx.FormValue(ctx, name)
}

func FormValueInt(ctx *gin.Context, name string) (int, error) {
	str := ginx.FormValue(ctx, name)
	if str == "" {
		return 0, paramError(name)
	}
	return strconv.Atoi(str)
}

func FormValueIntDefault(ctx *gin.Context, name string, def int) int {
	if v, err := FormValueInt(ctx, name); err == nil {
		return v
	}
	return def
}

func FormValueInt64(ctx *gin.Context, name string) (int64, error) {
	str := ginx.FormValue(ctx, name)
	if str == "" {
		return 0, paramError(name)
	}
	return strconv.ParseInt(str, 10, 64)
}

func FormValueInt64Default(ctx *gin.Context, name string, def int64) int64 {
	if v, err := FormValueInt64(ctx, name); err == nil {
		return v
	}
	return def
}

func FormValueInt64Array(ctx *gin.Context, name string) []int64 {
	str := strings.TrimSpace(ginx.FormValue(ctx, name))
	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		var ret []int64
		if err := jsons.Parse(str, &ret); err != nil {
			slog.Error(err.Error())
		}
		return ret
	}
	return StrSplitToInt64Arr(str)
}

func FormValueBool(ctx *gin.Context, name string) (bool, error) {
	str := ginx.FormValue(ctx, name)
	if str == "" {
		return false, paramError(name)
	}
	return strconv.ParseBool(str)
}

func FormValueBoolDefault(ctx *gin.Context, name string, def bool) bool {
	str := ginx.FormValue(ctx, name)
	if str == "" {
		return def
	}
	value, err := strconv.ParseBool(str)
	if err != nil {
		return def
	}
	return value
}

func GetPaging(ctx *gin.Context) *sqls.Paging {
	page := FormValueIntDefault(ctx, "page", 1)
	limit := FormValueIntDefault(ctx, "limit", 20)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	return &sqls.Paging{Page: page, Limit: limit}
}
