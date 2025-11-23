package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type StaffHandler struct {
	DB *gorm.DB
}

func NewStaffHandler(db *gorm.DB) *StaffHandler {
	return &StaffHandler{DB: db}
}

// GetStaff - GET /api/staff
func (h *StaffHandler) GetStaff(w http.ResponseWriter, r *http.Request) {
	var staff []models.Staff
	result := h.DB.Order("last_name ASC, first_name ASC").Find(&staff)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staff)
}

// GetStaffMember - GET /api/staff/{id}
func (h *StaffHandler) GetStaffMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var staff models.Staff
	result := h.DB.First(&staff, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Staff member not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staff)
}

// CreateStaff - POST /api/staff
func (h *StaffHandler) CreateStaff(w http.ResponseWriter, r *http.Request) {
	var staff models.Staff
	if err := json.NewDecoder(r.Body).Decode(&staff); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.DB.Create(&staff).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(staff)
}

// UpdateStaff - PUT /api/staff/{id}
func (h *StaffHandler) UpdateStaff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var staff models.Staff
	if err := h.DB.First(&staff, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Staff member not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.Staff
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	staff.FirstName = updateData.FirstName
	staff.LastName = updateData.LastName

	if err := h.DB.Save(&staff).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staff)
}

// DeleteStaff - DELETE /api/staff/{id}
func (h *StaffHandler) DeleteStaff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Sprawdź czy jest przypisany do odcinków
	var count int64
	h.DB.Model(&models.EpisodeStaff{}).Where("staff_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete staff member assigned to episodes", http.StatusConflict)
		return
	}

	if err := h.DB.Delete(&models.Staff{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
