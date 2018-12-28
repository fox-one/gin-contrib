package i18n

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const localizerContextKey = "i18n_localizer_context_key"

func NewBundle(defaultLang language.Tag, rootPath string) *i18n.Bundle {
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

func WithI18nBundle(b *i18n.Bundle) gin.HandlerFunc {
	return func(c *gin.Context) {
		localizer := i18n.NewLocalizer(b, c.GetHeader("Accept-Language"))
		c.Set(localizerContextKey, localizer)
	}
}

func ExtractLocalizer(c *gin.Context) *i18n.Localizer {
	return c.MustGet(localizerContextKey).(*i18n.Localizer)
}
