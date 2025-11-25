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

	// Sprawdź czy ten sam plik nie jest już przypisany do tego odcinka I tej sceny
	if media.FilePath != nil && *media.FilePath != "" {
		var existingMedia models.EpisodeMedia
		result := h.DB.Where("episode_id = ? AND scene_id = ? AND file_path = ?", episodeID, media.SceneID, *media.FilePath).First(&existingMedia)
		if result.Error == nil {
			// Znaleziono duplikat - ten sam plik w tej samej scenie
			http.Error(w, "Ten plik jest już przypisany do tej sceny w tym odcinku", http.StatusConflict)
			return
		}
	}

	// Automatycznie ustaw kolejność jeśli nie podano
	if media.Order == 0 {
		media.Order = models.GetNextEpisodeMediaOrder(h.DB, uint(episodeID), media.SceneID)
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
	media.Order = updateData.Order

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

// ReorderEpisodeMedia - PUT /api/episodes/{episode_id}/media/{id}/reorder
func (h *EpisodeMediaHandler) ReorderEpisodeMedia(w http.ResponseWriter, r *http.Request) {
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

	var data struct {
		Order int `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var media models.EpisodeMedia
		if err := tx.First(&media, mediaID).Error; err != nil {
			return err
		}

		oldOrder := media.Order
		newOrder := data.Order

		if newOrder > oldOrder {
			// Przesuń w dół elementy między starą a nową pozycją
			tx.Model(&models.EpisodeMedia{}).
				Where("episode_id = ? AND \"order\" > ? AND \"order\" <= ?", episodeID, oldOrder, newOrder).
				UpdateColumn("order", gorm.Expr("\"order\" - 1"))
		} else if newOrder < oldOrder {
			// Przesuń w górę elementy między nową a starą pozycją
			tx.Model(&models.EpisodeMedia{}).
				Where("episode_id = ? AND \"order\" >= ? AND \"order\" < ?", episodeID, newOrder, oldOrder).
				UpdateColumn("order", gorm.Expr("\"order\" + 1"))
		}

		// Ustaw nową kolejność
		if err := tx.Model(&media).Update("order", newOrder).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
	fmt.Println(dirPath)
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

// GetCurrentMediaForScene - GET /api/episodes/current/media/scene/{scene_name}
// Pobiera aktualny plik media dla danej sceny z aktualnego odcinka i ustawia go w OBS
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

	// Szukaj media z is_current=true dla tego odcinka i sceny
	var currentMedia models.EpisodeMedia
	result := h.DB.Where("episode_id = ? AND scene_id = ? AND is_current = ?",
		currentEpisode.ID, scene.ID, true).First(&currentMedia)

	if result.Error == gorm.ErrRecordNotFound {
		// Nie znaleziono media z is_current=true, szukaj z najniższym order
		result = h.DB.Where("episode_id = ? AND scene_id = ?",
			currentEpisode.ID, scene.ID).
			Order("\"order\" ASC").
			First(&currentMedia)

		if result.Error == gorm.ErrRecordNotFound {
			// Brak jakichkolwiek mediów dla tego odcinka i sceny
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "No media found",
			})
			return
		}

		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Ustaw to media jako current
		if err := models.SetCurrentEpisodeMedia(h.DB, currentEpisode.ID, currentMedia.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		currentMedia.IsCurrent = true
	}

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Jeśli mamy plik i OBS jest połączony, ustaw go w źródle
	if currentMedia.FilePath != nil && *currentMedia.FilePath != "" && h.OBSClient != nil && h.OBSClient.IsConnected() {
		// Pobierz bezwzględną ścieżkę do katalogu aplikacji
		absMediaPath, err := filepath.Abs(h.MediaPath)
		if err != nil {
			absMediaPath = h.MediaPath
		}

		// Zbuduj pełną ścieżkę: C:/Users/.../media/season_1/file.mp4
		fullPath := filepath.Join(absMediaPath, filepath.FromSlash(*currentMedia.FilePath))

		// Pobierz źródło z bazy dla tej sceny
		// Szukamy pierwszego źródła typu "Media Source" w tej scenie
		var source models.Source
		result := h.DB.Where("scene_id = ?", scene.ID).
			Order("source_order ASC").
			First(&source)

		if result.Error == nil && source.Name != "" {
			// Ustaw ustawienia źródła w OBS
			err = h.OBSClient.SetInputSettings(source.Name, map[string]interface{}{
				"local_file":          fullPath,
				"clear_on_media_end":  false,
				"close_when_inactive": true,
			})

			if err != nil {
				// Loguj błąd, ale nie przerywaj - zwróć dane mimo błędu OBS
				fmt.Printf("Błąd ustawiania pliku w OBS: %v\n", err)
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
	})
}
