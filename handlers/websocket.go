package handlers

import (
	"encoding/json"
	"log"
	"obs-controller/models"
	"obs-controller/obsws"
	"sync"

	socketio "github.com/googollee/go-socket.io"
	"gorm.io/gorm"
)

type SocketHandler struct {
	Server         *socketio.Server
	DB             *gorm.DB
	OBSClient      *obsws.Client
	VolumeMonitor  *VolumeMonitor                    // Monitor zmian głośności
	vlcAssignments map[uint]map[string]VLCAssignment // episode_id -> (source_name -> assignment)
	mu             sync.RWMutex
}

type VLCAssignment struct {
	GroupName string `json:"group_name"`
	GroupID   uint   `json:"group_id"`
}

type ActionRequest struct {
	SceneName  string                 `json:"scene_name"`
	SourceName string                 `json:"source_name"`
	Visible    bool                   `json:"visible"`
	Data       map[string]interface{} `json:"data"`
}

func NewSocketHandler(db *gorm.DB, obsClient *obsws.Client) (*SocketHandler, error) {
	server := socketio.NewServer(nil)

	handler := &SocketHandler{
		Server:         server,
		DB:             db,
		OBSClient:      obsClient,
		vlcAssignments: make(map[uint]map[string]VLCAssignment),
	}

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Printf("Połączono: %s", s.ID())
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("Rozłączono: %s", s.ID())
	})

	server.OnEvent("/", "get_sources", handler.handleGetSources)
	server.OnEvent("/", "toggle_source", handler.handleToggleSource)
	server.OnEvent("/", "send_to_overlay", handler.handleSendToOverlay)
	server.OnEvent("/", "set_source_index", handler.handleSetSourceIndex)
	server.OnEvent("/", "set_current_scene", handler.handleSetCurrentScene)
	server.OnEvent("/", "save_source_order", handler.handleSaveSourceOrder)
	server.OnEvent("/", "sync_source_order", handler.handleSyncSourceOrder)
	server.OnEvent("/", "mute_all_microphones", handler.handleMuteAllMicrophones)
	server.OnEvent("/", "restore_microphones", handler.handleRestoreMicrophones)
	server.OnEvent("/", "set_input_volume", handler.handleSetInputVolume)
	// server.OnEvent("/", "get_input_volume", handler.handleGetInputVolume)

	go server.Serve()

	return handler, nil
}

func (h *SocketHandler) handleGetSources(s socketio.Conn, sceneName string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	items, err := h.OBSClient.GetSceneItemList(sceneName)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	// Najpierw znajdź lub utwórz scenę
	var scene models.Scene
	result := h.DB.Where("name = ?", sceneName).First(&scene)
	if result.Error == gorm.ErrRecordNotFound {
		scene = models.Scene{
			Name: sceneName,
		}
		h.DB.Create(&scene)
	}

	// Sprawdź czy są nowe źródła lub zmiany
	hasChanges := false
	var dbSources []models.Source
	h.DB.Where("scene_id = ?", scene.ID).Find(&dbSources)

	// Mapa źródeł z bazy
	dbSourceMap := make(map[string]models.Source)
	for _, src := range dbSources {
		dbSourceMap[src.Name] = src
	}

	// Synchronizuj źródła - tylko dodawaj nowe, nie aktualizuj istniejących
	for _, item := range items {
		sourceName, _ := item["sourceName"].(string)
		sceneItemIndex, _ := item["sceneItemIndex"].(float64)
		sourceType, _ := item["sourceType"].(string)

		// Pomiń źródła typu SCENE i FILTER
		if sourceType == "OBS_SOURCE_TYPE_SCENE" || sourceType == "OBS_SOURCE_TYPE_FILTER" {
			log.Printf("Pomijam źródło typu %s: %s", sourceType, sourceName)
			continue
		}

		// Jeśli brak sourceType, ustaw domyślny
		if sourceType == "" {
			sourceType = "UNKNOWN"
		}

		if _, exists := dbSourceMap[sourceName]; !exists {
			// Nowe źródło - dodaj z kolejnością z OBS
			source := models.Source{
				SceneID:     scene.ID,
				Name:        sourceName,
				SourceType:  sourceType,
				SourceOrder: int(sceneItemIndex),
				IsVisible:   false,
			}
			h.DB.Create(&source)
			hasChanges = true

			log.Printf("Utworzono nowe źródło: %s (typ: %s)", sourceName, sourceType)
		}
	}

	return h.successResponse(map[string]interface{}{
		"sources":     items,
		"has_changes": hasChanges,
	})
}

