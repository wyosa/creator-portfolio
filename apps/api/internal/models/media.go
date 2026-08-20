package models

type Media struct {
	ID           int64  `json:"id"`
	Type         string `json:"type"`
	Source       string `json:"source"`
	Path         string `json:"path"`
	ExternalID   string `json:"external_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	InstagramURL string `json:"instagram_url"`
	YoutubeURL   string `json:"youtube_url"`
	VimeoURL     string `json:"vimeo_url"`
	PreviewPath  string `json:"preview_path"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Featured     bool   `json:"featured"`
	Position     int64  `json:"position"`
	CreatedAt    string `json:"created_at"`

	// Translations: field → lang → text (e.g. translations.title.ru)
	Translations map[string]map[string]string `json:"translations"`
}
