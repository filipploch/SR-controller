package handlers

import (
	"log"
	"net/http"
	"obs-controller/models"
	"obs-controller/obsws"

	"gorm.io/gorm"
)

type InitMiddleware struct {
	DB          *gorm.DB
	OBSClient   *obsws.Client
	initialized bool
}

func NewInitMiddleware(db *gorm.DB, obsClient *obsws.Client) *InitMiddleware {
	return &InitMiddleware{
		DB:          db,
		OBSClient:   obsClient,
		initialized: false,
	}
}

// InitializeOBSData inicjalizuje dane z OBS (sceny) jeśli jeszcze nie zostały zainicjalizowane
func (m *InitMiddleware) InitializeOBSData() error {
	if m.initialized {
		return nil
	}

	log.Println("Inicjalizacja danych z OBS...")

	// Pobierz listę scen z OBS
	sceneList, err := m.OBSClient.GetSceneList()
	if err != nil {
		log.Printf("Błąd pobierania scen z OBS: %v", err)
		return err
	}

	// Dla każdej sceny, utwórz rekord jeśli nie istnieje
	for _, sceneName := range sceneList {
		var scene models.Scene
		result := m.DB.Where("name = ?", sceneName).First(&scene)
		if result.Error == gorm.ErrRecordNotFound {
			scene = models.Scene{
				Name: sceneName,
			}
			if err := m.DB.Create(&scene).Error; err != nil {
				log.Printf("Błąd tworzenia sceny %s: %v", sceneName, err)
			} else {
				log.Printf("Utworzono scenę: %s", sceneName)
			}
		}

		// Pobierz źródła dla sceny
		items, err := m.OBSClient.GetSceneItemList(sceneName)
		if err != nil {
			log.Printf("Błąd pobierania źródeł dla sceny %s: %v", sceneName, err)
			continue
		}

		// Dla każdego źródła, utwórz rekord jeśli nie istnieje
		for _, item := range items {
			sourceName, ok := item["sourceName"].(string)
			if !ok {
				continue
			}

			sceneItemIndex, _ := item["sceneItemIndex"].(float64)
			sourceType, _ := item["sourceType"].(string)

			// Pomiń źródła typu SCENE i FILTER
			if sourceType == "OBS_SOURCE_TYPE_SCENE" || sourceType == "OBS_SOURCE_TYPE_FILTER" {
				log.Printf("Pomijam źródło typu %s: %s (scena: %s)", sourceType, sourceName, sceneName)
				continue
			}

			// Jeśli brak sourceType, ustaw domyślny
			if sourceType == "" {
				sourceType = "UNKNOWN"
			}

			var source models.Source
			result := m.DB.Where("scene_id = ? AND name = ?", scene.ID, sourceName).First(&source)
			if result.Error == gorm.ErrRecordNotFound {
				source = models.Source{
					SceneID:     scene.ID,
					Name:        sourceName,
					SourceType:  sourceType,
					SourceOrder: int(sceneItemIndex),
					IsVisible:   false,
				}
				if err := m.DB.Create(&source).Error; err != nil {
					log.Printf("Błąd tworzenia źródła %s w scenie %s: %v", sourceName, sceneName, err)
				} else {
					log.Printf("Utworzono źródło: %s (typ: %s) w scenie %s", sourceName, sourceType, sceneName)
				}
			}
		}
	}

	m.initialized = true
	log.Println("Inicjalizacja danych z OBS zakończona")
	return nil
}

// CheckRequirements sprawdza czy są spełnione wymagania do uruchomienia kontrolera
func (m *InitMiddleware) CheckRequirements(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Inicjalizuj dane z OBS jeśli jeszcze nie zostały zainicjalizowane
		if err := m.InitializeOBSData(); err != nil {
			log.Printf("Błąd inicjalizacji danych OBS: %v", err)
		}

		// Sprawdź czy jest aktualny sezon
		var season models.Season
		hasCurrentSeason := m.DB.Where("is_current = ?", true).First(&season).Error == nil

		// Sprawdź czy jest aktualny odcinek
		var episode models.Episode
		hasCurrentEpisode := m.DB.Where("is_current = ?", true).First(&episode).Error == nil

		// Jeśli brakuje któregoś, przekieruj na /settings
		if !hasCurrentSeason || !hasCurrentEpisode {
			log.Printf("Przekierowanie na /settings - brak wymaganych danych (sezon: %v, odcinek: %v)",
				hasCurrentSeason, hasCurrentEpisode)
			http.Redirect(w, r, "/settings", http.StatusSeeOther)
			return
		}

		// Wszystko OK, pozwól na dostęp
		next(w, r)
	}
}
