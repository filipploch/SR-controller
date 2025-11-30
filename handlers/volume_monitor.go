package handlers

import (
	"log"
	"obs-controller/obsws"
	"sync"
	"time"
)

type VolumeMonitor struct {
	OBSClient     *obsws.Client
	SocketHandler *SocketHandler

	// Mapa: source_name → ostatnia wartość ustawiona przez nas
	ourChanges    map[string]float64
	ourChangesMux sync.RWMutex

	// Cache: source_name → ostatnia znana głośność z OBS
	cachedVolumes map[string]float64
	cacheMu       sync.RWMutex
}

func NewVolumeMonitor(obsClient *obsws.Client, socketHandler *SocketHandler) *VolumeMonitor {
	return &VolumeMonitor{
		OBSClient:     obsClient,
		SocketHandler: socketHandler,
		ourChanges:    make(map[string]float64),
		cachedVolumes: make(map[string]float64),
	}
}

// Start rozpoczyna nasłuchiwanie na zmiany głośności z OBS
func (vm *VolumeMonitor) Start() {
	log.Println("Starting Volume Monitor...")

	// Subskrybuj event InputVolumeChanged
	vm.OBSClient.OnEvent("InputVolumeChanged", func(event map[string]interface{}) {
		inputName, ok1 := event["inputName"].(string)
		volumeDb, ok2 := event["inputVolumeDb"].(float64)

		if !ok1 || !ok2 {
			log.Printf("Invalid InputVolumeChanged event: %+v", event)
			return
		}

		// Sprawdź czy to zmiana z naszej aplikacji
		if vm.isOurChange(inputName, volumeDb) {
			// To nasza zmiana - ignoruj (nie broadcastuj)
			log.Printf("Volume change ignored (our change): %s = %.2f dB", inputName, volumeDb)
			vm.clearOurChange(inputName)

			// Ale AKTUALIZUJ CACHE (nasza zmiana też jest prawidłowa)
			vm.UpdateCache(inputName, volumeDb)
			return
		}

		// To zewnętrzna zmiana (z OBS UI) - broadcast do frontendu

		// Aktualizuj cache
		vm.UpdateCache(inputName, volumeDb)

		if vm.SocketHandler != nil {
			vm.SocketHandler.Server.BroadcastToNamespace("/", "volume_changed", map[string]interface{}{
				"source_name": inputName,
				"volume_db":   volumeDb,
			})
		}
	})

	log.Println("Volume Monitor started successfully")
}

// RegisterOurChange zapisuje że MY zmieniamy głośność (ignoruj nadchodzący event)
func (vm *VolumeMonitor) RegisterOurChange(sourceName string, volumeDb float64) {
	vm.ourChangesMux.Lock()
	vm.ourChanges[sourceName] = volumeDb
	vm.ourChangesMux.Unlock()

	// Auto-clear po 500ms (na wypadek gdyby event z OBS się zgubił)
	go func() {
		time.Sleep(500 * time.Millisecond)
		vm.clearOurChange(sourceName)
	}()
}

// isOurChange sprawdza czy to nasza zmiana
func (vm *VolumeMonitor) isOurChange(sourceName string, volumeDb float64) bool {
	vm.ourChangesMux.RLock()
	defer vm.ourChangesMux.RUnlock()

	expectedVolume, exists := vm.ourChanges[sourceName]
	if !exists {
		return false
	}

	// Porównaj z tolerancją (OBS może zaokrąglić)
	diff := volumeDb - expectedVolume
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.1 // Tolerancja 0.1 dB
}

// clearOurChange usuwa zarejestrowaną zmianę
func (vm *VolumeMonitor) clearOurChange(sourceName string) {
	vm.ourChangesMux.Lock()
	defer vm.ourChangesMux.Unlock()
	delete(vm.ourChanges, sourceName)
}

// GetCachedVolume pobiera głośność z cache (jeśli istnieje)
func (vm *VolumeMonitor) GetCachedVolume(sourceName string) (float64, bool) {
	vm.cacheMu.RLock()
	defer vm.cacheMu.RUnlock()

	volume, exists := vm.cachedVolumes[sourceName]
	return volume, exists
}

// UpdateCache aktualizuje cache głośności
func (vm *VolumeMonitor) UpdateCache(sourceName string, volumeDb float64) {
	vm.cacheMu.Lock()
	defer vm.cacheMu.Unlock()

	vm.cachedVolumes[sourceName] = volumeDb
}

// Stop zatrzymuje volume monitor
func (vm *VolumeMonitor) Stop() {
	log.Println("Stopping Volume Monitor...")
	// Opcjonalnie: wyczyść cache, zamknij kanały itp.
}
