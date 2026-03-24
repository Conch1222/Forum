package domain

import (
	"strconv"
	"time"
)

type PostSearchDocument struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	ViewCount int       `json:"view_count"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewPostSearchDoc(p Post) PostSearchDocument {
	return PostSearchDocument{
		ID:        strconv.FormatUint(uint64(p.ID), 10),
		UserID:    strconv.FormatUint(uint64(p.UserID), 10),
		Title:     p.Title,
		Content:   p.Content,
		Status:    p.Status,
		ViewCount: p.ViewCount,
		LikeCount: p.LikeCount,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

type EsPostSearchResponse struct {
	Hits struct {
		Total struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				ID        string    `json:"id"`
				UserID    string    `json:"user_id"`
				Title     string    `json:"title"`
				Content   string    `json:"content"`
				Status    string    `json:"status"`
				LikeCount int       `json:"like_count"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
