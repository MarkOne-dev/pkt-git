package i18n

import (
	"encoding/json"
	"os"
	"testing"
)

func TestI18nEnglish(t *testing.T) {
	Init("en")
	if got := CurrentLanguage(); got != "en" {
		t.Errorf("CurrentLanguage() = %q, want %q", got, "en")
	}

	got := T("BannerSubtitle")
	want := "Cisco Packet Tracer ⇄ Git Version Control"
	if got != want {
		t.Errorf("T(BannerSubtitle) = %q, want %q", got, want)
	}

	gotParam := T("WelcomeReady", map[string]interface{}{"Version": "v1.0.0"})
	wantParam := "Environment ready (Git installed, pkt-git v1.0.0)"
	if gotParam != wantParam {
		t.Errorf("T(WelcomeReady) = %q, want %q", gotParam, wantParam)
	}
}

func TestI18nSpanish(t *testing.T) {
	Init("es")
	if got := CurrentLanguage(); got != "es" {
		t.Errorf("CurrentLanguage() = %q, want %q", got, "es")
	}

	got := T("BannerSubtitle")
	want := "Cisco Packet Tracer ⇄ Control de versiones Git"
	if got != want {
		t.Errorf("T(BannerSubtitle) = %q, want %q", got, want)
	}

	gotParam := T("WelcomeReady", map[string]interface{}{"Version": "v1.0.0"})
	wantParam := "Entorno listo (Git instalado, pkt-git v1.0.0)"
	if gotParam != wantParam {
		t.Errorf("T(WelcomeReady) = %q, want %q", gotParam, wantParam)
	}
}

func TestI18nFallback(t *testing.T) {
	Init("fr") // unsupported language
	got := T("BannerSubtitle")
	want := "Cisco Packet Tracer ⇄ Git Version Control"
	if got != want {
		t.Errorf("T(BannerSubtitle) fallback = %q, want %q", got, want)
	}
}

func TestDetectLanguage(t *testing.T) {
	origLang := os.Getenv("LANG")
	origLCAll := os.Getenv("LC_ALL")
	oldGitCfg := GitConfigLanguageFunc
	defer func() {
		os.Setenv("LANG", origLang)
		os.Setenv("LC_ALL", origLCAll)
		GitConfigLanguageFunc = oldGitCfg
	}()

	// 1. Test Git config precedence
	GitConfigLanguageFunc = func() string { return "es" }
	os.Setenv("LANG", "en_US.UTF-8")
	if got := DetectLanguage(); got != "es" {
		t.Errorf("DetectLanguage() with git config 'es' = %q, want 'es'", got)
	}

	// 2. Test environment fallback when git config is empty
	GitConfigLanguageFunc = func() string { return "" }
	os.Unsetenv("LC_ALL")
	os.Setenv("LANG", "es_ES.UTF-8")
	if got := DetectLanguage(); got != "es" {
		t.Errorf("DetectLanguage() = %q, want %q", got, "es")
	}

	os.Setenv("LANG", "en_US.UTF-8")
	if got := DetectLanguage(); got != "en" {
		t.Errorf("DetectLanguage() = %q, want %q", got, "en")
	}

	os.Setenv("LANG", "fr_FR.UTF-8")
	if got := DetectLanguage(); got != "en" {
		t.Errorf("DetectLanguage() fallback = %q, want %q", got, "en")
	}
}

func TestCatalogKeyParity(t *testing.T) {
	enBytes, err := localeFS.ReadFile("locales/en.json")
	if err != nil {
		t.Fatalf("failed to read locales/en.json: %v", err)
	}
	esBytes, err := localeFS.ReadFile("locales/es.json")
	if err != nil {
		t.Fatalf("failed to read locales/es.json: %v", err)
	}

	var enMap, esMap map[string]interface{}
	if err := json.Unmarshal(enBytes, &enMap); err != nil {
		t.Fatalf("failed to parse en.json: %v", err)
	}
	if err := json.Unmarshal(esBytes, &esMap); err != nil {
		t.Fatalf("failed to parse es.json: %v", err)
	}

	for k := range enMap {
		if _, ok := esMap[k]; !ok {
			t.Errorf("key %q present in en.json but missing from es.json", k)
		}
	}
	for k := range esMap {
		if _, ok := enMap[k]; !ok {
			t.Errorf("key %q present in es.json but missing from en.json", k)
		}
	}
}
