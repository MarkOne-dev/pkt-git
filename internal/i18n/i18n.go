// Package i18n provides multi-language localization for pkt-git using embedded catalogs.
package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	mu          sync.RWMutex
	bundle      *i18n.Bundle
	localizer   *i18n.Localizer
	currentLang string
)

func init() {
	Init("")
}

// GitConfigLanguageFunc retrieves the configured language from git config (overridable in tests).
var GitConfigLanguageFunc = func() string {
	out, err := exec.Command("git", "config", "pkt.language").Output()
	if err == nil {
		val := strings.ToLower(strings.TrimSpace(string(out)))
		if val == "es" || val == "en" {
			return val
		}
	}
	return ""
}

// DetectLanguage returns the best language code ("en", "es") based on git config and environment variables.
func DetectLanguage() string {
	if cfg := GitConfigLanguageFunc(); cfg != "" {
		return cfg
	}

	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		val := strings.TrimSpace(os.Getenv(env))
		if val != "" {
			tag, err := language.Parse(strings.Split(val, ".")[0])
			if err == nil {
				base, _ := tag.Base()
				if base.String() == "es" {
					return "es"
				}
				if base.String() == "en" {
					return "en"
				}
			}
			lower := strings.ToLower(val)
			if strings.HasPrefix(lower, "es") {
				return "es"
			}
			if strings.HasPrefix(lower, "en") {
				return "en"
			}
		}
	}
	return "en"
}

// Init initializes or reconfigures the active language bundle.
// If lang is empty, it detects the language from the environment.
func Init(lang string) {
	mu.Lock()
	defer mu.Unlock()

	targetLang := strings.ToLower(strings.TrimSpace(lang))
	if targetLang == "" {
		targetLang = DetectLanguage()
	}

	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load embedded translation catalogs
	_, _ = bundle.LoadMessageFileFS(localeFS, "locales/en.json")
	_, _ = bundle.LoadMessageFileFS(localeFS, "locales/es.json")

	currentLang = targetLang
	localizer = i18n.NewLocalizer(bundle, targetLang, "en")
}

// CurrentLanguage returns the active language code.
func CurrentLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	if currentLang == "" {
		return "en"
	}
	return currentLang
}

// T translates a message by ID with optional template data.
func T(messageID string, templateData ...interface{}) string {
	mu.RLock()
	loc := localizer
	mu.RUnlock()

	if loc == nil {
		return messageID
	}

	var data interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil || msg == "" {
		return messageID
	}
	return msg
}
