package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"

	"gorm.io/gorm"
)

type SettingsHandler struct {
	DB *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{DB: db}
}

type SettingsStatus struct {
	HasCurrentSeason    bool `json:"has_current_season"`
	HasCurrentEpisode   bool `json:"has_current_episode"`
	CanAccessController bool `json:"can_access_controller"`
}

// GetStatus zwraca status aplikacji - czy ma aktualne dane
func (h *SettingsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := SettingsStatus{}

	// Sprawdź czy jest aktualny sezon
	var season models.Season
	err := h.DB.Where("is_current = ?", true).First(&season).Error
	status.HasCurrentSeason = err == nil

	// Sprawdź czy jest aktualny odcinek
	var episode models.Episode
	err = h.DB.Where("is_current = ?", true).First(&episode).Error
	status.HasCurrentEpisode = err == nil

	// Kontroler dostępny tylko gdy są oba
	status.CanAccessController = status.HasCurrentSeason && status.HasCurrentEpisode

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
