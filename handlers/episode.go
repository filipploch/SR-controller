package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EpisodeHandler struct {
	DB *gorm.DB
}

func NewEpisodeHandler(db *gorm.DB) *EpisodeHandler {
	return &EpisodeHandler{DB: db}
}

// GetEpisodes - GET /api/episodes
func (h *EpisodeHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	var episodes []models.Episode

	// Filtruj po sezonie jeśli podano
	query := h.DB.Preload("Season")
	if seasonID := r.URL.Query().Get("season_id"); seasonID != "" {
		query = query.Where("season_id = ?", seasonID)
	}

	result := query.Order("episode_number DESC").Find(&episodes)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
}

// GetEpisode - GET /api/episodes/{id}
func (h *EpisodeHandler) GetEpisode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var episode models.Episode
	result := h.DB.Preload("Season").
		Preload("Staff.Staff.StaffType").
		Preload("Guests.Guest.GuestType").
		Preload("Reportages.Source").
		Preload("Media.Source").
		First(&episode, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Episode not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episode)
}

// CreateEpisode - POST /api/episodes
func (h *EpisodeHandler) CreateEpisode(w http.ResponseWriter, r *http.Request) {
	var episode models.Episode
	if err := json.NewDecoder(r.Body).Decode(&episode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Jeśli ma być aktualny, użyj funkcji pomocniczej
	if episode.IsCurrent {
		if err := models.CreateEpisodeAsCurrent(h.DB, &episode); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.DB.Create(&episode).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(episode)
}

// UpdateEpisode - PUT /api/episodes/{id}
func (h *EpisodeHandler) UpdateEpisode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var episode models.Episode
	if err := h.DB.First(&episode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Episode not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.Episode
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Jeśli ustawiamy jako aktualny
	if updateData.IsCurrent && !episode.IsCurrent {
		if err := models.SetCurrentEpisode(h.DB, uint(id)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.DB.First(&episode, id)
	} else {
		episode.SeasonID = updateData.SeasonID
		episode.EpisodeNumber = updateData.EpisodeNumber
		episode.SeasonEpisode = updateData.SeasonEpisode
		episode.Title = updateData.Title
		episode.EpisodeDate = updateData.EpisodeDate
		episode.IsCurrent = updateData.IsCurrent

		if err := h.DB.Save(&episode).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episode)
}

// DeleteEpisode - DELETE /api/episodes/{id}
func (h *EpisodeHandler) DeleteEpisode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var episode models.Episode
	if err := h.DB.First(&episode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Episode not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if episode.IsCurrent {
		http.Error(w, "Cannot delete current episode", http.StatusConflict)
		return
	}

	if err := h.DB.Delete(&episode).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetCurrentEpisode - POST /api/episodes/{id}/set-current
func (h *EpisodeHandler) SetCurrentEpisode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := models.SetCurrentEpisode(h.DB, uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Episode not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var episode models.Episode
	h.DB.Preload("Season").First(&episode, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episode)
}
