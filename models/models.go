package models

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

// Season reprezentuje sezon audycji
type Season struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Number      int       `gorm:"uniqueIndex;not null" json:"number"` // Numer sezonu
	Description string    `gorm:"type:text" json:"description"`
	IsCurrent   bool      `gorm:"default:false;index" json:"is_current"` // Czy to aktualny sezon
	Episodes    []Episode `gorm:"foreignKey:SeasonID" json:"episodes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Episode reprezentuje odcinek audycji
type Episode struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	SeasonID      uint           `gorm:"index;not null" json:"season_id"`
	Season        Season         `gorm:"foreignKey:SeasonID" json:"season"`
	EpisodeNumber int            `gorm:"not null" json:"episode_number"` // Ogólny numer odcinka (ciągły)
	SeasonEpisode int            `gorm:"not null" json:"season_episode"` // Numer w sezonie
	Title         string         `gorm:"size:300;not null" json:"title"` // Tytuł odcinka
	EpisodeDate   time.Time      `json:"episode_date"`
	IsCurrent     bool           `gorm:"default:false;index" json:"is_current"` // Czy to aktualny odcinek
	Staff         []EpisodeStaff `gorm:"foreignKey:EpisodeID" json:"staff"`
	Guests        []EpisodeGuest `gorm:"foreignKey:EpisodeID" json:"guests"`
	MediaGroups   []MediaGroup   `gorm:"foreignKey:EpisodeID" json:"media_groups"`
	Media         []EpisodeMedia `gorm:"foreignKey:EpisodeID;constraint:OnDelete:CASCADE" json:"media"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// StaffType reprezentuje typ członka ekipy
type StaffType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;uniqueIndex;not null" json:"name"` // np. "Redaktor prowadzący", "Realizator dźwięku"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Staff reprezentuje członka ekipy
type Staff struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FirstName string    `gorm:"size:100;not null" json:"first_name"`
	LastName  string    `gorm:"size:100;not null" json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EpisodeStaff reprezentuje przypisanie członka ekipy do odcinka
type EpisodeStaff struct {
	ID         uint               `gorm:"primaryKey" json:"id"`
	EpisodeID  uint               `gorm:"index;not null" json:"episode_id"`
	Episode    Episode            `gorm:"foreignKey:EpisodeID" json:"episode"`
	StaffID    uint               `gorm:"index;not null" json:"staff_id"`
	Staff      Staff              `gorm:"foreignKey:StaffID" json:"staff"`
	StaffTypes []EpisodeStaffType `gorm:"foreignKey:EpisodeStaffID" json:"staff_types"` // Wiele typów dla tego przypisania
	CreatedAt  time.Time          `json:"created_at"`
}

// EpisodeStaffType reprezentuje przypisanie typu do członka ekipy w kontekście odcinka
type EpisodeStaffType struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	EpisodeStaffID uint         `gorm:"index;not null" json:"episode_staff_id"`
	EpisodeStaff   EpisodeStaff `gorm:"foreignKey:EpisodeStaffID" json:"episode_staff"`
	StaffTypeID    uint         `gorm:"index;not null" json:"staff_type_id"`
	StaffType      StaffType    `gorm:"foreignKey:StaffTypeID" json:"staff_type"`
	CreatedAt      time.Time    `json:"created_at"`
}

// GuestType reprezentuje typ gościa
type GuestType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;uniqueIndex;not null" json:"name"` // np. "Ekspert", "Artysta", "Polityk"
	Guests    []Guest   `gorm:"foreignKey:GuestTypeID" json:"guests"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Guest reprezentuje gościa
type Guest struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	GuestTypeID *uint      `gorm:"index" json:"guest_type_id"`
	GuestType   *GuestType `gorm:"foreignKey:GuestTypeID" json:"guest_type,omitempty"`
	FirstName   string     `gorm:"size:100;not null" json:"first_name"`
	LastName    string     `gorm:"size:100;not null" json:"last_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CameraType reprezentuje typ kamery (np. Centralna, Prowadzący, Goście, Dodatkowa)
type CameraType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Order     int       `gorm:"not null" json:"order"`                   // Kolejność wyświetlania i auto-przypisania
	IsSystem  bool      `gorm:"not null;default:false" json:"is_system"` // true = systemowy (nie można edytować/usunąć)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EpisodeGuest reprezentuje przypisanie gościa do odcinka
type EpisodeGuest struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	EpisodeID    uint       `gorm:"index;not null" json:"episode_id"`
	Episode      Episode    `gorm:"foreignKey:EpisodeID" json:"episode"`
	GuestID      uint       `gorm:"index;not null" json:"guest_id"`
	Guest        Guest      `gorm:"foreignKey:GuestID" json:"guest"`
	GuestTypeID  *uint      `gorm:"index" json:"guest_type_id"`                         // ← DODAJ
	GuestType    *GuestType `gorm:"foreignKey:GuestTypeID" json:"guest_type,omitempty"` // ← DODAJ
	SegmentOrder int        `json:"segment_order"`
	CreatedAt    time.Time  `json:"created_at"`
}

