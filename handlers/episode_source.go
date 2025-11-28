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

// AutoAssignVLCSources - POST /api/episodes/{episode_id}/auto-assign-vlc-sources
// Automatycznie przypisuje grupy do Media2 i Reportaze2
func (h *EpisodeSourceHandler) AutoAssignVLCSources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	results := make(map[string]interface{})

	// Przypisz dla Media2
	if assigned, groupID, groupName := h.autoAssignVLCForSource(uint(episodeID), "Media2", "MEDIA"); assigned {
		results["Media2"] = map[string]interface{}{
			"assigned":  true,
			"group_id":  groupID,
			"name":      groupName,
			"source":    "auto",
		}
	} else {
		results["Media2"] = map[string]interface{}{
			"assigned": false,
			"reason":   "no group with ≥2 files or already assigned",
		}
	}

	// Przypisz dla Reportaze2
	if assigned, groupID, groupName := h.autoAssignVLCForSource(uint(episodeID), "Reportaze2", "REPORTAZE"); assigned {
		results["Reportaze2"] = map[string]interface{}{
			"assigned":  true,
			"group_id":  groupID,
			"name":      groupName,
			"source":    "auto",
		}
	} else {
		results["Reportaze2"] = map[string]interface{}{
			"assigned": false,
			"reason":   "no group with ≥2 files or already assigned",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// autoAssignVLCForSource - funkcja pomocnicza do automatycznego przypisania grupy
// Priorytet: 1) grupa systemowa z ≥2 plikami, 2) pierwsza grupa użytkownika z ≥2 plikami
func (h *EpisodeSourceHandler) autoAssignVLCForSource(episodeID uint, sourceName string, systemGroupName string) (bool, uint, string) {
	// Sprawdź czy już jest przypisanie
	existing, err := models.GetEpisodeSourceAssignment(h.DB, episodeID, sourceName)
	if err == nil && existing != nil {
		// Jest przypisanie - nie nadpisuj
		return false, 0, ""
	}

	var selectedGroup *models.MediaGroup

	// PRIORYTET 1: Grupa systemowa (MEDIA/REPORTAZE) z ≥2 plikami
	var systemGroup models.MediaGroup
	err = h.DB.Preload("MediaItems.EpisodeMedia").
		Where("episode_id = ? AND name = ? AND is_system = ?", episodeID, systemGroupName, true).
		First(&systemGroup).Error

	if err == nil && len(systemGroup.MediaItems) >= 2 {
		selectedGroup = &systemGroup
		fmt.Printf("Auto-assign VLC: znaleziono grupę systemową %s z %d plikami\n", systemGroup.Name, len(systemGroup.MediaItems))
	}

	// PRIORYTET 2: Pierwsza grupa użytkownika z ≥2 plikami
	if selectedGroup == nil {
		var userGroups []models.MediaGroup
		err = h.DB.Preload("MediaItems.EpisodeMedia").
			Where("episode_id = ? AND is_system = ?", episodeID, false).
			Order("\"order\" ASC").
			Find(&userGroups).Error

		if err == nil {
			for _, group := range userGroups {
				if len(group.MediaItems) >= 2 {
					selectedGroup = &group
					fmt.Printf("Auto-assign VLC: znaleziono grupę użytkownika %s z %d plikami\n", group.Name, len(group.MediaItems))
					break
				}
			}
		}
	}

	// Brak odpowiedniej grupy
	if selectedGroup == nil {
		fmt.Printf("Auto-assign VLC: brak grupy z ≥2 plikami dla %s\n", sourceName)
		return false, 0, ""
	}

	// Przygotuj playlistę dla VLC Video Source
	playlist := make([]map[string]interface{}, 0)
	
	absMediaPath, err := filepath.Abs(h.MediaPath)
	if err != nil {
		absMediaPath = h.MediaPath
	}

	for _, item := range selectedGroup.MediaItems {
		// EpisodeMedia jest już załadowane przez Preload("MediaItems.EpisodeMedia")
		media := item.EpisodeMedia

		if media.FilePath != nil && *media.FilePath != "" {
			fullPath := filepath.Join(absMediaPath, filepath.FromSlash(*media.FilePath))
			playlist = append(playlist, map[string]interface{}{
				"value": fullPath,
			})
		}
	}

	if len(playlist) == 0 {
		fmt.Printf("Auto-assign VLC: brak prawidłowych plików w grupie %s\n", selectedGroup.Name)
		return false, 0, ""
	}

	// Wczytaj playlistę do OBS (jeśli połączony)
	if h.OBSClient != nil && h.OBSClient.IsConnected() {
		err = h.OBSClient.SetInputSettings(sourceName, map[string]interface{}{
			"playlist": playlist,
			"loop":     false,
			"shuffle":  false,
		})

		if err != nil {
			fmt.Printf("Błąd ustawiania automatycznej playlisty w OBS dla %s: %v\n", sourceName, err)
			return false, 0, ""
		}

		fmt.Printf("Auto-assign VLC: wczytano %d plików do źródła %s\n", len(playlist), sourceName)
	}

	// Zapisz przypisanie
	err = models.SetEpisodeSourceGroup(h.DB, episodeID, sourceName, selectedGroup.ID, "auto")
	if err != nil {
		fmt.Printf("Błąd zapisywania automatycznego przypisania grupy dla %s: %v\n", sourceName, err)
		return false, 0, ""
	}

	// Wyślij broadcast
	if h.SocketHandler != nil && h.SocketHandler.Server != nil {
		h.SocketHandler.Server.BroadcastToNamespace("/", "source_group_assigned", map[string]interface{}{
			"episode_id":  episodeID,
			"source_name": sourceName,
			"group_id":    selectedGroup.ID,
			"name":        selectedGroup.Name,
		})
	}

	return true, selectedGroup.ID, selectedGroup.Name
}

// AssignGroupToSource - POST /api/episodes/{episode_id}/sources/{source_name}/assign-group
// Ręczne przypisanie grupy do źródła VLC Video Source
func (h *EpisodeSourceHandler) AssignGroupToSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	var requestData struct {
		GroupID uint `json:"group_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Pobierz grupę i sprawdź czy należy do odcinka
	var group models.MediaGroup
	if err := h.DB.Preload("MediaItems.EpisodeMedia").First(&group, requestData.GroupID).Error; err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	if group.EpisodeID != uint(episodeID) {
		http.Error(w, "Group does not belong to this episode", http.StatusForbidden)
		return
	}

	// Sprawdź czy grupa ma ≥2 pliki
	if len(group.MediaItems) < 2 {
		http.Error(w, "Group must have at least 2 files for VLC Video Source", http.StatusBadRequest)
		return
	}

	// Przygotuj playlistę dla VLC Video Source
	playlist := make([]map[string]interface{}, 0)

	absMediaPath, err := filepath.Abs(h.MediaPath)
	if err != nil {
		absMediaPath = h.MediaPath
	}

	for _, item := range group.MediaItems {
		// EpisodeMedia jest już załadowane przez Preload("MediaItems.EpisodeMedia")
		media := item.EpisodeMedia

		if media.FilePath != nil && *media.FilePath != "" {
			fullPath := filepath.Join(absMediaPath, filepath.FromSlash(*media.FilePath))
			playlist = append(playlist, map[string]interface{}{
				"value": fullPath,
			})
		}
	}

	if len(playlist) == 0 {
		http.Error(w, "No valid files in group", http.StatusBadRequest)
		return
	}

	// Wczytaj playlistę do OBS (jeśli połączony)
	if h.OBSClient != nil && h.OBSClient.IsConnected() {
		err = h.OBSClient.SetInputSettings(sourceName, map[string]interface{}{
			"playlist": playlist,
			"loop":     false,
			"shuffle":  false,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to set OBS playlist: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Zapisz przypisanie jako "manual"
	err = models.SetEpisodeSourceGroup(h.DB, uint(episodeID), sourceName, group.ID, "manual")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save assignment: %v", err), http.StatusInternalServerError)
		return
	}

	// Wyślij broadcast
	if h.SocketHandler != nil && h.SocketHandler.Server != nil {
		h.SocketHandler.Server.BroadcastToNamespace("/", "source_group_assigned", map[string]interface{}{
			"episode_id":  uint(episodeID),
			"source_name": sourceName,
			"group_id":    group.ID,
			"name":        group.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"episode_id":  uint(episodeID),
		"source_name": sourceName,
		"group_id":    group.ID,
		"name":        group.Name,
	})
}

// GetGroupsForSourceModal - GET /api/episodes/{episode_id}/sources/{source_name}/groups-list
// Pobiera listę grup z ≥2 plikami dla modalu wyboru
func (h *EpisodeSourceHandler) GetGroupsForSourceModal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	// Pobierz aktualnie przypisaną grupę (jeśli istnieje)
	var currentGroupID *uint
	assignment, err := models.GetEpisodeSourceAssignment(h.DB, uint(episodeID), sourceName)
	if err == nil && assignment != nil && assignment.GroupID != nil {
		currentGroupID = assignment.GroupID
	}

	// Pobierz wszystkie grupy dla tego odcinka z plikami
	var groups []models.MediaGroup
	err = h.DB.Preload("MediaItems").
		Where("episode_id = ?", episodeID).
		Order("is_system DESC, \"order\" ASC").
		Find(&groups).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Przygotuj odpowiedź - tylko grupy z ≥2 plikami
	result := make([]map[string]interface{}, 0)

	for _, group := range groups {
		// Filtruj - tylko grupy z ≥2 plikami
		if len(group.MediaItems) < 2 {
			continue
		}

		isCurrent := currentGroupID != nil && group.ID == *currentGroupID

		result = append(result, map[string]interface{}{
			"id":         group.ID,
			"name":       group.Name,
			"is_system":  group.IsSystem,
			"file_count": len(group.MediaItems),
			"is_current": isCurrent,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_group_id": currentGroupID,
		"groups":           result,
	})
}
