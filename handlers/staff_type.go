package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type StaffTypeHandler struct {
	DB *gorm.DB
}

func NewStaffTypeHandler(db *gorm.DB) *StaffTypeHandler {
	return &StaffTypeHandler{DB: db}
}

// GetStaffTypes - GET /api/staff-types
func (h *StaffTypeHandler) GetStaffTypes(w http.ResponseWriter, r *http.Request) {
	var staffTypes []models.StaffType
	result := h.DB.Order("name ASC").Find(&staffTypes)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staffTypes)
}

// GetStaffType - GET /api/staff-types/{id}
func (h *StaffTypeHandler) GetStaffType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var staffType models.StaffType
	result := h.DB.Preload("Staff").First(&staffType, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "StaffType not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staffType)
}

// CreateStaffType - POST /api/staff-types
func (h *StaffTypeHandler) CreateStaffType(w http.ResponseWriter, r *http.Request) {
	var staffType models.StaffType
	if err := json.NewDecoder(r.Body).Decode(&staffType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy nazwa już istnieje
	var existing models.StaffType
	if err := h.DB.Where("name = ?", staffType.Name).First(&existing).Error; err == nil {
		http.Error(w, "StaffType with this name already exists", http.StatusConflict)
		return
	}

	if err := h.DB.Create(&staffType).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(staffType)
}

// UpdateStaffType - PUT /api/staff-types/{id}
func (h *StaffTypeHandler) UpdateStaffType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var staffType models.StaffType
	if err := h.DB.First(&staffType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "StaffType not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.StaffType
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy nazwa nie koliduje
	if updateData.Name != staffType.Name {
		var existing models.StaffType
		if err := h.DB.Where("name = ? AND id != ?", updateData.Name, id).First(&existing).Error; err == nil {
			http.Error(w, "StaffType with this name already exists", http.StatusConflict)
			return
		}
	}

	staffType.Name = updateData.Name

	if err := h.DB.Save(&staffType).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staffType)
}

// DeleteStaffType - DELETE /api/staff-types/{id}
func (h *StaffTypeHandler) DeleteStaffType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Sprawdź czy są przypisani pracownicy
	var count int64
	h.DB.Model(&models.Staff{}).Where("staff_type_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete StaffType with assigned staff members", http.StatusConflict)
		return
	}

	if err := h.DB.Delete(&models.StaffType{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
