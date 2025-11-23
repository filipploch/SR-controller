package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EpisodeGuestHandler struct {
	DB *gorm.DB
}

func NewEpisodeGuestHandler(db *gorm.DB) *EpisodeGuestHandler {
	return &EpisodeGuestHandler{DB: db}
}

// GetEpisodeGuests - GET /api/episodes/{episode_id}/guests
func (h *EpisodeGuestHandler) GetEpisodeGuests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	var episodeGuests []models.EpisodeGuest
	result := h.DB.Preload("Guest.GuestType").
		Where("episode_id = ?", episodeID).
		Order("segment_order ASC").
		Find(&episodeGuests)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodeGuests)
}

// AddGuestToEpisode - POST /api/episodes/{episode_id}/guests
func (h *EpisodeGuestHandler) AddGuestToEpisode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	var data struct {
		GuestID      uint   `json:"guest_id"`
		Topic        string `json:"topic"`
		SegmentOrder int    `json:"segment_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy odcinek istnieje
	var episode models.Episode
	if err := h.DB.First(&episode, episodeID).Error; err != nil {
		http.Error(w, "Episode not found", http.StatusNotFound)
		return
	}

	// Sprawdź czy gość istnieje
	var guest models.Guest
	if err := h.DB.First(&guest, data.GuestID).Error; err != nil {
		http.Error(w, "Guest not found", http.StatusNotFound)
		return
	}

	// Sprawdź czy już nie jest przypisany
	var existing models.EpisodeGuest
	if err := h.DB.Where("episode_id = ? AND guest_id = ?", episodeID, data.GuestID).First(&existing).Error; err == nil {
		http.Error(w, "Guest already assigned to this episode", http.StatusConflict)
		return
	}

	episodeGuest := models.EpisodeGuest{
		EpisodeID:    uint(episodeID),
		GuestID:      data.GuestID,
		Topic:        data.Topic,
		SegmentOrder: data.SegmentOrder,
	}

	if err := h.DB.Create(&episodeGuest).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Załaduj relacje
	h.DB.Preload("Guest.GuestType").First(&episodeGuest, episodeGuest.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(episodeGuest)
}

// UpdateEpisodeGuest - PUT /api/episodes/{episode_id}/guests/{id}
func (h *EpisodeGuestHandler) UpdateEpisodeGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var episodeGuest models.EpisodeGuest
	if err := h.DB.Where("id = ? AND episode_id = ?", id, episodeID).First(&episodeGuest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Assignment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData struct {
		Topic        string `json:"topic"`
		SegmentOrder int    `json:"segment_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	episodeGuest.Topic = updateData.Topic
	episodeGuest.SegmentOrder = updateData.SegmentOrder

	if err := h.DB.Save(&episodeGuest).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Załaduj relacje
	h.DB.Preload("Guest.GuestType").First(&episodeGuest, episodeGuest.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodeGuest)
}

// RemoveGuestFromEpisode - DELETE /api/episodes/{episode_id}/guests/{id}
func (h *EpisodeGuestHandler) RemoveGuestFromEpisode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var episodeGuest models.EpisodeGuest
	if err := h.DB.Where("id = ? AND episode_id = ?", id, episodeID).First(&episodeGuest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Assignment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := h.DB.Delete(&episodeGuest).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
