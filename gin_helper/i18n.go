package gin_helper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func NewI18nBundle(defaultLang language.Tag, rootPath string) *i18n.Bundle {
	b := &i18n.Bundle{DefaultLanguage: defaultLang}
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	if rootInfo, err := os.Stat(rootPath); err == nil && rootInfo.IsDir() {
		filepath.Walk(rootPath, func(file string, info os.FileInfo, err error) error {
			if !info.IsDir() && filepath.Ext(file) == ".toml" {
				b.MustLoadMessageFile(file)
			}

			return nil
		})
	} else {
		panic(fmt.Errorf("%s is not a valid Dir", rootPath))
	}

	return b
}

func MustLocalize(b *i18n.Bundle, lang, id string, paras ...interface{}) string {
	l := i18n.NewLocalizer(b, lang)
	r, err := localize(l, id, paras...)
	if err != nil {
		panic(err)
	}

	return r
}

func Localize(b *i18n.Bundle, lang, id string, paras ...interface{}) string {
	l := i18n.NewLocalizer(b, lang)
	r, _ := localize(l, id, paras...)
	return r
}

func localize(l *i18n.Localizer, id string, paras ...interface{}) (string, error) {
	data := make(map[string]interface{})
	for idx := 0; idx < len(paras)-1; idx += 2 {
		k, v := paras[idx].(string), paras[idx+1]
		data[k] = v
	}

	return l.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
}

// gin

const (
	i18nBundleContextKey    = "i18n_bundle_context_key"
	i18nLocalizerContextKey = "i18n_localizer_context_key"
)

func UseI18nBundle(b *i18n.Bundle) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(i18nBundleContextKey, b)
	}
}

func ExtractI18nBundle(c *gin.Context) *i18n.Bundle {
	return c.MustGet(i18nBundleContextKey).(*i18n.Bundle)
}

func ExtractLocalizer(c *gin.Context) *i18n.Localizer {
	if l, ok := c.Get(i18nLocalizerContextKey); ok {
		return l.(*i18n.Localizer)
	}

	b := ExtractI18nBundle(c)
	l := i18n.NewLocalizer(b, c.GetHeader("Accept-Language"))
	c.Set(i18nLocalizerContextKey, l)

	return l
}

func LocalizeCtx(c *gin.Context, id string, paras ...interface{}) string {
	l := ExtractLocalizer(c)
	r, _ := localize(l, id, paras...)
	return r
}
