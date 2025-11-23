package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type GuestTypeHandler struct {
	DB *gorm.DB
}

func NewGuestTypeHandler(db *gorm.DB) *GuestTypeHandler {
	return &GuestTypeHandler{DB: db}
}

// GetGuestTypes - GET /api/guest-types
func (h *GuestTypeHandler) GetGuestTypes(w http.ResponseWriter, r *http.Request) {
	var guestTypes []models.GuestType
	result := h.DB.Order("name ASC").Find(&guestTypes)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guestTypes)
}

// GetGuestType - GET /api/guest-types/{id}
func (h *GuestTypeHandler) GetGuestType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var guestType models.GuestType
	result := h.DB.Preload("Guests").First(&guestType, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "GuestType not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guestType)
}

// CreateGuestType - POST /api/guest-types
func (h *GuestTypeHandler) CreateGuestType(w http.ResponseWriter, r *http.Request) {
	var guestType models.GuestType
	if err := json.NewDecoder(r.Body).Decode(&guestType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy nazwa już istnieje
	var existing models.GuestType
	if err := h.DB.Where("name = ?", guestType.Name).First(&existing).Error; err == nil {
		http.Error(w, "GuestType with this name already exists", http.StatusConflict)
		return
	}

	if err := h.DB.Create(&guestType).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(guestType)
}

// UpdateGuestType - PUT /api/guest-types/{id}
func (h *GuestTypeHandler) UpdateGuestType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var guestType models.GuestType
	if err := h.DB.First(&guestType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "GuestType not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.GuestType
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy nazwa nie koliduje
	if updateData.Name != guestType.Name {
		var existing models.GuestType
		if err := h.DB.Where("name = ? AND id != ?", updateData.Name, id).First(&existing).Error; err == nil {
			http.Error(w, "GuestType with this name already exists", http.StatusConflict)
			return
		}
	}

	guestType.Name = updateData.Name

	if err := h.DB.Save(&guestType).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guestType)
}

// DeleteGuestType - DELETE /api/guest-types/{id}
func (h *GuestTypeHandler) DeleteGuestType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Sprawdź czy są przypisani goście
	var count int64
	h.DB.Model(&models.Guest{}).Where("guest_type_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete GuestType with assigned guests", http.StatusConflict)
		return
	}

	if err := h.DB.Delete(&models.GuestType{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