func (h *SocketHandler) handleToggleSource(s socketio.Conn, msg string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	var req ActionRequest
	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return h.errorResponse("Błąd")
	}

	if err := h.OBSClient.SetSourceVisibility(req.SceneName, req.SourceName, req.Visible); err != nil {
		return h.errorResponse(err.Error())
	}

	// Dla MIKROFONY zapisuj IsVisible do bazy (stan użytkownika)
	if req.SceneName == "MIKROFONY" {
		var scene models.Scene
		if err := h.DB.Where("name = ?", req.SceneName).First(&scene).Error; err == nil {
			var source models.Source
			if err := h.DB.Where("scene_id = ? AND name = ?", scene.ID, req.SourceName).First(&source).Error; err == nil {
				source.IsVisible = req.Visible
				h.DB.Save(&source)
				log.Printf("Zapisano IsVisible dla %s -> %s: %v", req.SceneName, req.SourceName, req.Visible)
			}
		}
	}

	h.Server.BroadcastToNamespace("/", "source_changed", map[string]interface{}{
		"scene_name":  req.SceneName,
		"source_name": req.SourceName,
		"visible":     req.Visible,
	})

	return h.successResponse(map[string]interface{}{
		"scene_name":  req.SceneName,
		"source_name": req.SourceName,
		"visible":     req.Visible,
	})
}

func (h *SocketHandler) handleSendToOverlay(s socketio.Conn, msg string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		return h.errorResponse("Błąd")
	}

	h.Server.BroadcastToNamespace("/", "overlay_message", data)
	return h.successResponse(data)
}

func (h *SocketHandler) handleSetSourceIndex(s socketio.Conn, msg string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	var req struct {
		SceneName  string `json:"scene_name"`
		SourceName string `json:"source_name"`
		ToTop      bool   `json:"to_top"`
	}

	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return h.errorResponse("Błąd")
	}

	if err := h.OBSClient.SetSceneItemIndex(req.SceneName, req.SourceName, req.ToTop); err != nil {
		return h.errorResponse(err.Error())
	}

	log.Printf("Zmieniono kolejność w OBS (NIE zapisano do bazy): %s -> %s", req.SceneName, req.SourceName)

	return h.successResponse(map[string]interface{}{
		"scene_name":  req.SceneName,
		"source_name": req.SourceName,
		"to_top":      req.ToTop,
	})
}

func (h *SocketHandler) handleSetCurrentScene(s socketio.Conn, msg string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	var req struct {
		SceneName string `json:"scene_name"`
	}

	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return h.errorResponse("Błąd")
	}

	if err := h.OBSClient.SetCurrentProgramScene(req.SceneName); err != nil {
		return h.errorResponse(err.Error())
	}

	return h.successResponse(map[string]interface{}{
		"scene_name": req.SceneName,
	})
}

func (h *SocketHandler) handleSaveSourceOrder(s socketio.Conn, msg string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	var req struct {
		SceneName string `json:"scene_name"`
	}

	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return h.errorResponse("Błąd")
	}

	// Pobierz aktualne źródła z OBS
	items, err := h.OBSClient.GetSceneItemList(req.SceneName)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	// Znajdź scenę w bazie
	var scene models.Scene
	if err := h.DB.Where("name = ?", req.SceneName).First(&scene).Error; err != nil {
		return h.errorResponse("Scena nie znaleziona w bazie")
	}

	// Aktualizuj kolejność źródeł w bazie na podstawie OBS
	for _, item := range items {
		sourceName, _ := item["sourceName"].(string)
		sceneItemIndex, _ := item["sceneItemIndex"].(float64)

		var source models.Source
		result := h.DB.Where("scene_id = ? AND name = ?", scene.ID, sourceName).First(&source)
		if result.Error == nil {
			source.SourceOrder = int(sceneItemIndex)
			h.DB.Save(&source)
			log.Printf("Zapisano do bazy: %s -> %s (order: %d)", req.SceneName, sourceName, int(sceneItemIndex))
		}
	}

	return h.successResponse(map[string]interface{}{
		"scene_name": req.SceneName,
		"updated":    len(items),
	})
}

