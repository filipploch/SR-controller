package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EpisodeStaffHandler struct {
	DB *gorm.DB
}

func NewEpisodeStaffHandler(db *gorm.DB) *EpisodeStaffHandler {
	return &EpisodeStaffHandler{DB: db}
}

// GetEpisodeStaff - GET /api/episodes/{episode_id}/staff
func (h *EpisodeStaffHandler) GetEpisodeStaff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	var episodeStaff []models.EpisodeStaff
	result := h.DB.Preload("Staff").
		Preload("StaffTypes.StaffType").
		Where("episode_id = ?", episodeID).
		Find(&episodeStaff)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodeStaff)
}

// AddStaffToEpisode - POST /api/episodes/{episode_id}/staff
func (h *EpisodeStaffHandler) AddStaffToEpisode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	var data struct {
		StaffID      uint   `json:"staff_id"`
		StaffTypeIDs []uint `json:"staff_type_ids"` // Lista typów dla tego przypisania
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

	// Sprawdź czy osoba istnieje
	var staff models.Staff
	if err := h.DB.First(&staff, data.StaffID).Error; err != nil {
		http.Error(w, "Staff member not found", http.StatusNotFound)
		return
	}

	// Sprawdź czy już nie jest przypisana
	var existing models.EpisodeStaff
	if err := h.DB.Where("episode_id = ? AND staff_id = ?", episodeID, data.StaffID).First(&existing).Error; err == nil {
		http.Error(w, "Staff member already assigned to this episode", http.StatusConflict)
		return
	}

	// Utwórz przypisanie
	episodeStaff := models.EpisodeStaff{
		EpisodeID: uint(episodeID),
		StaffID:   data.StaffID,
	}

	if err := h.DB.Create(&episodeStaff).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Dodaj typy jeśli podano
	if len(data.StaffTypeIDs) > 0 {
		for _, typeID := range data.StaffTypeIDs {
			// Sprawdź czy typ istnieje
			var staffType models.StaffType
			if err := h.DB.First(&staffType, typeID).Error; err != nil {
				continue // Pomijamy nieistniejące typy
			}

			episodeStaffType := models.EpisodeStaffType{
				EpisodeStaffID: episodeStaff.ID,
				StaffTypeID:    typeID,
			}
			h.DB.Create(&episodeStaffType)
		}
	}

	// Załaduj relacje
	h.DB.Preload("Staff").Preload("StaffTypes.StaffType").First(&episodeStaff, episodeStaff.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(episodeStaff)
}

// RemoveStaffFromEpisode - DELETE /api/episodes/{episode_id}/staff/{id}
func (h *EpisodeStaffHandler) RemoveStaffFromEpisode(w http.ResponseWriter, r *http.Request) {
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

	var episodeStaff models.EpisodeStaff
	if err := h.DB.Where("id = ? AND episode_id = ?", id, episodeID).First(&episodeStaff).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Assignment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := h.DB.Delete(&episodeStaff).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateEpisodeStaffTypes - PUT /api/episodes/{episode_id}/staff/{id}/types
func (h *EpisodeStaffHandler) UpdateEpisodeStaffTypes(w http.ResponseWriter, r *http.Request) {
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

	var episodeStaff models.EpisodeStaff
	if err := h.DB.Where("id = ? AND episode_id = ?", id, episodeID).First(&episodeStaff).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Assignment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var data struct {
		StaffTypeIDs []uint `json:"staff_type_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Usuń stare typy
	h.DB.Where("episode_staff_id = ?", id).Delete(&models.EpisodeStaffType{})

	// Dodaj nowe typy
	for _, typeID := range data.StaffTypeIDs {
		// Sprawdź czy typ istnieje
		var staffType models.StaffType
		if err := h.DB.First(&staffType, typeID).Error; err != nil {
			continue // Pomijamy nieistniejące typy
		}

		episodeStaffType := models.EpisodeStaffType{
			EpisodeStaffID: episodeStaff.ID,
			StaffTypeID:    typeID,
		}
		h.DB.Create(&episodeStaffType)
	}

	// Załaduj relacje
	h.DB.Preload("Staff").Preload("StaffTypes.StaffType").First(&episodeStaff, episodeStaff.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodeStaff)
}
