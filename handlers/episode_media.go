package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"obs-controller/models"
	"obs-controller/utils"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EpisodeMediaHandler struct {
	DB        *gorm.DB
	MediaPath string // Ścieżka bazowa do mediów
}

func NewEpisodeMediaHandler(db *gorm.DB, mediaPath string) *EpisodeMediaHandler {
	return &EpisodeMediaHandler{
		DB:        db,
		MediaPath: mediaPath,
	}
}

// GetEpisodeMedia - GET /api/episodes/{episode_id}/media
func (h *EpisodeMediaHandler) GetEpisodeMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	var media []models.EpisodeMedia
	result := h.DB.Preload("Scene").
		Preload("EpisodeStaff.Staff").
		Preload("EpisodeStaff.StaffTypes.StaffType").
		Preload("MediaGroups.MediaGroup").
		Where("episode_id = ?", episodeID).
		Order("created_at ASC").
		Find(&media)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}

// CreateEpisodeMedia - POST /api/episodes/{episode_id}/media
func (h *EpisodeMediaHandler) CreateEpisodeMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	var media models.EpisodeMedia
	if err := json.NewDecoder(r.Body).Decode(&media); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	media.EpisodeID = uint(episodeID)

	// Jeśli podano FilePath, odczytaj duration
	if media.FilePath != nil && *media.FilePath != "" {
		fullPath := filepath.Join(h.MediaPath, *media.FilePath)
		if duration, err := utils.GetMediaDuration(fullPath); err == nil {
			media.Duration = duration
		}
	}

	if err := h.DB.Create(&media).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Załaduj relacje
	h.DB.Preload("Scene").
		Preload("EpisodeStaff.Staff").
		Preload("EpisodeStaff.StaffTypes.StaffType").
		First(&media, media.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(media)
}

// UpdateEpisodeMedia - PUT /api/episodes/{episode_id}/media/{id}
func (h *EpisodeMediaHandler) UpdateEpisodeMedia(w http.ResponseWriter, r *http.Request) {
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

	var media models.EpisodeMedia
	if err := h.DB.Where("id = ? AND episode_id = ?", id, episodeID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Media not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData models.EpisodeMedia
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	media.SceneID = updateData.SceneID
	media.EpisodeStaffID = updateData.EpisodeStaffID
	media.Title = updateData.Title
	media.Description = updateData.Description
	media.FilePath = updateData.FilePath
	media.URL = updateData.URL

	// Jeśli zmieniono FilePath, odczytaj nowy duration
	if media.FilePath != nil && *media.FilePath != "" {
		fullPath := filepath.Join(h.MediaPath, *media.FilePath)
		if duration, err := utils.GetMediaDuration(fullPath); err == nil {
			media.Duration = duration
		}
	}

	if err := h.DB.Save(&media).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.DB.Preload("Scene").
		Preload("EpisodeStaff.Staff").
		Preload("EpisodeStaff.StaffTypes.StaffType").
		First(&media, media.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}

// DeleteEpisodeMedia - DELETE /api/episodes/{episode_id}/media/{id}
func (h *EpisodeMediaHandler) DeleteEpisodeMedia(w http.ResponseWriter, r *http.Request) {
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

	var media models.EpisodeMedia
	if err := h.DB.Where("id = ? AND episode_id = ?", id, episodeID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Media not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := h.DB.Delete(&media).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadMedia - POST /api/episodes/{episode_id}/media/upload
func (h *EpisodeMediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	// Pobierz episode aby znać sezon
	var episode models.Episode
	if err := h.DB.Preload("Season").First(&episode, episodeID).Error; err != nil {
		http.Error(w, "Episode not found", http.StatusNotFound)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100 MB max
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Utwórz folder docelowy: media/season_{number}/
	seasonFolder := fmt.Sprintf("season_%d", episode.Season.Number)
	targetDir := filepath.Join(h.MediaPath, seasonFolder)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	// Zapisz plik
	targetPath := filepath.Join(targetDir, handler.Filename)
	dst, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// Odczytaj duration
	duration, _ := utils.GetMediaDuration(targetPath)

	// Względna ścieżka od folderu media
	relativePath := filepath.Join(seasonFolder, handler.Filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"file_path": relativePath,
		"duration":  duration,
		"filename":  handler.Filename,
	})
}

// ListMediaFiles - GET /api/episodes/{episode_id}/media/files
func (h *EpisodeMediaHandler) ListMediaFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	// Pobierz episode aby znać sezon
	var episode models.Episode
	if err := h.DB.Preload("Season").First(&episode, episodeID).Error; err != nil {
		http.Error(w, "Episode not found", http.StatusNotFound)
		return
	}

	seasonFolder := fmt.Sprintf("season_%d", episode.Season.Number)
	dirPath := filepath.Join(h.MediaPath, seasonFolder)

	var files []map[string]interface{}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Folder nie istnieje, zwracamy pustą listę
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			relativePath := filepath.Join(seasonFolder, entry.Name())
			fullPath := filepath.Join(h.MediaPath, relativePath)

			duration, _ := utils.GetMediaDuration(fullPath)

			// Określ typ pliku na podstawie rozszerzenia
			ext := filepath.Ext(entry.Name())
			var fileType string
			switch ext {
			case ".mp4", ".avi", ".mov", ".mkv", ".webm":
				fileType = "video"
			case ".mp3", ".wav", ".flac", ".aac":
				fileType = "audio"
			case ".jpg", ".jpeg", ".png", ".gif", ".webp":
				fileType = "image"
			default:
				fileType = "other"
			}

			files = append(files, map[string]interface{}{
				"name":     entry.Name(),
				"path":     relativePath,
				"type":     fileType,
				"duration": duration,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// SetCurrentMedia - POST /api/episodes/{episode_id}/media/{id}/set-current
func (h *EpisodeMediaHandler) SetCurrentMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	if err := models.SetCurrentEpisodeMedia(h.DB, uint(episodeID), uint(mediaID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