func (h *SocketHandler) handleSyncSourceOrder(s socketio.Conn, msg string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	var req struct {
		SceneName string `json:"scene_name"`
	}

	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return h.errorResponse("Błąd")
	}

	// Znajdź scenę w bazie
	var scene models.Scene
	if err := h.DB.Where("name = ?", req.SceneName).First(&scene).Error; err != nil {
		return h.errorResponse("Scena nie znaleziona w bazie")
	}

	// Pobierz źródła z bazy
	var dbSources []models.Source
	h.DB.Where("scene_id = ?", scene.ID).Order("source_order DESC").Find(&dbSources)

	// Jeśli nie ma źródeł w bazie, pobierz z OBS i zapisz
	if len(dbSources) == 0 {
		log.Printf("Brak źródeł w bazie dla %s - pobieranie z OBS", req.SceneName)
		items, err := h.OBSClient.GetSceneItemList(req.SceneName)
		if err != nil {
			return h.errorResponse(err.Error())
		}

		// Zapisz źródła z OBS do bazy
		for _, item := range items {
			sourceName, _ := item["sourceName"].(string)
			sceneItemIndex, _ := item["sceneItemIndex"].(float64)

			source := models.Source{
				SceneID:     scene.ID,
				Name:        sourceName,
				SourceOrder: int(sceneItemIndex),
				IsVisible:   false,
			}
			h.DB.Create(&source)
		}

		return h.successResponse(map[string]interface{}{
			"scene_name": req.SceneName,
			"action":     "saved_from_obs",
		})
	}

	// Ustaw kolejność w OBS na podstawie bazy danych
	log.Printf("Synchronizacja kolejności dla %s z bazy do OBS", req.SceneName)
	for _, source := range dbSources {
		if err := h.OBSClient.SetSceneItemIndexByValue(req.SceneName, source.Name, source.SourceOrder); err != nil {
			log.Printf("Błąd ustawiania kolejności %s: %v", source.Name, err)
		}
	}

	return h.successResponse(map[string]interface{}{
		"scene_name": req.SceneName,
		"action":     "synced_to_obs",
		"count":      len(dbSources),
	})
}

func (h *SocketHandler) handleMuteAllMicrophones(s socketio.Conn, msg string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	// Znajdź scenę MIKROFONY
	var scene models.Scene
	if err := h.DB.Where("name = ?", "MIKROFONY").First(&scene).Error; err != nil {
		return h.errorResponse("Scena MIKROFONY nie znaleziona")
	}

	// Pobierz wszystkie mikrofony
	var sources []models.Source
	h.DB.Where("scene_id = ?", scene.ID).Find(&sources)

	// Wyłącz wszystkie mikrofony w OBS (BEZ zmiany is_visible)
	mutedCount := 0
	for _, source := range sources {
		if err := h.OBSClient.SetSourceVisibility("MIKROFONY", source.Name, false); err != nil {
			log.Printf("Błąd wyłączania mikrofonu %s: %v", source.Name, err)
		} else {
			mutedCount++
			// Broadcast zmiany do klientów
			h.Server.BroadcastToNamespace("/", "source_changed", map[string]interface{}{
				"scene_name":  "MIKROFONY",
				"source_name": source.Name,
				"visible":     false,
			})
		}
	}

	log.Printf("Wyciszono %d mikrofonów (reportaż)", mutedCount)

	return h.successResponse(map[string]interface{}{
		"muted": mutedCount,
	})
}

