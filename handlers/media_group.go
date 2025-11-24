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
	query := h.DB.Order("\"order\" ASC")

	// Filtruj po episode_id jeśli podano
	if episodeID := r.URL.Query().Get("episode_id"); episodeID != "" {
		query = query.Where("episode_id = ?", episodeID)
	}

	result := query.Find(&groups)

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

	// Automatycznie ustaw kolejność jeśli nie podano
	if group.Order == 0 {
		group.Order = models.GetNextMediaGroupOrder(h.DB, group.EpisodeID)
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
		Order("\"order\" ASC").
		Find(&items)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// AddItemToGroup - POST /api/media-groups/{id}/items
// Dodaje media do grupy przyjmując episode_media_id w body
func (h *MediaGroupHandler) AddItemToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var data struct {
		EpisodeMediaID uint `json:"episode_media_id"`
		Order          int  `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy media istnieje
	var media models.EpisodeMedia
	if err := h.DB.First(&media, data.EpisodeMediaID).Error; err != nil {
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
	if err := h.DB.Where("episode_media_id = ? AND media_group_id = ?", data.EpisodeMediaID, groupID).First(&existing).Error; err == nil {
		http.Error(w, "Media already in group", http.StatusConflict)
		return
	}

	// Automatycznie ustaw kolejność jeśli nie podano
	order := data.Order
	if order == 0 {
		order = models.GetNextMediaGroupItemOrder(h.DB, uint(groupID))
	}

	assignment := models.EpisodeMediaGroup{
		EpisodeMediaID: data.EpisodeMediaID,
		MediaGroupID:   uint(groupID),
		Order:          order,
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
		Order int `json:"order"`
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
		Order:          data.Order,
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

// ReorderMediaGroup - PUT /api/media-groups/{id}/reorder
func (h *MediaGroupHandler) ReorderMediaGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var data struct {
		Order int `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var group models.MediaGroup
		if err := tx.First(&group, groupID).Error; err != nil {
			return err
		}

		oldOrder := group.Order
		newOrder := data.Order

		if newOrder > oldOrder {
			tx.Model(&models.MediaGroup{}).
				Where("episode_id = ? AND \"order\" > ? AND \"order\" <= ?", group.EpisodeID, oldOrder, newOrder).
				UpdateColumn("order", gorm.Expr("\"order\" - 1"))
		} else if newOrder < oldOrder {
			tx.Model(&models.MediaGroup{}).
				Where("episode_id = ? AND \"order\" >= ? AND \"order\" < ?", group.EpisodeID, newOrder, oldOrder).
				UpdateColumn("order", gorm.Expr("\"order\" + 1"))
		}

		if err := tx.Model(&group).Update("order", newOrder).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ReorderMediaGroupItem - PUT /api/media-groups/{group_id}/items/{id}/reorder
func (h *MediaGroupHandler) ReorderMediaGroupItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.ParseUint(vars["group_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var data struct {
		Order int `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var item models.EpisodeMediaGroup
		if err := tx.First(&item, itemID).Error; err != nil {
			return err
		}

		oldOrder := item.Order
		newOrder := data.Order

		if newOrder > oldOrder {
			tx.Model(&models.EpisodeMediaGroup{}).
				Where("media_group_id = ? AND \"order\" > ? AND \"order\" <= ?", groupID, oldOrder, newOrder).
				UpdateColumn("order", gorm.Expr("\"order\" - 1"))
		} else if newOrder < oldOrder {
			tx.Model(&models.EpisodeMediaGroup{}).
				Where("media_group_id = ? AND \"order\" >= ? AND \"order\" < ?", groupID, newOrder, oldOrder).
				UpdateColumn("order", gorm.Expr("\"order\" + 1"))
		}

		if err := tx.Model(&item).Update("order", newOrder).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
