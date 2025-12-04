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

	// Wczytaj plik do OBS używając helper function
	if err := h.setVLCPlaylistFromMedia(sourceName, &media); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set media in OBS: %v", err), http.StatusInternalServerError)
		return
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
			"assigned": true,
			"media_id": mediaID,
			"title":    title,
			"source":   "auto",
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
			"assigned": true,
			"media_id": mediaID,
			"title":    title,
			"source":   "auto",
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

	// Wczytaj plik do OBS używając helper function
	if err := h.setVLCPlaylistFromMedia(sourceName, &media); err != nil {
		fmt.Printf("Błąd ustawiania automatycznego pliku w OBS dla %s: %v\n", sourceName, err)
		return false, 0, ""
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
			"assigned": true,
			"group_id": groupID,
			"name":     groupName,
			"source":   "auto",
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
			"assigned": true,
			"group_id": groupID,
			"name":     groupName,
			"source":   "auto",
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

	// Wczytaj playlistę do OBS używając helper function
	if err := h.setVLCPlaylistFromGroup(sourceName, selectedGroup); err != nil {
		fmt.Printf("Błąd ustawiania automatycznej playlisty w OBS dla %s: %v\n", sourceName, err)
		return false, 0, ""
	}

	fmt.Printf("Auto-assign VLC: wczytano grupę %s (%d plików) do źródła %s\n",
		selectedGroup.Name, len(selectedGroup.MediaItems), sourceName)

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

	// Wczytaj playlistę do OBS używając helper function
	if err := h.setVLCPlaylistFromGroup(sourceName, &group); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set OBS playlist: %v", err), http.StatusInternalServerError)
		return
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

// AutoAssignCameraTypes - POST /api/episodes/{episode_id}/auto-assign-camera-types
// Automatycznie przypisuje typy kamer do Kamera1-4 według kolejności (order)
func (h *EpisodeSourceHandler) AutoAssignCameraTypes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	// Mapowanie: source_name → order typu kamery
	cameraMapping := map[string]int{
		"Kamera1": 1, // Centralna
		"Kamera2": 2, // Prowadzący
		"Kamera3": 3, // Goście
		"Kamera4": 4, // Dodatkowa
	}

	results := make(map[string]interface{})

	for sourceName, order := range cameraMapping {
		// Sprawdź czy już jest przypisanie
		existing, err := models.GetEpisodeSourceAssignment(h.DB, uint(episodeID), sourceName)
		if err == nil && existing != nil && existing.AssignedBy == "manual" {
			// Manual assignment - nie nadpisuj
			results[sourceName] = map[string]interface{}{
				"assigned": false,
				"reason":   "manual assignment exists",
			}
			continue
		}

		// Znajdź typ kamery według kolejności
		var cameraType models.CameraType
		err = h.DB.Where("\"order\" = ?", order).First(&cameraType).Error
		if err != nil {
			results[sourceName] = map[string]interface{}{
				"assigned": false,
				"reason":   fmt.Sprintf("camera type with order %d not found", order),
			}
			continue
		}

		// Przypisz typ kamery
		err = models.SetEpisodeSourceCameraType(h.DB, uint(episodeID), sourceName, cameraType.ID, "auto")
		if err != nil {
			results[sourceName] = map[string]interface{}{
				"assigned": false,
				"reason":   err.Error(),
			}
			continue
		}

		// Sukces
		results[sourceName] = map[string]interface{}{
			"assigned":         true,
			"camera_type_id":   cameraType.ID,
			"camera_type_name": cameraType.Name,
			"source":           "auto",
		}

		// Broadcast WebSocket
		if h.SocketHandler != nil {
			h.SocketHandler.Server.BroadcastToNamespace("/", "source_camera_assigned", map[string]interface{}{
				"episode_id":       uint(episodeID),
				"source_name":      sourceName,
				"camera_type_id":   cameraType.ID,
				"camera_type_name": cameraType.Name,
				"is_disabled":      false,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// AssignCameraTypeToSource - POST /api/episodes/{episode_id}/sources/{source_name}/assign-camera-type
// Ręczne przypisanie typu kamery do źródła
func (h *EpisodeSourceHandler) AssignCameraTypeToSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	var data struct {
		CameraTypeID *uint `json:"camera_type_id"` // null = wyłącz kamerę
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Jeśli camera_type_id = null → wyłącz kamerę
	if data.CameraTypeID == nil {
		err := models.DisableEpisodeSourceCamera(h.DB, uint(episodeID), sourceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Broadcast WebSocket
		if h.SocketHandler != nil {
			h.SocketHandler.Server.BroadcastToNamespace("/", "source_camera_assigned", map[string]interface{}{
				"episode_id":       uint(episodeID),
				"source_name":      sourceName,
				"camera_type_id":   nil,
				"camera_type_name": nil,
				"is_disabled":      true,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"is_disabled": true,
		})
		return
	}

	// Walidacja: typ kamery istnieje
	var cameraType models.CameraType
	if err := h.DB.First(&cameraType, *data.CameraTypeID).Error; err != nil {
		http.Error(w, "Camera type not found", http.StatusNotFound)
		return
	}

	// Walidacja: typ nie jest już przypisany do innej kamery w tym odcinku
	var existingAssignment models.EpisodeSource
	result := h.DB.Where("episode_id = ? AND camera_type_id = ? AND source_name != ?",
		episodeID, *data.CameraTypeID, sourceName).First(&existingAssignment)

	if result.Error == nil {
		// Ten typ jest już przypisany do innej kamery
		http.Error(w, fmt.Sprintf("Camera type '%s' is already assigned to %s",
			cameraType.Name, existingAssignment.SourceName), http.StatusConflict)
		return
	}

	// Przypisz typ kamery
	err = models.SetEpisodeSourceCameraType(h.DB, uint(episodeID), sourceName, *data.CameraTypeID, "manual")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast WebSocket
	if h.SocketHandler != nil {
		h.SocketHandler.Server.BroadcastToNamespace("/", "source_camera_assigned", map[string]interface{}{
			"episode_id":       uint(episodeID),
			"source_name":      sourceName,
			"camera_type_id":   *data.CameraTypeID,
			"camera_type_name": cameraType.Name,
			"is_disabled":      false,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":          true,
		"camera_type_id":   cameraType.ID,
		"camera_type_name": cameraType.Name,
	})
}

// GetCameraTypesForModal - GET /api/episodes/{episode_id}/sources/{source_name}/camera-types-list
// Zwraca listę typów kamer dla modalu wyboru
func (h *EpisodeSourceHandler) GetCameraTypesForModal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	// Pobierz aktualne przypisanie dla tej kamery
	var currentCameraTypeID *uint
	existing, err := models.GetEpisodeSourceAssignment(h.DB, uint(episodeID), sourceName)
	if err == nil && existing != nil {
		currentCameraTypeID = existing.CameraTypeID
	}

	// KROK 1: Pobierz WSZYSTKIE typy kamer z tabeli camera_types
	var cameraTypes []models.CameraType
	if err := h.DB.Order("\"order\" ASC").Find(&cameraTypes).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// KROK 2: Sprawdź które typy są już przypisane do INNYCH kamer w tym odcinku
	// (tylko po to, żeby zapobiec duplikatom)
	var episodeSources []models.EpisodeSource
	h.DB.Where("episode_id = ? AND camera_type_id IS NOT NULL AND source_name != ?",
		episodeID, sourceName).Find(&episodeSources)

	// Mapa: camera_type_id → source_name (która kamera ma który typ)
	assignedTypes := make(map[uint]string)
	for _, es := range episodeSources {
		if es.CameraTypeID != nil {
			assignedTypes[*es.CameraTypeID] = es.SourceName
		}
	}

	// KROK 3: Przygotuj wynik - WSZYSTKIE typy z info o zajętości
	result := make([]map[string]interface{}, 0)
	for _, ct := range cameraTypes {
		assignedTo := assignedTypes[ct.ID]
		isCurrent := currentCameraTypeID != nil && *currentCameraTypeID == ct.ID

		result = append(result, map[string]interface{}{
			"id":          ct.ID,
			"name":        ct.Name,
			"order":       ct.Order,
			"is_system":   ct.IsSystem,
			"is_assigned": assignedTo != "", // true jeśli zajęty przez INNĄ kamerę
			"assigned_to": assignedTo,       // nazwa kamery która go używa
			"is_current":  isCurrent,        // true jeśli to aktualny typ TEJ kamery
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_camera_type_id": currentCameraTypeID,
		"camera_types":           result,
	})
}

// GetMicrophonePeopleList zwraca listę Staff + Guests dla modalu mikrofonów
func (h *EpisodeSourceHandler) GetMicrophonePeopleList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	// Pobierz aktualne przypisanie
	var currentPersonID *uint
	var currentPersonType string
	existing, err := models.GetEpisodeSourceAssignment(h.DB, uint(episodeID), sourceName)
	if err == nil && existing != nil {
		if existing.StaffID != nil {
			currentPersonID = existing.StaffID
			currentPersonType = "staff"
		} else if existing.GuestID != nil {
			currentPersonID = existing.GuestID
			currentPersonType = "guest"
		}
	}

	// Pobierz listę Staff dla tego odcinka
	var episodeStaff []models.EpisodeStaff
	h.DB.Where("episode_id = ?", episodeID).
		Preload("Staff").
		Preload("StaffTypes").
		Order("staff_id ASC").
		Find(&episodeStaff)

	// Pobierz listę Guests dla tego odcinka
	var episodeGuests []models.EpisodeGuest
	h.DB.Where("episode_id = ?", episodeID).
		Preload("Guest.GuestType").
		Order("guest_id ASC").
		Find(&episodeGuests)

	// Pobierz wszystkie przypisania mikrofonów w tym odcinku
	var episodeSources []models.EpisodeSource
	h.DB.Where("episode_id = ? AND (staff_id IS NOT NULL OR guest_id IS NOT NULL)", episodeID).
		Find(&episodeSources)

	// Mapa: person_type:person_id → []source_names
	assignedMicrophones := make(map[string][]string)
	for _, es := range episodeSources {
		if es.StaffID != nil {
			key := fmt.Sprintf("staff:%d", *es.StaffID)
			assignedMicrophones[key] = append(assignedMicrophones[key], es.SourceName)
		} else if es.GuestID != nil {
			key := fmt.Sprintf("guest:%d", *es.GuestID)
			assignedMicrophones[key] = append(assignedMicrophones[key], es.SourceName)
		}
	}

	// Przygotuj listę Staff
	staffList := make([]map[string]interface{}, 0)
	for _, es := range episodeStaff {
		key := fmt.Sprintf("staff:%d", es.StaffID)
		microphones := assignedMicrophones[key]
		isCurrent := currentPersonType == "staff" && currentPersonID != nil && *currentPersonID == es.StaffID

		// Pobierz nazwę typu z pierwszego typu (jeśli istnieje)
		var staffTypeName string
		if len(es.StaffTypes) > 0 {
			var staffType models.StaffType
			if err := h.DB.First(&staffType, es.StaffTypes[0].StaffTypeID).Error; err == nil {
				staffTypeName = staffType.Name
			}
		}

		staffList = append(staffList, map[string]interface{}{
			"id":                   es.StaffID,
			"type":                 "staff",
			"first_name":           es.Staff.FirstName,
			"last_name":            es.Staff.LastName,
			"full_name":            es.Staff.FirstName + " " + es.Staff.LastName,
			"staff_type_name":      staffTypeName,
			"is_current":           isCurrent,
			"assigned_microphones": microphones,
		})
	}

	// Przygotuj listę Guests
	guestList := make([]map[string]interface{}, 0)
	for _, eg := range episodeGuests {
		key := fmt.Sprintf("guest:%d", eg.GuestID)
		microphones := assignedMicrophones[key]
		isCurrent := currentPersonType == "guest" && currentPersonID != nil && *currentPersonID == eg.GuestID

		guestList = append(guestList, map[string]interface{}{
			"id":                   eg.GuestID,
			"type":                 "guest",
			"first_name":           eg.Guest.FirstName,
			"last_name":            eg.Guest.LastName,
			"full_name":            eg.Guest.FirstName + " " + eg.Guest.LastName,
			"guest_type_name":      eg.Guest.GuestType.Name,
			"is_current":           isCurrent,
			"assigned_microphones": microphones,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_person_id":   currentPersonID,
		"current_person_type": currentPersonType,
		"staff":               staffList,
		"guests":              guestList,
	})
}

// AssignMicrophonePerson przypisuje osobę (staff/guest) do mikrofonu
func (h *EpisodeSourceHandler) AssignMicrophonePerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	sourceName := vars["source_name"]

	var data struct {
		PersonID   *uint  `json:"person_id"`   // null = usuń przypisanie
		PersonType string `json:"person_type"` // "staff" lub "guest"
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Jeśli PersonID == null, usuń przypisanie
	if data.PersonID == nil {
		if err := models.UnassignEpisodeSourceMicrophone(h.DB, uint(episodeID), sourceName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Broadcast WebSocket
		if h.SocketHandler != nil {
			h.SocketHandler.Server.BroadcastToNamespace("/", "source_microphone_assigned", map[string]interface{}{
				"episode_id":  episodeID,
				"source_name": sourceName,
				"person_id":   nil,
				"person_type": "",
				"person_name": sourceName, // Przywróć oryginalną nazwę
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}

	// Walidacja: sprawdź czy osoba istnieje w odcinku
	var personName string
	if data.PersonType == "staff" {
		var episodeStaff models.EpisodeStaff
		if err := h.DB.Where("episode_id = ? AND staff_id = ?", episodeID, *data.PersonID).
			Preload("Staff").First(&episodeStaff).Error; err != nil {
			http.Error(w, "Staff not found in this episode", http.StatusNotFound)
			return
		}
		personName = episodeStaff.Staff.LastName + " " + episodeStaff.Staff.FirstName
	} else if data.PersonType == "guest" {
		var episodeGuest models.EpisodeGuest
		if err := h.DB.Where("episode_id = ? AND guest_id = ?", episodeID, *data.PersonID).
			Preload("Guest").First(&episodeGuest).Error; err != nil {
			http.Error(w, "Guest not found in this episode", http.StatusNotFound)
			return
		}
		personName = episodeGuest.Guest.LastName + " " + episodeGuest.Guest.FirstName
	} else {
		http.Error(w, "Invalid person_type", http.StatusBadRequest)
		return
	}

	// Przypisz osobę do mikrofonu
	if err := models.SetEpisodeSourceMicrophone(h.DB, uint(episodeID), sourceName, *data.PersonID, data.PersonType, "manual"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast WebSocket
	if h.SocketHandler != nil {
		h.SocketHandler.Server.BroadcastToNamespace("/", "source_microphone_assigned", map[string]interface{}{
			"episode_id":  episodeID,
			"source_name": sourceName,
			"person_id":   *data.PersonID,
			"person_type": data.PersonType,
			"person_name": personName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Assigned %s to %s", personName, sourceName),
	})
}

// GetCameraAssignments - GET /api/episodes/{episode_id}/camera-assignments
// Pobiera wszystkie przypisania kamer dla odcinka (bez auto-assign, bez broadcast)
func (h *EpisodeSourceHandler) GetCameraAssignments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	episodeID, err := strconv.ParseUint(vars["episode_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid episode ID", http.StatusBadRequest)
		return
	}

	// Pobierz wszystkie episode_sources dla tego odcinka gdzie camera_type_id != NULL
	var episodeSources []models.EpisodeSource
	err = h.DB.Where("episode_id = ? AND camera_type_id IS NOT NULL", uint(episodeID)).
		Preload("CameraType").
		Find(&episodeSources).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Przygotuj mapę: source_name → camera_type_info
	result := make(map[string]interface{})

	for _, es := range episodeSources {
		if es.CameraType != nil {
			result[es.SourceName] = map[string]interface{}{
				"camera_type_id":   es.CameraType.ID,
				"camera_type_name": es.CameraType.Name,
				"assigned_by":      es.AssignedBy,
				"is_disabled":      false,
			}
		}
	}

	// Zwróć przypisania (bez żadnego broadcast)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ===================================================================
// HELPER FUNCTIONS - VLC Playlist Management
// ===================================================================

// setVLCPlaylistSingle ustawia pojedynczy plik w źródle VLC (Media1, Reportaze1)
// Zwraca error jeśli wystąpił problem z OBS
func (h *EpisodeSourceHandler) setVLCPlaylistSingle(sourceName string, filePath string) error {
	if h.OBSClient == nil || !h.OBSClient.IsConnected() {
		return nil // OBS nie połączony - to nie jest błąd
	}

	// Przygotuj playlist z jednym plikiem
	absMediaPath, err := filepath.Abs(h.MediaPath)
	if err != nil {
		absMediaPath = h.MediaPath
	}

	fullPath := filepath.Join(absMediaPath, filepath.FromSlash(filePath))
	playlist := []map[string]interface{}{
		{"value": fullPath},
	}

	// Ustaw w OBS
	return h.OBSClient.SetInputSettings(sourceName, map[string]interface{}{
		"playlist": playlist,
		"loop":     false,
		"shuffle":  false,
	})
}

// setVLCPlaylistMultiple ustawia wiele plików w źródle VLC (Media2, Reportaze2)
// Zwraca error jeśli wystąpił problem z OBS
func (h *EpisodeSourceHandler) setVLCPlaylistMultiple(sourceName string, filePaths []string) error {
	if h.OBSClient == nil || !h.OBSClient.IsConnected() {
		return nil // OBS nie połączony - to nie jest błąd
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("no files provided for playlist")
	}

	// Przygotuj playlist z wieloma plikami
	absMediaPath, err := filepath.Abs(h.MediaPath)
	if err != nil {
		absMediaPath = h.MediaPath
	}

	playlist := make([]map[string]interface{}, 0, len(filePaths))
	for _, relPath := range filePaths {
		if relPath != "" {
			fullPath := filepath.Join(absMediaPath, filepath.FromSlash(relPath))
			playlist = append(playlist, map[string]interface{}{
				"value": fullPath,
			})
		}
	}

	if len(playlist) == 0 {
		return fmt.Errorf("no valid files in playlist")
	}

	// Ustaw w OBS
	return h.OBSClient.SetInputSettings(sourceName, map[string]interface{}{
		"playlist": playlist,
		"loop":     false,
		"shuffle":  false,
	})
}

// setVLCPlaylistFromMedia ustawia playlist dla pojedynczego media (wrapper dla setVLCPlaylistSingle)
func (h *EpisodeSourceHandler) setVLCPlaylistFromMedia(sourceName string, media *models.EpisodeMedia) error {
	if media.FilePath == nil || *media.FilePath == "" {
		return fmt.Errorf("media has no file path")
	}
	return h.setVLCPlaylistSingle(sourceName, *media.FilePath)
}

// setVLCPlaylistFromGroup ustawia playlist z grupy mediów (wrapper dla setVLCPlaylistMultiple)
func (h *EpisodeSourceHandler) setVLCPlaylistFromGroup(sourceName string, group *models.MediaGroup) error {
	if len(group.MediaItems) == 0 {
		return fmt.Errorf("group has no media items")
	}

	// Zbierz ścieżki plików z grupy
	filePaths := make([]string, 0, len(group.MediaItems))
	for _, item := range group.MediaItems {
		media := item.EpisodeMedia
		if media.FilePath != nil && *media.FilePath != "" {
			filePaths = append(filePaths, *media.FilePath)
		}
	}

	return h.setVLCPlaylistMultiple(sourceName, filePaths)
}