func (h *SocketHandler) handleRestoreMicrophones(s socketio.Conn, msg string) string {
	if h.OBSClient == nil {
		return h.errorResponse("OBS nie jest połączony")
	}

	// Znajdź scenę MIKROFONY
	var scene models.Scene
	if err := h.DB.Where("name = ?", "MIKROFONY").First(&scene).Error; err != nil {
		return h.errorResponse("Scena MIKROFONY nie znaleziona")
	}

	// Pobierz mikrofony z is_visible = true
	var sources []models.Source
	h.DB.Where("scene_id = ? AND is_visible = ?", scene.ID, true).Find(&sources)

	// Włącz mikrofony które były aktywne
	restoredCount := 0
	for _, source := range sources {
		if err := h.OBSClient.SetSourceVisibility("MIKROFONY", source.Name, true); err != nil {
			log.Printf("Błąd przywracania mikrofonu %s: %v", source.Name, err)
		} else {
			restoredCount++
			// Broadcast zmiany do klientów
			h.Server.BroadcastToNamespace("/", "source_changed", map[string]interface{}{
				"scene_name":  "MIKROFONY",
				"source_name": source.Name,
				"visible":     true,
			})
		}
	}

	log.Printf("Przywrócono %d mikrofonów (kamery)", restoredCount)

	return h.successResponse(map[string]interface{}{
		"restored": restoredCount,
	})
}

func (h *SocketHandler) successResponse(data interface{}) string {
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}
	jsonData, _ := json.Marshal(response)
	return string(jsonData)
}

func (h *SocketHandler) errorResponse(message string) string {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}
	jsonData, _ := json.Marshal(response)
	return string(jsonData)
}

// SaveVLCAssignment zapisuje przypisanie grupy do źródła VLC dla danego odcinka
func (h *SocketHandler) SaveVLCAssignment(episodeID uint, sourceName string, groupName string, groupID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.vlcAssignments[episodeID] == nil {
		h.vlcAssignments[episodeID] = make(map[string]VLCAssignment)
	}

	h.vlcAssignments[episodeID][sourceName] = VLCAssignment{
		GroupName: groupName,
		GroupID:   groupID,
	}

	log.Printf("Zapisano przypisanie VLC: Episode %d, Source %s -> Group %s (ID: %d)", episodeID, sourceName, groupName, groupID)
}

// GetVLCAssignments pobiera przypisania VLC dla danego odcinka
func (h *SocketHandler) GetVLCAssignments(episodeID uint) map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]interface{})

	if assignments, exists := h.vlcAssignments[episodeID]; exists {
		for sourceName, assignment := range assignments {
			result[sourceName] = map[string]interface{}{
				"group_name": assignment.GroupName,
				"group_id":   assignment.GroupID,
			}
		}
	}

	return result
}

// handleSetInputVolume - ustaw głośność źródła audio
func (h *SocketHandler) handleSetInputVolume(s socketio.Conn, msg string) string {
	var data struct {
		InputName     string  `json:"inputName"`
		InputVolumeDb float64 `json:"inputVolumeDb"`
	}

	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		return h.errorResponse("Invalid data")
	}

	if h.OBSClient == nil {
		return h.errorResponse("OBS not connected")
	}

	// Zarejestruj że TO NASZA ZMIANA (dla VolumeMonitor)
	if h.VolumeMonitor != nil {
		h.VolumeMonitor.RegisterOurChange(data.InputName, data.InputVolumeDb)
	}

	// Ustaw głośność w OBS
	err := h.OBSClient.SetInputVolume(data.InputName, data.InputVolumeDb)
	if err != nil {
		log.Printf("Error setting volume for %s: %v", data.InputName, err)
		return h.errorResponse(err.Error())
	}

	log.Printf("Volume set: %s = %.2f dB", data.InputName, data.InputVolumeDb)

	return h.successResponse(map[string]interface{}{
		"source_name": data.InputName,
		"volume_db":   data.InputVolumeDb,
	})
}

// handleGetInputVolume - pobierz aktualną głośność źródła audio
// func (h *SocketHandler) handleGetInputVolume(s socketio.Conn, inputName string) string {
// 	if h.OBSClient == nil {
// 		return h.errorResponse("OBS not connected")
// 	}

// 	volumeDb, err := h.OBSClient.GetInputVolume(inputName)
// 	if err != nil {
// 		log.Printf("Error getting volume for %s: %v", inputName, err)
// 		return h.errorResponse(err.Error())
// 	}

// 	return h.successResponse(map[string]interface{}{
// 		"source_name": inputName,
// 		"volume_db":   volumeDb,
// 	})
//}
