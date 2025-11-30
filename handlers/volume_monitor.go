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
}

func NewVolumeMonitor(obsClient *obsws.Client, socketHandler *SocketHandler) *VolumeMonitor {
	return &VolumeMonitor{
		OBSClient:     obsClient,
		SocketHandler: socketHandler,
		ourChanges:    make(map[string]float64),
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
			return
		}

		// To zewnętrzna zmiana (z OBS UI) - broadcast do frontendu
		log.Printf("Volume changed externally: %s = %.2f dB", inputName, volumeDb)

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

// Stop zatrzymuje volume monitor
func (vm *VolumeMonitor) Stop() {
	log.Println("Stopping Volume Monitor...")
	// Opcjonalnie: wyczyść cache, zamknij kanały itp.
}