// MediaGroup reprezentuje grupę mediów dla odcinka
// Każdy odcinek ma dwie grupy systemowe: MEDIA i REPORTAZE
// Użytkownik może tworzyć dodatkowe grupy
type MediaGroup struct {
	ID             uint                `gorm:"primaryKey" json:"id"`
	EpisodeID      uint                `gorm:"index;not null" json:"episode_id"`
	Episode        Episode             `gorm:"foreignKey:EpisodeID" json:"episode"`
	Name           string              `gorm:"size:200;not null" json:"name"`
	Description    string              `gorm:"type:text" json:"description"`
	Order          int                 `gorm:"not null" json:"order"`          // Kolejność w odcinku
	IsSystem       bool                `gorm:"default:false" json:"is_system"` // true dla grup MEDIA i REPORTAZE
	CurrentInScene *uint               `gorm:"index" json:"current_in_scene"`  // NULL = nieużywana, 0 = w obu scenach, scene_id = w konkretnej scenie
	CurrentScene   *Scene              `gorm:"foreignKey:CurrentInScene" json:"current_scene,omitempty"`
	MediaItems     []EpisodeMediaGroup `gorm:"foreignKey:MediaGroupID" json:"media_items"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// EpisodeMediaGroup reprezentuje przypisanie media do grupy
// CurrentInScene określa gdzie plik jest aktywny:
// NULL = nieaktywny w żadnej scenie
// 0 = aktywny w obu scenach (MEDIA i REPORTAZE)
// scene_id = aktywny tylko w tej konkretnej scenie
type EpisodeMediaGroup struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	EpisodeMediaID uint         `gorm:"index;not null" json:"episode_media_id"`
	EpisodeMedia   EpisodeMedia `gorm:"foreignKey:EpisodeMediaID" json:"episode_media"`
	MediaGroupID   uint         `gorm:"index;not null" json:"media_group_id"`
	MediaGroup     MediaGroup   `gorm:"foreignKey:MediaGroupID" json:"media_group"`
	Order          int          `gorm:"not null" json:"order"`         // Kolejność w grupie
	CurrentInScene *uint        `gorm:"index" json:"current_in_scene"` // NULL = nieaktywny, 0 = w obu scenach, scene_id = w konkretnej scenie
	CurrentScene   *Scene       `gorm:"foreignKey:CurrentInScene" json:"current_scene,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

// Scene reprezentuje scenę OBS
type Scene struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;uniqueIndex;not null" json:"name"` // Nazwa sceny w OBS
	Sources   []Source  `gorm:"foreignKey:SceneID" json:"sources"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Source reprezentuje źródło w scenie OBS
type Source struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SceneID     uint      `gorm:"index;not null" json:"scene_id"`
	Scene       Scene     `gorm:"foreignKey:SceneID" json:"scene"`
	Name        string    `gorm:"size:200;not null" json:"name"`                          // Nazwa źródła w OBS
	SourceType  string    `gorm:"size:100;not null;default:'UNKNOWN'" json:"source_type"` // Typ źródła z OBS
	SourceOrder int       `gorm:"default:0" json:"source_order"`                          // Domyślna kolejność
	IsVisible   bool      `gorm:"default:false" json:"is_visible"`                        // Stan użytkownika (dla mikrofonów)
	IconURL     *string   `gorm:"size:500" json:"icon_url"`                               // Ikona dla przycisku (nullable)
	Color       string    `gorm:"size:20" json:"color"`                                   // Kolor przycisku (hex)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EpisodeSource reprezentuje przypisanie źródła do odcinka
// Uniwersalny model obsługujący różne typy źródeł:
// - Media1/Reportaze1: MediaID (konkretny plik)
// - Media2/Reportaze2: GroupID (grupa plików)
// - Mikrofony: StaffID lub GuestID (osoba)
// - Kamery: PresetID (preset kamery) - przyszłość
type EpisodeSource struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	EpisodeID  uint    `gorm:"index;not null" json:"episode_id"`
	Episode    Episode `gorm:"foreignKey:EpisodeID" json:"episode"`
	SourceName string  `gorm:"size:200;not null" json:"source_name"` // "Media1", "Media2", "Kamera1", "Mikrofon1"

	// Przypisania (tylko jedno wypełnione w zależności od typu źródła)
	MediaID      *uint `gorm:"index" json:"media_id"`       // Dla Media1/Reportaze1 → ID pliku
	GroupID      *uint `gorm:"index" json:"group_id"`       // Dla Media2/Reportaze2 → ID grupy
	CameraTypeID *uint `gorm:"index" json:"camera_type_id"` // Dla kamer → ID typu kamery
	StaffID      *uint `gorm:"index" json:"staff_id"`       // Dla mikrofonów → ID prowadzącego
	GuestID      *uint `gorm:"index" json:"guest_id"`       // Dla mikrofonów → ID gościa

	// Relacje
	CameraType *CameraType `gorm:"foreignKey:CameraTypeID" json:"camera_type"`

	// Metadane
	AssignedBy string `gorm:"size:50;default:'auto'" json:"assigned_by"` // "auto" lub "manual"

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EpisodeMedia reprezentuje media (reportaże, filmy) przypisane do odcinka
type EpisodeMedia struct {
	ID             uint                `gorm:"primaryKey" json:"id"`
	EpisodeID      uint                `gorm:"index;not null" json:"episode_id"`
	Episode        Episode             `gorm:"foreignKey:EpisodeID" json:"episode"`
	EpisodeStaffID *uint               `gorm:"index" json:"episode_staff_id"`                  // Autor z przypisanej ekipy (nullable)
	EpisodeStaff   *EpisodeStaff       `gorm:"foreignKey:EpisodeStaffID" json:"episode_staff"` // Autor z przypisanej ekipy (nullable)
	Title          string              `gorm:"size:300;not null" json:"title"`
	Description    string              `gorm:"type:text" json:"description"`
	FilePath       *string             `gorm:"size:1000" json:"file_path"`                    // Ścieżka do pliku (nullable)
	URL            *string             `gorm:"size:1000" json:"url"`                          // URL jeśli zewnętrzne (nullable)
	Duration       int                 `json:"duration"`                                      // Czas trwania w sekundach
	MediaGroups    []EpisodeMediaGroup `gorm:"foreignKey:EpisodeMediaID" json:"media_groups"` // Przynależność do grup
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// InitDB inicjalizuje bazę danych
func InitDB(db *gorm.DB) error {
	err := db.AutoMigrate(
		&Season{},
		&Episode{},
		&StaffType{},
		&Staff{},
		&EpisodeStaff{},
		&EpisodeStaffType{},
		&GuestType{},
		&Guest{},
		&EpisodeGuest{},
		&CameraType{}, // NOWE: typy kamer
		&MediaGroup{},
		&Scene{},
		&Source{},
		&EpisodeSource{}, // NOWE: tabela pomostowa episode-source
		&EpisodeMedia{},
		&EpisodeMediaGroup{},
	)

	if err != nil {
		return err
	}

	// Seed systemowych typów kamer
	if err := SeedCameraTypes(db); err != nil {
		return err
	}

	// Ustaw domyślny typ dla istniejących źródeł (migracja)
	db.Exec("UPDATE sources SET source_type = 'UNKNOWN' WHERE source_type = '' OR source_type IS NULL")

	// Dodaj unikalny indeks na parę (episode_id, source_name) dla episode_sources
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_episode_source ON episode_sources(episode_id, source_name)")

	// Załaduj dane testowe jeśli baza jest pusta i plik test-data.json istnieje
	if err := LoadTestDataIfEmpty(db); err != nil {
		return err
	}

	return nil
}

// GetCurrentSeason pobiera aktualny sezon
func GetCurrentSeason(db *gorm.DB) (*Season, error) {
	var season Season
	result := db.Where("is_current = ?", true).First(&season)
	if result.Error != nil {
		return nil, result.Error
	}
	return &season, nil
}

// SetCurrentSeason ustawia wskazany sezon jako aktualny i wyłącza pozostałe
func SetCurrentSeason(db *gorm.DB, seasonID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Sprawdź czy sezon istnieje
		var season Season
		if err := tx.First(&season, seasonID).Error; err != nil {
			return err
		}

		// Wyłącz wszystkie sezony
		if err := tx.Model(&Season{}).Where("is_current = ?", true).Update("is_current", false).Error; err != nil {
			return err
		}

		// Włącz wybrany sezon
		if err := tx.Model(&season).Update("is_current", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateSeasonAsCurrent tworzy nowy sezon i ustawia go jako aktualny
func CreateSeasonAsCurrent(db *gorm.DB, season *Season) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Wyłącz wszystkie sezony
		if err := tx.Model(&Season{}).Where("is_current = ?", true).Update("is_current", false).Error; err != nil {
			return err
		}

		// Utwórz nowy sezon jako aktualny
		season.IsCurrent = true
		if err := tx.Create(season).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetCurrentEpisode pobiera aktualny odcinek
func GetCurrentEpisode(db *gorm.DB) (*Episode, error) {
	var episode Episode
	result := db.Preload("Season").Where("is_current = ?", true).First(&episode)
	if result.Error != nil {
		return nil, result.Error
	}
	return &episode, nil
}

// SetCurrentEpisode ustawia wskazany odcinek jako aktualny i wyłącza pozostałe
func SetCurrentEpisode(db *gorm.DB, episodeID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Sprawdź czy odcinek istnieje
		var episode Episode
		if err := tx.First(&episode, episodeID).Error; err != nil {
			return err
		}

		// Wyłącz wszystkie odcinki
		if err := tx.Model(&Episode{}).Where("is_current = ?", true).Update("is_current", false).Error; err != nil {
			return err
		}

		// Włącz wybrany odcinek
		if err := tx.Model(&episode).Update("is_current", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateEpisodeAsCurrent tworzy nowy odcinek i ustawia go jako aktualny
// Automatycznie tworzy dwie grupy systemowe: MEDIA i REPORTAZE
func CreateEpisodeAsCurrent(db *gorm.DB, episode *Episode) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Wyłącz wszystkie odcinki
		if err := tx.Model(&Episode{}).Where("is_current = ?", true).Update("is_current", false).Error; err != nil {
			return err
		}

		// Utwórz nowy odcinek jako aktualny
		episode.IsCurrent = true
		if err := tx.Create(episode).Error; err != nil {
			return err
		}

		// Utwórz grupy systemowe MEDIA i REPORTAZE
		systemGroups := []MediaGroup{
			{
				EpisodeID:   episode.ID,
				Name:        "MEDIA",
				Description: "Grupa systemowa dla mediów",
				Order:       0,
				IsSystem:    true,
			},
			{
				EpisodeID:   episode.ID,
				Name:        "REPORTAZE",
				Description: "Grupa systemowa dla reportaży",
				Order:       1,
				IsSystem:    true,
			},
		}

		for _, group := range systemGroups {
			if err := tx.Create(&group).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// CreateSystemMediaGroupsForEpisode tworzy grupy systemowe dla istniejącego odcinka
// Używane przy migracji lub jeśli grupy nie zostały utworzone
func CreateSystemMediaGroupsForEpisode(db *gorm.DB, episodeID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Sprawdź czy odcinek istnieje
		var episode Episode
		if err := tx.First(&episode, episodeID).Error; err != nil {
			return err
		}

		// Sprawdź czy grupy systemowe już istnieją
		var existingGroups []MediaGroup
		tx.Where("episode_id = ? AND is_system = ?", episodeID, true).Find(&existingGroups)

		hasMedia := false
		hasReportaze := false
		for _, g := range existingGroups {
			if g.Name == "MEDIA" {
				hasMedia = true
			}
			if g.Name == "REPORTAZE" {
				hasReportaze = true
			}
		}

		// Utwórz brakujące grupy systemowe
		if !hasMedia {
			group := MediaGroup{
				EpisodeID:   episodeID,
				Name:        "MEDIA",
				Description: "Grupa systemowa dla mediów",
				Order:       0,
				IsSystem:    true,
			}
			if err := tx.Create(&group).Error; err != nil {
				return err
			}
		}

		if !hasReportaze {
			group := MediaGroup{
				EpisodeID:   episodeID,
				Name:        "REPORTAZE",
				Description: "Grupa systemowa dla reportaży",
				Order:       1,
				IsSystem:    true,
			}
			if err := tx.Create(&group).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetMediaScenes zwraca sceny MEDIA i REPORTAZE
func GetMediaScenes(db *gorm.DB) ([]Scene, error) {
	var scenes []Scene
	result := db.Where("name IN ?", []string{"MEDIA", "REPORTAZE"}).
		Preload("Sources").
		Find(&scenes)

	if result.Error != nil {
		return nil, result.Error
	}
	return scenes, nil
}

// GetMediaSceneByName zwraca scenę po nazwie (MEDIA lub REPORTAZE)
func GetMediaSceneByName(db *gorm.DB, name string) (*Scene, error) {
	var scene Scene
	result := db.Where("name = ?", name).
		Preload("Sources").
		First(&scene)

	if result.Error != nil {
		return nil, result.Error
	}
	return &scene, nil
}

// SetCurrentMediaInGroup ustawia media jako aktywne w danej scenie lub w obu scenach
// sceneID: konkretne ID sceny, lub 0 dla obu scen
// Wyłącza inne media z tej samej grupy w tej scenie/scenach
// WAŻNE: Jeśli stare media miało current_in_scene = 0 (obie sceny) i zmieniamy tylko jedną scenę,
// to stare media zostaje z current_in_scene ustawionym na drugą scenę (split)
func SetCurrentMediaInGroup(db *gorm.DB, groupID uint, mediaID uint, sceneID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Sprawdź czy przypisanie istnieje
		var assignment EpisodeMediaGroup
		if err := tx.Where("media_group_id = ? AND episode_media_id = ?", groupID, mediaID).First(&assignment).Error; err != nil {
			return err
		}

		if sceneID == 0 {
			// Ustawiamy media w obu scenach - wyłącz wszystkie inne media w grupie
			if err := tx.Model(&EpisodeMediaGroup{}).
				Where("media_group_id = ? AND id != ? AND current_in_scene IS NOT NULL", groupID, assignment.ID).
				Update("current_in_scene", nil).Error; err != nil {
				return err
			}
		} else {
			// Szukamy media które jest aktywne w obu scenach (current_in_scene = 0)
			var mediaInBothScenes EpisodeMediaGroup
			err := tx.Where("media_group_id = ? AND current_in_scene = 0", groupID).First(&mediaInBothScenes).Error

			if err == nil && mediaInBothScenes.ID != assignment.ID {
				// Znaleziono media aktywne w obu scenach - wykonaj split
				// Pobierz ID drugiej sceny (MEDIA i REPORTAZE)
				var scenes []Scene
				if err := tx.Where("name IN ?", []string{"MEDIA", "REPORTAZE"}).Find(&scenes).Error; err != nil {
					return err
				}

				var otherSceneID uint
				for _, scene := range scenes {
					if scene.ID != sceneID {
						otherSceneID = scene.ID
						break
					}
				}

				// Ustaw stare media tylko na drugą scenę
				if err := tx.Model(&mediaInBothScenes).Update("current_in_scene", otherSceneID).Error; err != nil {
					return err
				}
			}

			// Wyłącz media które są aktywne tylko w tej konkretnej scenie
			if err := tx.Model(&EpisodeMediaGroup{}).
				Where("media_group_id = ? AND id != ? AND current_in_scene = ?", groupID, assignment.ID, sceneID).
				Update("current_in_scene", nil).Error; err != nil {
				return err
			}
		}

		// Ustaw wybrane media jako aktywne
		if err := tx.Model(&assignment).Update("current_in_scene", sceneID).Error; err != nil {
			return err
		}

		return nil
	})
}

// ClearCurrentMediaInGroup wyłącza aktywne media w grupie dla danej sceny lub obu scen
// sceneID: konkretne ID sceny, lub 0 dla obu scen
func ClearCurrentMediaInGroup(db *gorm.DB, groupID uint, sceneID uint) error {
	if sceneID == 0 {
		// Wyłącz wszystkie media w grupie
		return db.Model(&EpisodeMediaGroup{}).
			Where("media_group_id = ? AND current_in_scene IS NOT NULL", groupID).
			Update("current_in_scene", nil).Error
	}

	// Wyłącz media w konkretnej scenie lub te co są w obu scenach
	return db.Model(&EpisodeMediaGroup{}).
		Where("media_group_id = ? AND (current_in_scene = ? OR current_in_scene = 0)", groupID, sceneID).
		Update("current_in_scene", nil).Error
}

// GetCurrentMediaInGroup pobiera aktywne media w grupie dla danej sceny
// Zwraca media które są aktywne w tej scenie lub w obu scenach (current_in_scene = 0)
func GetCurrentMediaInGroup(db *gorm.DB, groupID uint, sceneID uint) (*EpisodeMediaGroup, error) {
	var assignment EpisodeMediaGroup

	if sceneID == 0 {
		// Szukaj media aktywnego w obu scenach
		result := db.Preload("EpisodeMedia").
			Preload("MediaGroup").
			Where("media_group_id = ? AND current_in_scene = 0", groupID).
			First(&assignment)

		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		// Szukaj media aktywnego w tej konkretnej scenie lub w obu scenach
		result := db.Preload("EpisodeMedia").
			Preload("MediaGroup").
			Where("media_group_id = ? AND (current_in_scene = ? OR current_in_scene = 0)", groupID, sceneID).
			First(&assignment)

		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &assignment, nil
}

// GetNextMediaGroupOrder zwraca następny dostępny numer kolejności dla grupy w odcinku
func GetNextMediaGroupOrder(db *gorm.DB, episodeID uint) int {
	var maxGroup MediaGroup
	result := db.Where("episode_id = ?", episodeID).Order("\"order\" DESC").First(&maxGroup)
	if result.Error != nil {
		return 0
	}
	return maxGroup.Order + 1
}

// GetNextMediaGroupItemOrder zwraca następny dostępny numer kolejności dla media w grupie
func GetNextMediaGroupItemOrder(db *gorm.DB, groupID uint) int {
	var maxItem EpisodeMediaGroup
	result := db.Where("media_group_id = ?", groupID).Order("\"order\" DESC").First(&maxItem)
	if result.Error != nil {
		return 0
	}
	return maxItem.Order + 1
}

// SetCurrentMediaGroup ustawia grupę jako aktywną w danej scenie lub w obu scenach
// sceneID: konkretne ID sceny, lub 0 dla obu scen
// Wyłącza inne grupy w tym odcinku dla tej sceny/scen
// WAŻNE: Jeśli stara grupa miała current_in_scene = 0 (obie sceny) i zmieniamy tylko jedną scenę,
// to stara grupa zostaje z current_in_scene ustawionym na drugą scenę (split)
func SetCurrentMediaGroup(db *gorm.DB, episodeID uint, groupID uint, sceneID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Sprawdź czy grupa istnieje
		var group MediaGroup
		if err := tx.First(&group, groupID).Error; err != nil {
			return err
		}

		if sceneID == 0 {
			// Ustawiamy grupę w obu scenach - wyłącz wszystkie inne grupy
			if err := tx.Model(&MediaGroup{}).
				Where("episode_id = ? AND id != ? AND current_in_scene IS NOT NULL", episodeID, groupID).
				Update("current_in_scene", nil).Error; err != nil {
				return err
			}
		} else {
			// Szukamy grupy która jest aktywna w obu scenach (current_in_scene = 0)
			var groupInBothScenes MediaGroup
			err := tx.Where("episode_id = ? AND current_in_scene = 0", episodeID).First(&groupInBothScenes).Error

			if err == nil && groupInBothScenes.ID != groupID {
				// Znaleziono grupę aktywną w obu scenach - wykonaj split
				// Pobierz ID drugiej sceny (MEDIA i REPORTAZE)
				var scenes []Scene
				if err := tx.Where("name IN ?", []string{"MEDIA", "REPORTAZE"}).Find(&scenes).Error; err != nil {
					return err
				}

				var otherSceneID uint
				for _, scene := range scenes {
					if scene.ID != sceneID {
						otherSceneID = scene.ID
						break
					}
				}

				// Ustaw starą grupę tylko na drugą scenę
				if err := tx.Model(&groupInBothScenes).Update("current_in_scene", otherSceneID).Error; err != nil {
					return err
				}
			}

			// Wyłącz grupy które są aktywne tylko w tej konkretnej scenie
			if err := tx.Model(&MediaGroup{}).
				Where("episode_id = ? AND id != ? AND current_in_scene = ?", episodeID, groupID, sceneID).
				Update("current_in_scene", nil).Error; err != nil {
				return err
			}
		}

		// Ustaw wybraną grupę jako aktywną
		if err := tx.Model(&group).Update("current_in_scene", sceneID).Error; err != nil {
			return err
		}

		return nil
	})
}

// ClearCurrentMediaGroup wyłącza aktywną grupę w odcinku dla danej sceny lub obu scen
// sceneID: konkretne ID sceny, lub 0 dla obu scen
func ClearCurrentMediaGroup(db *gorm.DB, episodeID uint, sceneID uint) error {
	if sceneID == 0 {
		// Wyłącz wszystkie grupy w odcinku
		return db.Model(&MediaGroup{}).
			Where("episode_id = ? AND current_in_scene IS NOT NULL", episodeID).
			Update("current_in_scene", nil).Error
	}

	// Wyłącz grupy w konkretnej scenie lub te co są w obu scenach
	return db.Model(&MediaGroup{}).
		Where("episode_id = ? AND (current_in_scene = ? OR current_in_scene = 0)", episodeID, sceneID).
		Update("current_in_scene", nil).Error
}

// GetCurrentMediaGroup pobiera aktywną grupę w odcinku dla danej sceny
// Zwraca grupę która jest aktywna w tej scenie lub w obu scenach (current_in_scene = 0)
func GetCurrentMediaGroup(db *gorm.DB, episodeID uint, sceneID uint) (*MediaGroup, error) {
	var group MediaGroup

	if sceneID == 0 {
		// Szukaj grupy aktywnej w obu scenach
		result := db.Preload("Episode").
			Where("episode_id = ? AND current_in_scene = 0", episodeID).
			First(&group)

		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		// Szukaj grupy aktywnej w tej konkretnej scenie lub w obu scenach
		result := db.Preload("Episode").
			Where("episode_id = ? AND (current_in_scene = ? OR current_in_scene = 0)", episodeID, sceneID).
			First(&group)

		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &group, nil
}

// ===== FUNKCJE DLA EPISODE_SOURCE =====

// SetEpisodeSourceMedia ustawia przypisanie pliku media do źródła
func SetEpisodeSourceMedia(db *gorm.DB, episodeID uint, sourceName string, mediaID uint, assignedBy string) error {
	var episodeSource EpisodeSource

	// Sprawdź czy wpis już istnieje
	result := db.Where("episode_id = ? AND source_name = ?", episodeID, sourceName).First(&episodeSource)

	if result.Error == gorm.ErrRecordNotFound {
		// Utwórz nowy wpis
		episodeSource = EpisodeSource{
			EpisodeID:  episodeID,
			SourceName: sourceName,
			MediaID:    &mediaID,
			AssignedBy: assignedBy,
		}
		return db.Create(&episodeSource).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Zaktualizuj istniejący wpis
	episodeSource.MediaID = &mediaID
	episodeSource.GroupID = nil // Wyczyść poprzednie przypisanie grupy (jeśli było)
	episodeSource.AssignedBy = assignedBy
	return db.Save(&episodeSource).Error
}

// SetEpisodeSourceGroup ustawia przypisanie grupy do źródła VLC
func SetEpisodeSourceGroup(db *gorm.DB, episodeID uint, sourceName string, groupID uint, assignedBy string) error {
	var episodeSource EpisodeSource

	// Sprawdź czy wpis już istnieje
	result := db.Where("episode_id = ? AND source_name = ?", episodeID, sourceName).First(&episodeSource)

	if result.Error == gorm.ErrRecordNotFound {
		// Utwórz nowy wpis
		episodeSource = EpisodeSource{
			EpisodeID:  episodeID,
			SourceName: sourceName,
			GroupID:    &groupID,
			AssignedBy: assignedBy,
		}
		return db.Create(&episodeSource).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Zaktualizuj istniejący wpis
	episodeSource.GroupID = &groupID
	episodeSource.MediaID = nil // Wyczyść poprzednie przypisanie media (jeśli było)
	episodeSource.AssignedBy = assignedBy
	return db.Save(&episodeSource).Error
}

// SetEpisodeSourceCameraType ustawia przypisanie typu kamery do źródła
func SetEpisodeSourceCameraType(db *gorm.DB, episodeID uint, sourceName string, cameraTypeID uint, assignedBy string) error {
	var episodeSource EpisodeSource

	// Sprawdź czy wpis już istnieje
	result := db.Where("episode_id = ? AND source_name = ?", episodeID, sourceName).First(&episodeSource)

	if result.Error == gorm.ErrRecordNotFound {
		// Utwórz nowy wpis
		episodeSource = EpisodeSource{
			EpisodeID:    episodeID,
			SourceName:   sourceName,
			CameraTypeID: &cameraTypeID,
			AssignedBy:   assignedBy,
		}
		return db.Create(&episodeSource).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Zaktualizuj istniejący wpis
	episodeSource.CameraTypeID = &cameraTypeID
	episodeSource.AssignedBy = assignedBy
	return db.Save(&episodeSource).Error
}

// DisableEpisodeSourceCamera wyłącza kamerę (CameraTypeID=NULL, AssignedBy=manual)
func DisableEpisodeSourceCamera(db *gorm.DB, episodeID uint, sourceName string) error {
	var episodeSource EpisodeSource

	// Sprawdź czy wpis już istnieje
	result := db.Where("episode_id = ? AND source_name = ?", episodeID, sourceName).First(&episodeSource)

	if result.Error == gorm.ErrRecordNotFound {
		// Utwórz nowy wpis jako wyłączony
		episodeSource = EpisodeSource{
			EpisodeID:    episodeID,
			SourceName:   sourceName,
			CameraTypeID: nil,
			AssignedBy:   "manual",
		}
		return db.Create(&episodeSource).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Zaktualizuj istniejący wpis - wyłącz
	episodeSource.CameraTypeID = nil
	episodeSource.AssignedBy = "manual"
	return db.Save(&episodeSource).Error
}

// SetEpisodeSourceMicrophone ustawia przypisanie osoby (staff/guest) do mikrofonu
func SetEpisodeSourceMicrophone(db *gorm.DB, episodeID uint, sourceName string, personID uint, personType string, assignedBy string) error {
	var episodeSource EpisodeSource

	// Sprawdź czy wpis już istnieje
	result := db.Where("episode_id = ? AND source_name = ?", episodeID, sourceName).First(&episodeSource)

	if result.Error == gorm.ErrRecordNotFound {
		// Utwórz nowy wpis
		episodeSource = EpisodeSource{
			EpisodeID:  episodeID,
			SourceName: sourceName,
			AssignedBy: assignedBy,
		}

		if personType == "staff" {
			episodeSource.StaffID = &personID
			episodeSource.GuestID = nil
		} else {
			episodeSource.GuestID = &personID
			episodeSource.StaffID = nil
		}

		return db.Create(&episodeSource).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Zaktualizuj istniejący wpis
	if personType == "staff" {
		episodeSource.StaffID = &personID
		episodeSource.GuestID = nil
	} else {
		episodeSource.GuestID = &personID
		episodeSource.StaffID = nil
	}
	episodeSource.AssignedBy = assignedBy
	return db.Save(&episodeSource).Error
}

// UnassignEpisodeSourceMicrophone usuwa przypisanie osoby z mikrofonu
func UnassignEpisodeSourceMicrophone(db *gorm.DB, episodeID uint, sourceName string) error {
	var episodeSource EpisodeSource

	// Sprawdź czy wpis już istnieje
	result := db.Where("episode_id = ? AND source_name = ?", episodeID, sourceName).First(&episodeSource)

	if result.Error == gorm.ErrRecordNotFound {
		// Jeśli nie istnieje, nie ma co usuwać
		return nil
	}

	if result.Error != nil {
		return result.Error
	}

	// Wyczyść przypisanie
	episodeSource.StaffID = nil
	episodeSource.GuestID = nil
	episodeSource.AssignedBy = "manual"
	return db.Save(&episodeSource).Error
}

// GetEpisodeSourceAssignment pobiera przypisanie dla źródła
func GetEpisodeSourceAssignment(db *gorm.DB, episodeID uint, sourceName string) (*EpisodeSource, error) {
	var episodeSource EpisodeSource

	result := db.Where("episode_id = ? AND source_name = ?", episodeID, sourceName).First(&episodeSource)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &episodeSource, nil
}

// GetAllEpisodeSourceAssignments pobiera wszystkie przypisania dla odcinka
func GetAllEpisodeSourceAssignments(db *gorm.DB, episodeID uint) (map[string]interface{}, error) {
	var episodeSources []EpisodeSource

	result := db.Where("episode_id = ?", episodeID).Find(&episodeSources)

	if result.Error != nil {
		return nil, result.Error
	}

	assignments := make(map[string]interface{})

	for _, es := range episodeSources {
		if es.MediaID != nil {
			// Pobierz tytuł media
			var media EpisodeMedia
			if err := db.First(&media, *es.MediaID).Error; err == nil {
				assignments[es.SourceName] = map[string]interface{}{
					"type":        "media",
					"media_id":    *es.MediaID,
					"button_text": media.Title,
					"assigned_by": es.AssignedBy,
				}
			}
		} else if es.GroupID != nil {
			// Pobierz nazwę grupy
			var group MediaGroup
			if err := db.First(&group, *es.GroupID).Error; err == nil {
				assignments[es.SourceName] = map[string]interface{}{
					"type":        "group",
					"group_id":    *es.GroupID,
					"button_text": group.Name,
					"assigned_by": es.AssignedBy,
				}
			}
		} else if es.StaffID != nil {
			// Pobierz dane osoby
			var staff Staff
			if err := db.First(&staff, *es.StaffID).Error; err == nil {
				assignments[es.SourceName] = map[string]interface{}{
					"type":        "staff",
					"staff_id":    *es.StaffID,
					"button_text": fmt.Sprintf("%s %s", staff.FirstName, staff.LastName),
					"assigned_by": es.AssignedBy,
				}
			}
		} else if es.GuestID != nil {
			// Pobierz dane osoby
			var guest Guest
			if err := db.First(&guest, *es.GuestID).Error; err == nil {
				assignments[es.SourceName] = map[string]interface{}{
					"type":        "guest",
					"staff_id":    *es.GuestID,
					"button_text": fmt.Sprintf("%s %s", guest.FirstName, guest.LastName),
					"assigned_by": es.AssignedBy,
				}
			}
		} else if es.CameraTypeID != nil {
			// Pobierz nazwę typu kamery
			var cameraType CameraType
			if err := db.First(&cameraType, *es.CameraTypeID).Error; err == nil {
				assignments[es.SourceName] = map[string]interface{}{
					"type":             "camera",
					"camera_type_id":   *es.CameraTypeID,
					"camera_type_name": cameraType.Name,
					"button_text":      cameraType.Name,
					"assigned_by":      es.AssignedBy,
					"is_disabled":      false,
				}
			}
		} else if es.AssignedBy == "manual" && (es.SourceName == "Kamera1" || es.SourceName == "Kamera2" || es.SourceName == "Kamera3" || es.SourceName == "Kamera4") {
			// CameraTypeID=NULL + AssignedBy=manual = wyłączona kamera
			assignments[es.SourceName] = map[string]interface{}{
				"type":             "camera",
				"camera_type_id":   nil,
				"camera_type_name": nil,
				"button_text":      es.SourceName, // "Kamera1"
				"assigned_by":      "manual",
				"is_disabled":      true,
			}
		}
	}

	return assignments, nil
}

// SeedCameraTypes tworzy 4 predefiniowane systemowe typy kamer
func SeedCameraTypes(db *gorm.DB) error {
	systemTypes := []CameraType{
		{Name: "Centralna", Order: 1, IsSystem: true},
		{Name: "Prowadzący", Order: 2, IsSystem: true},
		{Name: "Goście", Order: 3, IsSystem: true},
		{Name: "Dodatkowa", Order: 4, IsSystem: true},
	}

	for _, cameraType := range systemTypes {
		var existing CameraType
		result := db.Where("name = ?", cameraType.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			// Nie istnieje - utwórz
			if err := db.Create(&cameraType).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// ===================================================================
// TEST DATA LOADING
// ===================================================================

// TestData reprezentuje strukturę pliku test-data.json
type TestData struct {
	Seasons        []TestSeason        `json:"seasons"`
	Episodes       []TestEpisode       `json:"episodes"`
	StaffTypes     []TestStaffType     `json:"staff_types"`
	Staff          []TestStaff         `json:"staff"`
	EpisodeStaff   []TestEpisodeStaff  `json:"episode_staff"`
	GuestTypes     []TestGuestType     `json:"guest_types"`
	Guests         []TestGuest         `json:"guests"`
	EpisodeGuests  []TestEpisodeGuest  `json:"episode_guests"`
	Scenes         []TestScene         `json:"scenes"`
	Sources        []TestSource        `json:"sources"`
	EpisodeSources []TestEpisodeSource `json:"episode_sources"`
	EpisodeMedia   []TestEpisodeMedia  `json:"episode_media"`
}

type TestSeason struct {
	Number      int    `json:"number"`
	Description string `json:"description"`
	IsCurrent   bool   `json:"is_current"`
}

type TestEpisode struct {
	SeasonID      int    `json:"season_id"`
	EpisodeNumber int    `json:"episode_number"`
	SeasonEpisode int    `json:"season_episode"`
	Title         string `json:"title"`
	EpisodeDate   string `json:"episode_date"`
	IsCurrent     bool   `json:"is_current"`
}

type TestStaffType struct {
	Name string `json:"name"`
}

type TestStaff struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TestEpisodeStaff struct {
	EpisodeID    int   `json:"episode_id"`
	StaffID      int   `json:"staff_id"`
	StaffTypeIDs []int `json:"staff_type_ids"`
}

type TestGuestType struct {
	Name string `json:"name"`
}

type TestGuest struct {
	GuestTypeID int    `json:"guest_type_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

type TestEpisodeGuest struct {
	EpisodeID    int `json:"episode_id"`
	GuestID      int `json:"guest_id"`
	GuestTypeID  int `json:"guest_type_id"` // ← DODAJ
	SegmentOrder int `json:"segment_order"`
}

type TestScene struct {
	Name string `json:"name"`
}

type TestSource struct {
	SceneName  string `json:"scene_name"`
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
	Order      int    `json:"order"`
	IsVisible  bool   `json:"is_visible"`
}

type TestEpisodeSource struct {
	EpisodeID    int    `json:"episode_id"`
	SourceName   string `json:"source_name"`
	MediaID      *int   `json:"media_id"`
	GroupID      *int   `json:"group_id"`
	CameraTypeID *int   `json:"camera_type_id"`
	StaffID      *int   `json:"staff_id"`
	GuestID      *int   `json:"guest_id"`
	AssignedBy   string `json:"assigned_by"`
}

type TestEpisodeMedia struct {
	EpisodeID      int                     `json:"episode_id"`
	EpisodeStaffID *int                    `json:"episode_staff_id"`
	Title          string                  `json:"title"`
	Description    string                  `json:"description"`
	FilePath       string                  `json:"file_path"`
	URL            string                  `json:"url"`
	Duration       int                     `json:"duration"`
	Groups         []TestEpisodeMediaGroup `json:"groups"`
}

type TestEpisodeMediaGroup struct {
	GroupID int `json:"group_id"`
	Order   int `json:"order"`
}

// LoadTestDataIfEmpty ładuje dane testowe z pliku test-data.json jeśli baza jest pusta
func LoadTestDataIfEmpty(db *gorm.DB) error {
	// Sprawdź czy baza jest pusta (brak sezonów)
	var count int64
	if err := db.Model(&Season{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		// Baza nie jest pusta - pomiń ładowanie
		fmt.Println("Baza danych zawiera dane - pomijam ładowanie test-data.json")
		return nil
	}

	// Sprawdź czy plik test-data.json istnieje
	testDataPath := "test-data.json"
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		// Plik nie istnieje - pomiń ładowanie
		fmt.Println("Plik test-data.json nie istnieje - pomijam ładowanie danych testowych")
		return nil
	}

	fmt.Println("Wykryto pusty bazę danych i plik test-data.json - ładuję dane testowe...")

	// Wczytaj plik JSON
	data, err := os.ReadFile(testDataPath)
	if err != nil {
		return fmt.Errorf("błąd odczytu test-data.json: %v", err)
	}

	// Parsuj JSON
	var testData TestData
	if err := json.Unmarshal(data, &testData); err != nil {
		return fmt.Errorf("błąd parsowania test-data.json: %v", err)
	}

	// Załaduj dane w transakcji
	return db.Transaction(func(tx *gorm.DB) error {
		// Mapy do przechowywania ID utworzonych rekordów
		seasonIDs := make(map[int]uint)       // test season index -> real ID
		episodeIDs := make(map[int]uint)      // test episode index -> real ID
		staffTypeIDs := make(map[int]uint)    // test staff_type index -> real ID
		staffIDs := make(map[int]uint)        // test staff index -> real ID
		episodeStaffIDs := make(map[int]uint) // test episode_staff index -> real ID
		guestTypeIDs := make(map[int]uint)    // test guest_type index -> real ID
		guestIDs := make(map[int]uint)        // test guest index -> real ID
		sceneIDs := make(map[string]uint)     // scene name -> real ID
		mediaIDs := make(map[int]uint)        // test media index -> real ID

		// 1. Załaduj sezony
		for idx, testSeason := range testData.Seasons {
			season := Season{
				Number:      testSeason.Number,
				Description: testSeason.Description,
				IsCurrent:   testSeason.IsCurrent,
			}

			if err := tx.Create(&season).Error; err != nil {
				return fmt.Errorf("błąd tworzenia sezonu: %v", err)
			}

			seasonIDs[idx+1] = season.ID
			fmt.Printf("  ✓ Utworzono sezon #%d: %s (ID=%d)\n", season.Number, season.Description, season.ID)
		}

		// 2. Załaduj odcinki
		for idx, testEpisode := range testData.Episodes {
			// Parsuj datę
			episodeDate, err := time.Parse(time.RFC3339, testEpisode.EpisodeDate)
			if err != nil {
				return fmt.Errorf("błąd parsowania daty odcinka: %v", err)
			}

			episode := Episode{
				SeasonID:      seasonIDs[testEpisode.SeasonID],
				EpisodeNumber: testEpisode.EpisodeNumber,
				SeasonEpisode: testEpisode.SeasonEpisode,
				Title:         testEpisode.Title,
				EpisodeDate:   episodeDate,
				IsCurrent:     testEpisode.IsCurrent,
			}

			if err := tx.Create(&episode).Error; err != nil {
				return fmt.Errorf("błąd tworzenia odcinka: %v", err)
			}

			episodeIDs[idx+1] = episode.ID
			fmt.Printf("  ✓ Utworzono odcinek #%d: %s (ID=%d)\n", episode.EpisodeNumber, episode.Title, episode.ID)

			// Jeśli odcinek jest aktualny, utwórz grupy systemowe
			if episode.IsCurrent {
				if err := CreateSystemMediaGroupsForEpisode(tx, episode.ID); err != nil {
					return fmt.Errorf("błąd tworzenia grup systemowych: %v", err)
				}
				fmt.Printf("  ✓ Utworzono grupy systemowe dla odcinka #%d\n", episode.EpisodeNumber)
			}
		}

		// 3. Załaduj staff types
		for idx, testStaffType := range testData.StaffTypes {
			staffType := StaffType{
				Name: testStaffType.Name,
			}

			if err := tx.Create(&staffType).Error; err != nil {
				return fmt.Errorf("błąd tworzenia staff type: %v", err)
			}

			staffTypeIDs[idx+1] = staffType.ID
			fmt.Printf("  ✓ Utworzono staff type: %s (ID=%d)\n", staffType.Name, staffType.ID)
		}

		// 4. Załaduj staff
		for idx, testStaff := range testData.Staff {
			staff := Staff{
				FirstName: testStaff.FirstName,
				LastName:  testStaff.LastName,
			}

			if err := tx.Create(&staff).Error; err != nil {
				return fmt.Errorf("błąd tworzenia staff: %v", err)
			}

			staffIDs[idx+1] = staff.ID
			fmt.Printf("  ✓ Utworzono staff: %s %s (ID=%d)\n", staff.FirstName, staff.LastName, staff.ID)
		}

		// 5. Załaduj episode_staff
		for idx, testEpisodeStaff := range testData.EpisodeStaff {
			episodeStaff := EpisodeStaff{
				EpisodeID: episodeIDs[testEpisodeStaff.EpisodeID],
				StaffID:   staffIDs[testEpisodeStaff.StaffID],
			}

			if err := tx.Create(&episodeStaff).Error; err != nil {
				return fmt.Errorf("błąd tworzenia episode staff: %v", err)
			}

			episodeStaffIDs[idx+1] = episodeStaff.ID
			fmt.Printf("  ✓ Utworzono episode staff (ID=%d)\n", episodeStaff.ID)

			// Dodaj typy dla tego przypisania
			for _, staffTypeID := range testEpisodeStaff.StaffTypeIDs {
				episodeStaffType := EpisodeStaffType{
					EpisodeStaffID: episodeStaff.ID,
					StaffTypeID:    staffTypeIDs[staffTypeID],
				}

				if err := tx.Create(&episodeStaffType).Error; err != nil {
					return fmt.Errorf("błąd tworzenia episode staff type: %v", err)
				}
			}
		}

		// 6. Załaduj guest types
		for idx, testGuestType := range testData.GuestTypes {
			guestType := GuestType{
				Name: testGuestType.Name,
			}

			if err := tx.Create(&guestType).Error; err != nil {
				return fmt.Errorf("błąd tworzenia guest type: %v", err)
			}

			guestTypeIDs[idx+1] = guestType.ID
			fmt.Printf("  ✓ Utworzono guest type: %s (ID=%d)\n", guestType.Name, guestType.ID)
		}

		// 7. Załaduj guests
		for idx, testGuest := range testData.Guests {
			guest := Guest{
				FirstName: testGuest.FirstName,
				LastName:  testGuest.LastName,
			}

			// Ustaw GuestTypeID tylko jeśli jest podany w test-data.json
			if testGuest.GuestTypeID != 0 {
				typeID := guestTypeIDs[testGuest.GuestTypeID]
				guest.GuestTypeID = &typeID
			}

			if err := tx.Create(&guest).Error; err != nil {
				return fmt.Errorf("błąd tworzenia guest: %v", err)
			}

			guestIDs[idx+1] = guest.ID
			if guest.GuestTypeID != nil {
				fmt.Printf("  ✓ Utworzono guest: %s %s (Type ID=%d, ID=%d)\n", guest.FirstName, guest.LastName, *guest.GuestTypeID, guest.ID)
			} else {
				fmt.Printf("  ✓ Utworzono guest: %s %s (bez typu, ID=%d)\n", guest.FirstName, guest.LastName, guest.ID)
			}
		}

		// 8. Załaduj episode_guests
		for _, testEpisodeGuest := range testData.EpisodeGuests {
			episodeGuest := EpisodeGuest{
				EpisodeID:    episodeIDs[testEpisodeGuest.EpisodeID],
				GuestID:      guestIDs[testEpisodeGuest.GuestID],
				SegmentOrder: testEpisodeGuest.SegmentOrder,
			}

			// Ustaw GuestTypeID tylko jeśli podano
			if testEpisodeGuest.GuestTypeID != 0 {
				typeID := guestTypeIDs[testEpisodeGuest.GuestTypeID]
				episodeGuest.GuestTypeID = &typeID
			}

			if err := tx.Create(&episodeGuest).Error; err != nil {
				return fmt.Errorf("błąd tworzenia episode guest: %v", err)
			}

			fmt.Printf("  ✓ Utworzono episode guest (ID=%d)\n", episodeGuest.ID)
		}

		// 9. Załaduj scenes
		for _, testScene := range testData.Scenes {
			scene := Scene{
				Name: testScene.Name,
			}

			if err := tx.Create(&scene).Error; err != nil {
				return fmt.Errorf("błąd tworzenia scene: %v", err)
			}

			sceneIDs[scene.Name] = scene.ID
			fmt.Printf("  ✓ Utworzono scene: %s (ID=%d)\n", scene.Name, scene.ID)
		}

		// 10. Załaduj sources
		for _, testSource := range testData.Sources {
			source := Source{
				SceneID:     sceneIDs[testSource.SceneName],
				Name:        testSource.Name,
				SourceType:  testSource.SourceType,
				SourceOrder: testSource.Order,
				IsVisible:   testSource.IsVisible,
			}

			if err := tx.Create(&source).Error; err != nil {
				return fmt.Errorf("błąd tworzenia source: %v", err)
			}

			fmt.Printf("  ✓ Utworzono source: %s w scenie %s (ID=%d)\n", source.Name, testSource.SceneName, source.ID)
		}

		// 11. Załaduj episode_media
		for idx, testMedia := range testData.EpisodeMedia {
			var episodeStaffID *uint
			if testMedia.EpisodeStaffID != nil {
				id := episodeStaffIDs[*testMedia.EpisodeStaffID]
				episodeStaffID = &id
			}

			// Konwertuj string na *string
			var filePath, url *string
			if testMedia.FilePath != "" {
				filePath = &testMedia.FilePath
			}
			if testMedia.URL != "" {
				url = &testMedia.URL
			}

			media := EpisodeMedia{
				EpisodeID:      episodeIDs[testMedia.EpisodeID],
				EpisodeStaffID: episodeStaffID,
				Title:          testMedia.Title,
				Description:    testMedia.Description,
				FilePath:       filePath,
				URL:            url,
				Duration:       testMedia.Duration,
			}

			if err := tx.Create(&media).Error; err != nil {
				return fmt.Errorf("błąd tworzenia episode media: %v", err)
			}

			mediaIDs[idx+1] = media.ID
			fmt.Printf("  ✓ Utworzono episode media: %s (ID=%d)\n", media.Title, media.ID)

			// Dodaj do grup
			for _, testGroup := range testMedia.Groups {
				episodeMediaGroup := EpisodeMediaGroup{
					EpisodeMediaID: media.ID,
					MediaGroupID:   uint(testGroup.GroupID),
					Order:          testGroup.Order,
				}

				if err := tx.Create(&episodeMediaGroup).Error; err != nil {
					return fmt.Errorf("błąd tworzenia episode media group: %v", err)
				}
			}
		}

		// 12. Załaduj episode_sources
		for _, testEpisodeSource := range testData.EpisodeSources {
			var mediaID, groupID, cameraTypeID, staffID, guestID *uint

			if testEpisodeSource.MediaID != nil {
				id := mediaIDs[*testEpisodeSource.MediaID]
				mediaID = &id
			}
			if testEpisodeSource.GroupID != nil {
				id := uint(*testEpisodeSource.GroupID)
				groupID = &id
			}
			if testEpisodeSource.CameraTypeID != nil {
				id := uint(*testEpisodeSource.CameraTypeID)
				cameraTypeID = &id
			}
			if testEpisodeSource.StaffID != nil {
				id := staffIDs[*testEpisodeSource.StaffID]
				staffID = &id
			}
			if testEpisodeSource.GuestID != nil {
				id := guestIDs[*testEpisodeSource.GuestID]
				guestID = &id
			}

			episodeSource := EpisodeSource{
				EpisodeID:    episodeIDs[testEpisodeSource.EpisodeID],
				SourceName:   testEpisodeSource.SourceName,
				MediaID:      mediaID,
				GroupID:      groupID,
				CameraTypeID: cameraTypeID,
				StaffID:      staffID,
				GuestID:      guestID,
				AssignedBy:   testEpisodeSource.AssignedBy,
			}

			if err := tx.Create(&episodeSource).Error; err != nil {
				return fmt.Errorf("błąd tworzenia episode source: %v", err)
			}

			fmt.Printf("  ✓ Utworzono episode source: %s (ID=%d)\n", episodeSource.SourceName, episodeSource.ID)
		}

		fmt.Println("✓ Dane testowe załadowane pomyślnie!")
		return nil
	})
}
