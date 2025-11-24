package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type SeasonHandler struct {
	DB *gorm.DB
}

func NewSeasonHandler(db *gorm.DB) *SeasonHandler {
	return &SeasonHandler{DB: db}
}

// GetSeasons - GET /api/seasons
func (h *SeasonHandler) GetSeasons(w http.ResponseWriter, r *http.Request) {
	var seasons []models.Season
	result := h.DB.Order("number DESC").Find(&seasons)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seasons)
}

// GetSeason - GET /api/seasons/{id}
func (h *SeasonHandler) GetSeason(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var season models.Season
	result := h.DB.Preload("Episodes").First(&season, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Season not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

// CreateSeason - POST /api/seasons
func (h *SeasonHandler) CreateSeason(w http.ResponseWriter, r *http.Request) {
	var season models.Season
	if err := json.NewDecoder(r.Body).Decode(&season); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy numer sezonu już istnieje
	var existing models.Season
	if err := h.DB.Where("number = ?", season.Number).First(&existing).Error; err == nil {
		http.Error(w, "Season with this number already exists", http.StatusConflict)
		return
	}

	// Jeśli ma być aktualny, użyj funkcji pomocniczej
	if season.IsCurrent {
		if err := models.CreateSeasonAsCurrent(h.DB, &season); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.DB.Create(&season).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(season)
}

// UpdateSeason - PUT /api/seasons/{id}
func (h *SeasonHandler) UpdateSeason(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var season models.Season
	if err := h.DB.First(&season, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Season not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.Season
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy numer sezonu nie koliduje
	if updateData.Number != season.Number {
		var existing models.Season
		if err := h.DB.Where("number = ? AND id != ?", updateData.Number, id).First(&existing).Error; err == nil {
			http.Error(w, "Season with this number already exists", http.StatusConflict)
			return
		}
	}

	// Jeśli ustawiamy jako aktualny
	if updateData.IsCurrent && !season.IsCurrent {
		if err := models.SetCurrentSeason(h.DB, uint(id)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Odśwież dane
		h.DB.First(&season, id)
	} else {
		season.Number = updateData.Number
		season.Description = updateData.Description
		season.IsCurrent = updateData.IsCurrent

		if err := h.DB.Save(&season).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

// DeleteSeason - DELETE /api/seasons/{id}
func (h *SeasonHandler) DeleteSeason(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var season models.Season
	if err := h.DB.First(&season, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Season not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Nie pozwól usunąć aktualnego sezonu
	if season.IsCurrent {
		http.Error(w, "Cannot delete current season", http.StatusConflict)
		return
	}

	if err := h.DB.Delete(&season).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetCurrentSeason - POST /api/seasons/{id}/set-current
func (h *SeasonHandler) SetCurrentSeason(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := models.SetCurrentSeason(h.DB, uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Season not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var season models.Season
	h.DB.First(&season, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

// GetNextSeasonNumber - GET /api/seasons/next-number
func (h *SeasonHandler) GetNextSeasonNumber(w http.ResponseWriter, r *http.Request) {
	var maxSeason models.Season
	result := h.DB.Order("number DESC").First(&maxSeason)

	nextNumber := 1
	if result.Error == nil {
		nextNumber = maxSeason.Number + 1
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"next_number": nextNumber})
}
