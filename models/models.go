package models

import (
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
	ID          uint      `gorm:"primaryKey" json:"id"`
	GuestTypeID uint      `gorm:"index" json:"guest_type_id"`
	GuestType   GuestType `gorm:"foreignKey:GuestTypeID" json:"guest_type"`
	FirstName   string    `gorm:"size:100;not null" json:"first_name"`
	LastName    string    `gorm:"size:100;not null" json:"last_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EpisodeGuest reprezentuje przypisanie gościa do odcinka
type EpisodeGuest struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EpisodeID    uint      `gorm:"index;not null" json:"episode_id"`
	Episode      Episode   `gorm:"foreignKey:EpisodeID" json:"episode"`
	GuestID      uint      `gorm:"index;not null" json:"guest_id"`
	Guest        Guest     `gorm:"foreignKey:GuestID" json:"guest"`
	Topic        string    `gorm:"size:500" json:"topic"` // Temat rozmowy z tym gościem
	SegmentOrder int       `json:"segment_order"`         // Kolejność wystąpienia w odcinku
	CreatedAt    time.Time `json:"created_at"`
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
	Name        string    `gorm:"size:200;not null" json:"name"`        // Nazwa źródła w OBS
	SourceType  string    `gorm:"size:100;not null;default:'UNKNOWN'" json:"source_type"` // Typ źródła z OBS
	SourceOrder int       `gorm:"default:0" json:"source_order"`        // Domyślna kolejność
	IsVisible   bool      `gorm:"default:false" json:"is_visible"`      // Stan użytkownika (dla mikrofonów)
	IconURL     *string   `gorm:"size:500" json:"icon_url"`             // Ikona dla przycisku (nullable)
	Color       string    `gorm:"size:20" json:"color"`                 // Kolor przycisku (hex)
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
	ID         uint      `gorm:"primaryKey" json:"id"`
	EpisodeID  uint      `gorm:"index;not null" json:"episode_id"`
	Episode    Episode   `gorm:"foreignKey:EpisodeID" json:"episode"`
	SourceName string    `gorm:"size:200;not null" json:"source_name"` // "Media1", "Media2", "Kamera1", "Mikrofon1"
	
	// Przypisania (tylko jedno wypełnione w zależności od typu źródła)
	MediaID    *uint     `gorm:"index" json:"media_id"`   // Dla Media1/Reportaze1 → ID pliku
	GroupID    *uint     `gorm:"index" json:"group_id"`   // Dla Media2/Reportaze2 → ID grupy
	StaffID    *uint     `gorm:"index" json:"staff_id"`   // Dla mikrofonów → ID prowadzącego
	GuestID    *uint     `gorm:"index" json:"guest_id"`   // Dla mikrofonów → ID gościa
	PresetID   *uint     `gorm:"index" json:"preset_id"`  // Dla kamer → ID presetu (przyszłość)
	
	// Metadane
	AssignedBy string    `gorm:"size:50;default:'manual'" json:"assigned_by"` // "auto" lub "manual"
	
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

	// Ustaw domyślny typ dla istniejących źródeł (migracja)
	db.Exec("UPDATE sources SET source_type = 'UNKNOWN' WHERE source_type = '' OR source_type IS NULL")

	// Dodaj unikalny indeks na parę (episode_id, source_name) dla episode_sources
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_episode_source ON episode_sources(episode_id, source_name)")

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
		}
	}

	return assignments, nil
}
