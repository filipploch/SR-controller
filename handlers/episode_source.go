package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"obs-controller/models"
	"obs-controller/obsws"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EpisodeSourceHandler struct {
	DB            *gorm.DB
	OBSClient     *obsws.Client
	MediaPath     string
	SocketHandler *SocketHandler
}

func NewEpisodeSourceHandler(db *gorm.DB, obsClient *obsws.Client, mediaPath string, socketHandler *SocketHandler) *EpisodeSourceHandler {
	return &EpisodeSourceHandler{
		DB:            db,
		OBSClient:     obsClient,
		MediaPath:     mediaPath,
		SocketHandler: socketHandler,
	}
}

// AssignMediaToSource - POST /api/episodes/{episode_id}/sources/{source_name}/assign-media
// Przypisuje konkretny plik media do źródła Media Source (Media1 lub Reportaze1)
func (h *EpisodeSourceHandler) AssignMediaToSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	var data struct {
		MediaID uint `json:"media_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Sprawdź czy media należy do tego odcinka
	var media models.EpisodeMedia
	if err := h.DB.Where("id = ? AND episode_id = ?", data.MediaID, episodeID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Media not found or doesn't belong to this episode", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Sprawdź czy plik istnieje
	if media.FilePath == nil || *media.FilePath == "" {
		http.Error(w, "Media has no file path", http.StatusBadRequest)
		return
	}

	// Wczytaj plik do OBS (jeśli połączony)
	if h.OBSClient != nil && h.OBSClient.IsConnected() {
		// Pobierz bezwzględną ścieżkę
		absMediaPath, err := filepath.Abs(h.MediaPath)
		if err != nil {
			absMediaPath = h.MediaPath
		}

		// Zbuduj pełną ścieżkę
		fullPath := filepath.Join(absMediaPath, filepath.FromSlash(*media.FilePath))

		// Ustaw plik w źródle Media Source
		err = h.OBSClient.SetInputSettings(sourceName, map[string]interface{}{
			"local_file":          fullPath,
			"clear_on_media_end":  false,
			"close_when_inactive": true,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to set media in OBS: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Zapisz przypisanie w bazie danych
	err = models.SetEpisodeSourceMedia(h.DB, uint(episodeID), sourceName, data.MediaID, "manual")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save assignment: %v", err), http.StatusInternalServerError)
		return
	}

	// Wyślij broadcast do wszystkich klientów
	if h.SocketHandler != nil && h.SocketHandler.Server != nil {
		h.SocketHandler.Server.BroadcastToNamespace("/", "source_media_assigned", map[string]interface{}{
			"episode_id":  episodeID,
			"source_name": sourceName,
			"media_id":    data.MediaID,
			"title":       media.Title,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"episode_id":  episodeID,
		"source_name": sourceName,
		"media_id":    data.MediaID,
		"title":       media.Title,
	})
}

// GetSourceAssignments - GET /api/episodes/{episode_id}/source-assignments
// Pobiera wszystkie przypisania źródeł dla danego odcinka
func (h *EpisodeSourceHandler) GetSourceAssignments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	assignments, err := models.GetAllEpisodeSourceAssignments(h.DB, uint(episodeID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assignments)
}

// GetMediaForSourceModal - GET /api/episodes/{episode_id}/sources/{source_name}/media-list
// Pobiera listę mediów dla modalu - aktualnie przypisany plik + wszystkie grupy
func (h *EpisodeSourceHandler) GetMediaForSourceModal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	// Pobierz aktualnie przypisany plik (jeśli istnieje)
	var currentMediaID *uint
	assignment, err := models.GetEpisodeSourceAssignment(h.DB, uint(episodeID), sourceName)
	if err == nil && assignment != nil && assignment.MediaID != nil {
		currentMediaID = assignment.MediaID
	}

	// Pobierz wszystkie grupy dla tego odcinka z plikami
	var groups []models.MediaGroup
	err = h.DB.Preload("MediaItems.EpisodeMedia").
		Preload("MediaItems", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("episode_id = ?", episodeID).
		Order("is_system DESC, \"order\" ASC").
		Find(&groups).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Przygotuj odpowiedź
	result := make([]map[string]interface{}, 0)
	var currentGroupID *uint

	for _, group := range groups {
		// Przygotuj listę mediów w grupie
		mediaList := make([]map[string]interface{}, 0)
		for _, item := range group.MediaItems {
			mediaList = append(mediaList, map[string]interface{}{
				"id":    item.EpisodeMedia.ID,
				"title": item.EpisodeMedia.Title,
				"order": item.Order,
			})

			// Sprawdź czy to jest grupa z aktualnym plikiem
			if currentMediaID != nil && item.EpisodeMedia.ID == *currentMediaID {
				currentGroupID = &group.ID
			}
		}

		result = append(result, map[string]interface{}{
			"id":          group.ID,
			"name":        group.Name,
			"is_system":   group.IsSystem,
			"media_items": mediaList,
			"is_current":  currentGroupID != nil && *currentGroupID == group.ID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_media_id": currentMediaID,
		"groups":           result,
	})
}

// AutoAssignMediaSources - POST /api/episodes/{episode_id}/auto-assign-media-sources
// Automatycznie przypisuje pierwszy plik z odpowiednich grup do Media1 i Reportaze1
func (h *EpisodeSourceHandler) AutoAssignMediaSources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	results := make(map[string]interface{})

	// Przypisz dla Media1
	if assigned, mediaID, title := h.autoAssignForSource(uint(episodeID), "Media1", "MEDIA"); assigned {
		results["Media1"] = map[string]interface{}{
			"assigned":  true,
			"media_id":  mediaID,
			"title":     title,
			"source":    "auto",
		}
	} else {
		results["Media1"] = map[string]interface{}{
			"assigned": false,
			"reason":   "no media in MEDIA group or already assigned manually",
		}
	}

	// Przypisz dla Reportaze1
	if assigned, mediaID, title := h.autoAssignForSource(uint(episodeID), "Reportaze1", "REPORTAZE"); assigned {
		results["Reportaze1"] = map[string]interface{}{
			"assigned":  true,
			"media_id":  mediaID,
			"title":     title,
			"source":    "auto",
		}
	} else {
		results["Reportaze1"] = map[string]interface{}{
			"assigned": false,
			"reason":   "no media in REPORTAZE group or already assigned manually",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// autoAssignForSource - funkcja pomocnicza do automatycznego przypisania
func (h *EpisodeSourceHandler) autoAssignForSource(episodeID uint, sourceName string, groupName string) (bool, uint, string) {
	// Sprawdź czy już jest przypisanie
	existing, err := models.GetEpisodeSourceAssignment(h.DB, episodeID, sourceName)
	if err == nil && existing != nil {
		// Jest przypisanie - nie nadpisuj
		return false, 0, ""
	}

	// Znajdź grupę o podanej nazwie
	var group models.MediaGroup
	if err := h.DB.Where("episode_id = ? AND name = ?", episodeID, groupName).First(&group).Error; err != nil {
		// Brak grupy
		return false, 0, ""
	}

	// Znajdź pierwszy plik w tej grupie (najniższy order)
	var assignment models.EpisodeMediaGroup
	err = h.DB.Preload("EpisodeMedia").
		Where("media_group_id = ?", group.ID).
		Order("\"order\" ASC").
		First(&assignment).Error

	if err != nil {
		// Brak mediów w grupie
		return false, 0, ""
	}

	media := assignment.EpisodeMedia

	// Sprawdź czy plik istnieje
	if media.FilePath == nil || *media.FilePath == "" {
		return false, 0, ""
	}

	// Wczytaj plik do OBS (jeśli połączony)
	if h.OBSClient != nil && h.OBSClient.IsConnected() {
		absMediaPath, err := filepath.Abs(h.MediaPath)
		if err != nil {
			absMediaPath = h.MediaPath
		}

		fullPath := filepath.Join(absMediaPath, filepath.FromSlash(*media.FilePath))

		err = h.OBSClient.SetInputSettings(sourceName, map[string]interface{}{
			"local_file":          fullPath,
			"clear_on_media_end":  false,
			"close_when_inactive": true,
		})

		if err != nil {
			fmt.Printf("Błąd ustawiania automatycznego pliku w OBS dla %s: %v\n", sourceName, err)
			return false, 0, ""
		}
	}

	// Zapisz przypisanie
	err = models.SetEpisodeSourceMedia(h.DB, episodeID, sourceName, media.ID, "auto")
	if err != nil {
		fmt.Printf("Błąd zapisywania automatycznego przypisania dla %s: %v\n", sourceName, err)
		return false, 0, ""
	}

	// Wyślij broadcast
	if h.SocketHandler != nil && h.SocketHandler.Server != nil {
		h.SocketHandler.Server.BroadcastToNamespace("/", "source_media_assigned", map[string]interface{}{
			"episode_id":  episodeID,
			"source_name": sourceName,
			"media_id":    media.ID,
			"title":       media.Title,
		})
	}

	return true, media.ID, media.Title
}
