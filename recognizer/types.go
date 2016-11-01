package recognizer

import (
	"time"
)

type inference struct {
	ID    int     `json:"id"`
	Score float32 `json:"score"`
	Face  *face   `json:"face"`
	Label *label  `json:"label"`
}

type face struct {
	ID       int    `json:"id"`
	ImageURL string `json:"image_url"`
	Photo    *photo `json:"photo"`
}

type photo struct {
	SourceURL string     `json:"source_url"`
	PhotoURL  string     `json:"photo_url"`
	Caption   string     `json:"caption"`
	PostedAt  *time.Time `json:"posted_at"`
}

type label struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Twitter     string `json:"twitter"`
}
