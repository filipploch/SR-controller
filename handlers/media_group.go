package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type MediaGroupHandler struct {
	DB *gorm.DB
}

func NewMediaGroupHandler(db *gorm.DB) *MediaGroupHandler {
	return &MediaGroupHandler{DB: db}
}

// GetMediaGroups - GET /api/media-groups
func (h *MediaGroupHandler) GetMediaGroups(w http.ResponseWriter, r *http.Request) {
	var groups []models.MediaGroup
	result := h.DB.Order("name ASC").Find(&groups)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// GetMediaGroup - GET /api/media-groups/{id}
func (h *MediaGroupHandler) GetMediaGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var group models.MediaGroup
	result := h.DB.First(&group, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Media group not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// CreateMediaGroup - POST /api/media-groups
func (h *MediaGroupHandler) CreateMediaGroup(w http.ResponseWriter, r *http.Request) {
	var group models.MediaGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.DB.Create(&group).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// UpdateMediaGroup - PUT /api/media-groups/{id}
func (h *MediaGroupHandler) UpdateMediaGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var group models.MediaGroup
	if err := h.DB.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Media group not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.MediaGroup
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group.Name = updateData.Name
	group.Description = updateData.Description

	if err := h.DB.Save(&group).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// DeleteMediaGroup - DELETE /api/media-groups/{id}
func (h *MediaGroupHandler) DeleteMediaGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Usuń najpierw przypisania
	h.DB.Where("media_group_id = ?", id).Delete(&models.EpisodeMediaGroup{})

	if err := h.DB.Delete(&models.MediaGroup{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMediaGroupItems - GET /api/media-groups/{id}/items
// Zwraca media przypisane do grupy
func (h *MediaGroupHandler) GetMediaGroupItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var items []models.EpisodeMediaGroup
	result := h.DB.Preload("EpisodeMedia.Scene").
		Preload("EpisodeMedia.EpisodeStaff.Staff").
		Where("media_group_id = ?", id).
		Order("episode_order ASC").
		Find(&items)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// AddMediaToGroup - POST /api/media-groups/{group_id}/media/{media_id}
func (h *MediaGroupHandler) AddMediaToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.ParseUint(vars["group_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.ParseUint(vars["media_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	var data struct {
		EpisodeOrder int `json:"episode_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy media istnieje
	var media models.EpisodeMedia
	if err := h.DB.First(&media, mediaID).Error; err != nil {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	// Sprawdź czy grupa istnieje
	var group models.MediaGroup
	if err := h.DB.First(&group, groupID).Error; err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Sprawdź czy już nie jest przypisane
	var existing models.EpisodeMediaGroup
	if err := h.DB.Where("episode_media_id = ? AND media_group_id = ?", mediaID, groupID).First(&existing).Error; err == nil {
		http.Error(w, "Media already in group", http.StatusConflict)
		return
	}

	assignment := models.EpisodeMediaGroup{
		EpisodeMediaID: uint(mediaID),
		MediaGroupID:   uint(groupID),
		EpisodeOrder:   data.EpisodeOrder,
		IsCurrent:      false,
	}

	if err := h.DB.Create(&assignment).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Załaduj relacje
	h.DB.Preload("EpisodeMedia").Preload("MediaGroup").First(&assignment, assignment.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assignment)
}

// RemoveMediaFromGroup - DELETE /api/media-groups/{group_id}/media/{media_id}
func (h *MediaGroupHandler) RemoveMediaFromGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.ParseUint(vars["group_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.ParseUint(vars["media_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	var assignment models.EpisodeMediaGroup
	if err := h.DB.Where("episode_media_id = ? AND media_group_id = ?", mediaID, groupID).First(&assignment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Assignment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := h.DB.Delete(&assignment).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetCurrentMediaGroup - POST /api/episodes/{episode_id}/media-groups/{group_id}/set-current
func (h *MediaGroupHandler) SetCurrentMediaGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseUint(vars["group_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	if err := models.SetCurrentMediaGroup(h.DB, uint(episodeID), uint(groupID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
