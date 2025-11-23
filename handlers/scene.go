package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"

	"gorm.io/gorm"
)

type SceneHandler struct {
	DB *gorm.DB
}

func NewSceneHandler(db *gorm.DB) *SceneHandler {
	return &SceneHandler{DB: db}
}

// GetScenes - GET /api/scenes
func (h *SceneHandler) GetScenes(w http.ResponseWriter, r *http.Request) {
	var scenes []models.Scene
	result := h.DB.Preload("Sources").Find(&scenes)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scenes)
}

// GetMediaScenes - GET /api/scenes/media
// Zwraca tylko sceny MEDIA i REPORTAZE
func (h *SceneHandler) GetMediaScenes(w http.ResponseWriter, r *http.Request) {
	scenes, err := models.GetMediaScenes(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scenes)
}
