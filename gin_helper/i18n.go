package gin_helper

import (
	"github.com/fox-one/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

func Localize(c *gin.Context, id string, paras ...interface{}) string {
	l := i18n.ExtractLocalizer(c)
	data := make(map[string]interface{})
	for idx := 0; idx < len(paras)-1; idx += 1 {
		k, v := paras[idx].(string), paras[idx+1]
		data[k] = v
	}

	return l.MustLocalize(&goi18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
}
