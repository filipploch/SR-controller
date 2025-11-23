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
	result := h.DB.Preload("Source.Scene").
		Preload("Staff.StaffType").
		Where("episode_id = ?", episodeID).
		Order("order ASC").
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
	h.DB.Preload("Source.Scene").Preload("Staff.StaffType").First(&media, media.ID)

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

	media.SourceID = updateData.SourceID
	media.StaffID = updateData.StaffID
	media.Title = updateData.Title
	media.Description = updateData.Description
	media.FilePath = updateData.FilePath
	media.URL = updateData.URL
	media.Order = updateData.Order

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

	h.DB.Preload("Source.Scene").Preload("Staff.StaffType").First(&media, media.ID)

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

	mediaType := r.FormValue("media_type") // "reportages", "videos", "images"
	if mediaType == "" {
		mediaType = "videos"
	}

	// Utwórz folder docelowy: media/{mediaType}/S{season}/E{episode}/
	seasonFolder := fmt.Sprintf("S%02d", episode.Season.Number)
	episodeFolder := fmt.Sprintf("E%03d", episode.SeasonEpisode)
	targetDir := filepath.Join(h.MediaPath, mediaType, seasonFolder, episodeFolder)

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
	relativePath := filepath.Join(mediaType, seasonFolder, episodeFolder, handler.Filename)

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

	seasonFolder := fmt.Sprintf("S%02d", episode.Season.Number)
	episodeFolder := fmt.Sprintf("E%03d", episode.SeasonEpisode)

	// Skanuj wszystkie typy mediów
	mediaTypes := []string{"reportages", "videos", "images", "audio"}
	var files []map[string]interface{}

	for _, mediaType := range mediaTypes {
		dirPath := filepath.Join(h.MediaPath, mediaType, seasonFolder, episodeFolder)

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue // Folder nie istnieje, pomijamy
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				relativePath := filepath.Join(mediaType, seasonFolder, episodeFolder, entry.Name())
				fullPath := filepath.Join(h.MediaPath, relativePath)

				duration, _ := utils.GetMediaDuration(fullPath)

				files = append(files, map[string]interface{}{
					"name":     entry.Name(),
					"path":     relativePath,
					"type":     mediaType,
					"duration": duration,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
