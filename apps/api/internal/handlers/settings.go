package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"api/internal/store"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	store *store.Store
}

func NewSettingsHandler(st *store.Store) *SettingsHandler {
	return &SettingsHandler{store: st}
}

type infoLink struct {
	Label   string `json:"label"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

var defaultInfoLinks = []infoLink{
	{Label: "email", URL: "mailto:hello@example.com", Enabled: true},
	{Label: "instagram", URL: "https://instagram.com/", Enabled: true},
	{Label: "telegram", URL: "https://t.me/", Enabled: true},
}

const defaultInfoText = "i'm a photographer & videographer. this site is a selection of personal and commissioned work."

// allowed setting keys
const (
	keySiteName     = "site_name"
	keySiteSubtitle = "site_subtitle"
	keyInfoText     = "info_text"
	keyInfoLinks    = "info_links"
	keyLanguages    = "languages"
)

var settingsTranslatableFields = map[string]bool{
	keySiteName:     true,
	keySiteSubtitle: true,
	keyInfoText:     true,
}

var defaultLanguages = []string{"en"}

func (h *SettingsHandler) Get(c *gin.Context) {
	kv, err := h.store.GetSettings(c.Request.Context())
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to read settings", err)
		return
	}

	links := defaultInfoLinks
	if raw := kv[keyInfoLinks]; raw != "" {
		var parsed []infoLink
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			log.Printf("settings: ignoring invalid %s JSON: %v", keyInfoLinks, err)
		} else if parsed != nil {
			links = parsed
		}
	}

	languages := defaultLanguages
	if raw := kv[keyLanguages]; raw != "" {
		var parsed []string
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			log.Printf("settings: ignoring invalid %s JSON: %v", keyLanguages, err)
		} else if len(parsed) > 0 {
			languages = parsed
		}
	}

	trMap, err := h.store.LoadTranslations(c.Request.Context(), "settings", []int64{0})
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to read settings", err)
		return
	}
	translations := trMap[0]
	if translations == nil {
		translations = map[string]map[string]string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"site_name":     orDefault(kv[keySiteName], "your name"),
		"site_subtitle": orDefault(kv[keySiteSubtitle], "photographer & videographer"),
		"info_text":     orDefault(kv[keyInfoText], defaultInfoText),
		"info_links":    links,
		"languages":     languages,
		"translations":  translations,
	})
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

type updateSettingsRequest struct {
	SiteName     *string                      `json:"site_name"`
	SiteSubtitle *string                      `json:"site_subtitle"`
	InfoText     *string                      `json:"info_text"`
	InfoLinks    *[]infoLink                  `json:"info_links"`
	Languages    *[]string                    `json:"languages"`
	Translations map[string]map[string]string `json:"translations"`
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid request body")
		return
	}

	kv := make(map[string]string)
	if req.SiteName != nil {
		kv[keySiteName] = strings.TrimSpace(*req.SiteName)
	}
	if req.SiteSubtitle != nil {
		kv[keySiteSubtitle] = strings.TrimSpace(*req.SiteSubtitle)
	}
	if req.InfoText != nil {
		kv[keyInfoText] = *req.InfoText
	}
	if req.InfoLinks != nil {
		links := *req.InfoLinks
		if len(links) > 20 {
			fail(c, http.StatusBadRequest, "too many links")
			return
		}
		for i := range links {
			links[i].Label = strings.TrimSpace(links[i].Label)
			links[i].URL = strings.TrimSpace(links[i].URL)
			if links[i].Label == "" || links[i].URL == "" {
				fail(c, http.StatusBadRequest, "links need both label and url")
				return
			}
			u, err := url.Parse(links[i].URL)
			if err != nil {
				fail(c, http.StatusBadRequest, "invalid link url")
				return
			}
			switch strings.ToLower(u.Scheme) {
			case "http", "https", "mailto":
			default:
				fail(c, http.StatusBadRequest, "link urls must use http, https or mailto")
				return
			}
		}
		raw, err := json.Marshal(links)
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid links")
			return
		}
		kv[keyInfoLinks] = string(raw)
	}

	if req.Languages != nil {
		langs := *req.Languages
		if len(langs) == 0 || len(langs) > len(ValidLangs) {
			fail(c, http.StatusBadRequest, "languages must be a non-empty subset of en, ru, es, et, de")
			return
		}
		for _, l := range langs {
			if !ValidLangs[l] {
				fail(c, http.StatusBadRequest, "unsupported lang "+l)
				return
			}
		}
		raw, err := json.Marshal(langs)
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid languages")
			return
		}
		kv[keyLanguages] = string(raw)
	}

	if req.Translations != nil {
		if err := validateTranslations(req.Translations, settingsTranslatableFields); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		if err := h.store.SetTranslations(c.Request.Context(), "settings", 0, req.Translations); err != nil {
			failErr(c, http.StatusInternalServerError, "failed to save translations", err)
			return
		}
	}

	if len(kv) == 0 {
		h.Get(c)
		return
	}
	if err := h.store.SetSettings(c.Request.Context(), kv); err != nil {
		failErr(c, http.StatusInternalServerError, "failed to save settings", err)
		return
	}
	h.Get(c)
}
