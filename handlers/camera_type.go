package handlers

import (
	"encoding/json"
	"net/http"
	"obs-controller/models"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type CameraTypeHandler struct {
	DB *gorm.DB
}

func NewCameraTypeHandler(db *gorm.DB) *CameraTypeHandler {
	return &CameraTypeHandler{DB: db}
}

// GetCameraTypes - GET /api/camera-types
func (h *CameraTypeHandler) GetCameraTypes(w http.ResponseWriter, r *http.Request) {
	var cameraTypes []models.CameraType
	result := h.DB.Order("\"order\" ASC").Find(&cameraTypes)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cameraTypes)
}

// GetCameraType - GET /api/camera-types/{id}
func (h *CameraTypeHandler) GetCameraType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var cameraType models.CameraType
	result := h.DB.First(&cameraType, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "CameraType not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cameraType)
}

// CreateCameraType - POST /api/camera-types
func (h *CameraTypeHandler) CreateCameraType(w http.ResponseWriter, r *http.Request) {
	var cameraType models.CameraType
	if err := json.NewDecoder(r.Body).Decode(&cameraType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy nazwa już istnieje
	var existing models.CameraType
	if err := h.DB.Where("name = ?", cameraType.Name).First(&existing).Error; err == nil {
		http.Error(w, "CameraType with this name already exists", http.StatusConflict)
		return
	}

	// Użytkownik nie może tworzyć systemowych typów
	cameraType.IsSystem = false

	// Ustaw kolejność jako ostatnią
	var maxOrder int
	h.DB.Model(&models.CameraType{}).Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder)
	cameraType.Order = maxOrder + 1

	if err := h.DB.Create(&cameraType).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cameraType)
}

// UpdateCameraType - PUT /api/camera-types/{id}
func (h *CameraTypeHandler) UpdateCameraType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var cameraType models.CameraType
	if err := h.DB.First(&cameraType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "CameraType not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Nie można edytować systemowych typów
	if cameraType.IsSystem {
		http.Error(w, "Cannot edit system camera type", http.StatusForbidden)
		return
	}

	var updateData models.CameraType
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy nazwa nie koliduje
	if updateData.Name != cameraType.Name {
		var existing models.CameraType
		if err := h.DB.Where("name = ? AND id != ?", updateData.Name, id).First(&existing).Error; err == nil {
			http.Error(w, "CameraType with this name already exists", http.StatusConflict)
			return
		}
	}

	cameraType.Name = updateData.Name

	if err := h.DB.Save(&cameraType).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cameraType)
}

// DeleteCameraType - DELETE /api/camera-types/{id}
func (h *CameraTypeHandler) DeleteCameraType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var cameraType models.CameraType
	if err := h.DB.First(&cameraType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "CameraType not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Nie można usuwać systemowych typów
	if cameraType.IsSystem {
		http.Error(w, "Cannot delete system camera type", http.StatusForbidden)
		return
	}

	// Sprawdź czy są przypisane kamery
	var count int64
	h.DB.Model(&models.EpisodeSource{}).Where("camera_type_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete CameraType with assigned cameras", http.StatusConflict)
		return
	}

	if err := h.DB.Delete(&cameraType).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
