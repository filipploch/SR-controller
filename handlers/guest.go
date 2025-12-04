package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type GuestHandler struct {
	DB *gorm.DB
}

func NewGuestHandler(db *gorm.DB) *GuestHandler {
	return &GuestHandler{DB: db}
}

// GetGuests - GET /api/guests
func (h *GuestHandler) GetGuests(w http.ResponseWriter, r *http.Request) {
	var guests []models.Guest

	query := h.DB.Preload("GuestType")

	// Filtruj po typie jeśli podano
	if typeID := r.URL.Query().Get("type_id"); typeID != "" {
		query = query.Where("guest_type_id = ?", typeID)
	}

	result := query.Order("last_name ASC, first_name ASC").Find(&guests)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guests)
}

// GetGuest - GET /api/guests/{id}
func (h *GuestHandler) GetGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var guest models.Guest
	result := h.DB.Preload("GuestType").First(&guest, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Guest not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guest)
}

// CreateGuest - POST /api/guests
func (h *GuestHandler) CreateGuest(w http.ResponseWriter, r *http.Request) {
	var guest models.Guest
	if err := json.NewDecoder(r.Body).Decode(&guest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy typ istnieje (tylko jeśli podano)
	if guest.GuestTypeID != nil {
		var guestType models.GuestType
		if err := h.DB.First(&guestType, *guest.GuestTypeID).Error; err != nil {
			http.Error(w, "GuestType not found", http.StatusBadRequest)
			return
		}
	}

	if err := h.DB.Create(&guest).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Załaduj relacje
	h.DB.Preload("GuestType").First(&guest, guest.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(guest)
}

// UpdateGuest - PUT /api/guests/{id}
func (h *GuestHandler) UpdateGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var guest models.Guest
	if err := h.DB.First(&guest, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Guest not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.Guest
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy typ istnieje (tylko jeśli podano)
	if updateData.GuestTypeID != nil {
		var guestType models.GuestType
		if err := h.DB.First(&guestType, *updateData.GuestTypeID).Error; err != nil {
			http.Error(w, "GuestType not found", http.StatusBadRequest)
			return
		}
	}

	guest.GuestTypeID = updateData.GuestTypeID
	guest.FirstName = updateData.FirstName
	guest.LastName = updateData.LastName

	if err := h.DB.Save(&guest).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Załaduj relacje
	h.DB.Preload("GuestType").First(&guest, guest.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guest)
}

// DeleteGuest - DELETE /api/guests/{id}
func (h *GuestHandler) DeleteGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Sprawdź czy jest przypisany do odcinków
	var count int64
	h.DB.Model(&models.EpisodeGuest{}).Where("guest_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete guest assigned to episodes", http.StatusConflict)
		return
	}

	if err := h.DB.Delete(&models.Guest{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
