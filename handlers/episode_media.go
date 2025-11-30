package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"obs-controller/models"
	"obs-controller/obsws"
	"obs-controller/utils"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EpisodeMediaHandler struct {
	DB        *gorm.DB
	MediaPath string        // Ścieżka bazowa do mediów
	OBSClient *obsws.Client // Klient OBS-WebSocket
}

func NewEpisodeMediaHandler(db *gorm.DB, mediaPath string, obsClient *obsws.Client) *EpisodeMediaHandler {
	return &EpisodeMediaHandler{
		DB:        db,
		MediaPath: mediaPath,
		OBSClient: obsClient,
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
	result := h.DB.Preload("EpisodeStaff.Staff").
		Preload("EpisodeStaff.StaffTypes.StaffType").
		Preload("MediaGroups.MediaGroup").
		Preload("MediaGroups.CurrentScene").
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

	// Sprawdź czy ten sam plik nie jest już przypisany do tego odcinka
	if media.FilePath != nil && *media.FilePath != "" {
		var existingMedia models.EpisodeMedia
		result := h.DB.Where("episode_id = ? AND file_path = ?", episodeID, *media.FilePath).First(&existingMedia)
		if result.Error == nil {
			// Znaleziono duplikat
			http.Error(w, "Ten plik jest już przypisany do tego odcinka", http.StatusConflict)
			return
		}
	}

	// Jeśli podano FilePath, odczytaj duration
	if media.FilePath != nil && *media.FilePath != "" {
		// Konwertuj ścieżkę z bazy (zawsze /) na format systemu operacyjnego
		fullPath := filepath.Join(h.MediaPath, filepath.FromSlash(*media.FilePath))
		if duration, err := utils.GetMediaDuration(fullPath); err == nil {
			media.Duration = duration
		}
	}

	if err := h.DB.Create(&media).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Załaduj relacje
	h.DB.Preload("EpisodeStaff.Staff").
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

	media.EpisodeStaffID = updateData.EpisodeStaffID
	media.Title = updateData.Title
	media.Description = updateData.Description
	media.FilePath = updateData.FilePath
	media.URL = updateData.URL

	// Jeśli zmieniono FilePath, odczytaj nowy duration
	if media.FilePath != nil && *media.FilePath != "" {
		// Konwertuj ścieżkę z bazy (zawsze /) na format systemu operacyjnego
		fullPath := filepath.Join(h.MediaPath, filepath.FromSlash(*media.FilePath))
		if duration, err := utils.GetMediaDuration(fullPath); err == nil {
			media.Duration = duration
		}
	}

	if err := h.DB.Save(&media).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.DB.Preload("EpisodeStaff.Staff").
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

	// Usuń najpierw wszystkie przypisania do grup
	h.DB.Where("episode_media_id = ?", id).Delete(&models.EpisodeMediaGroup{})

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

	// Względna ścieżka od folderu media - używamy / dla bazy danych
	relativePath := seasonFolder + "/" + handler.Filename

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
			// Względna ścieżka - używamy / dla bazy danych
			relativePath := seasonFolder + "/" + entry.Name()
			// Konwertuj na format systemu operacyjnego dla dostępu do pliku
			fullPath := filepath.Join(h.MediaPath, filepath.FromSlash(relativePath))

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

// GetCurrentMediaForScene - GET /api/episodes/current/media/scene/{scene_name}
// Pobiera aktywny plik media dla danej sceny z aktualnego odcinka
// Zwraca plik z current_in_scene dla grupy MEDIA/REPORTAZE
// Jeśli nie ma aktywnego, zwraca plik z najniższym order w tej grupie
func (h *EpisodeMediaHandler) GetCurrentMediaForScene(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sceneName := vars["scene_name"]

	// Pobierz aktualny odcinek
	currentEpisode, err := models.GetCurrentEpisode(h.DB)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No current episode",
		})
		return
	}

	// Pobierz scenę po nazwie
	scene, err := models.GetMediaSceneByName(h.DB, sceneName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Scene not found",
		})
		return
	}

	// Określ nazwę grupy na podstawie nazwy sceny
	var groupName string
	if sceneName == "MEDIA" {
		groupName = "MEDIA"
	} else if sceneName == "REPORTAZE" {
		groupName = "REPORTAZE"
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid scene name",
		})
		return
	}

	// Znajdź grupę o tej nazwie dla tego odcinka
	var mediaGroup models.MediaGroup
	result := h.DB.Where("episode_id = ? AND name = ?", currentEpisode.ID, groupName).First(&mediaGroup)
	if result.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Group not found",
		})
		return
	}

	// 1. Najpierw szukaj przypisania które jest aktywne (current_in_scene = scene.ID lub 0)
	var assignment models.EpisodeMediaGroup
	result = h.DB.Preload("EpisodeMedia").
		Preload("MediaGroup").
		Where("media_group_id = ? AND (current_in_scene = ? OR current_in_scene = 0)",
			mediaGroup.ID, scene.ID).
		Order("\"order\" ASC").
		First(&assignment)

	// 2. Jeśli nie znaleziono aktywnego, weź plik z najniższym order
	if result.Error == gorm.ErrRecordNotFound {
		result = h.DB.Preload("EpisodeMedia").
			Preload("MediaGroup").
			Where("media_group_id = ?", mediaGroup.ID).
			Order("\"order\" ASC").
			First(&assignment)
	}

	if result.Error == gorm.ErrRecordNotFound {
		// Brak mediów w tej grupie
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No media in this group",
		})
		return
	}

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	currentMedia := assignment.EpisodeMedia

	// Jeśli mamy plik i OBS jest połączony, ustaw go w źródle
	if currentMedia.FilePath != nil && *currentMedia.FilePath != "" && h.OBSClient != nil && h.OBSClient.IsConnected() {
		// Pobierz bezwzględną ścieżkę do katalogu aplikacji
		absMediaPath, err := filepath.Abs(h.MediaPath)
		if err != nil {
			absMediaPath = h.MediaPath
		}

		// Zbuduj pełną ścieżkę: C:/Users/.../media/season_1/file.mp4
		fullPath := filepath.Join(absMediaPath, filepath.FromSlash(*currentMedia.FilePath))

		playlist := make([]map[string]interface{}, 0)
		playlist = append(playlist, map[string]interface{}{
			"value": fullPath,
		})
		// Określ nazwę źródła w OBS na podstawie nazwy sceny
		var inputName string
		if sceneName == "MEDIA" {
			inputName = "Media1"
		} else if sceneName == "REPORTAZE" {
			inputName = "Reportaze1"
		}

		if inputName != "" {
			// Ustaw ustawienia źródła w OBS
			err = h.OBSClient.SetInputSettings(inputName, map[string]interface{}{
				"playlist": playlist,
				"loop":     false,
				"shuffle":  false,
			})

			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to set OBS playlist: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"media_id":  currentMedia.ID,
		"title":     currentMedia.Title,
		"file_path": currentMedia.FilePath,
		"url":       currentMedia.URL,
		"duration":  currentMedia.Duration,
		"group":     assignment.MediaGroup.Name,
	})
}
