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
	Reportages    []EpisodeMedia `gorm:"foreignKey:EpisodeID;constraint:OnDelete:CASCADE" json:"reportages"`
	Media         []EpisodeMedia `gorm:"foreignKey:EpisodeID;constraint:OnDelete:CASCADE" json:"media"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// StaffType reprezentuje typ członka ekipy
type StaffType struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:100;uniqueIndex;not null" json:"name"` // np. "Redaktor prowadzący", "Realizator dźwięku"
	// Staff     []Staff   `gorm:"foreignKey:StaffTypeID" json:"staff"`
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

// MediaGroup reprezentuje grupę mediów (np. "Blok reportaży", "Playlista muzyczna")
type MediaGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EpisodeID   uint      `gorm:"index;not null" json:"episode_id"`    // Przypisanie do odcinka
	Episode     Episode   `gorm:"foreignKey:EpisodeID" json:"episode"` // Relacja do odcinka
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Order       int       `gorm:"not null" json:"order"`                 // Kolejność w odcinku (unique per episode)
	IsCurrent   bool      `gorm:"default:false;index" json:"is_current"` // Czy aktywna w odcinku
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EpisodeMediaGroup reprezentuje przypisanie media do grupy w kontekście odcinka
// Pozwala na tworzenie playlist i grup reportaży
type EpisodeMediaGroup struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	EpisodeMediaID uint         `gorm:"index;not null" json:"episode_media_id"`
	EpisodeMedia   EpisodeMedia `gorm:"foreignKey:EpisodeMediaID" json:"episode_media"`
	MediaGroupID   uint         `gorm:"index;not null" json:"media_group_id"`
	MediaGroup     MediaGroup   `gorm:"foreignKey:MediaGroupID" json:"media_group"`
	Order          int          `gorm:"not null" json:"order"`           // Kolejność w grupie (unique per group)
	IsCurrent      bool         `gorm:"default:false" json:"is_current"` // Czy media jest aktywne w źródle List
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
	Name        string    `gorm:"size:200;not null" json:"name"`   // Nazwa źródła w OBS
	DisplayName *string   `gorm:"size:200" json:"display_name"`    // Tekst na przycisku kontrolera (nullable)
	SourceOrder int       `gorm:"default:0" json:"source_order"`   // Domyślna kolejność
	IsVisible   bool      `gorm:"default:false" json:"is_visible"` // Stan użytkownika (dla mikrofonów)
	IconURL     *string   `gorm:"size:500" json:"icon_url"`        // Ikona dla przycisku (nullable)
	Color       string    `gorm:"size:20" json:"color"`            // Kolor przycisku (hex)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EpisodeMedia reprezentuje media (reportaże, filmy) przypisane do odcinka
// Scena określa typ media: MEDIA lub REPORTAZE
type EpisodeMedia struct {
	ID             uint                `gorm:"primaryKey" json:"id"`
	EpisodeID      uint                `gorm:"index;not null" json:"episode_id"`
	Episode        Episode             `gorm:"foreignKey:EpisodeID" json:"episode"`
	SceneID        uint                `gorm:"index;not null" json:"scene_id"` // MEDIA lub REPORTAZE
	Scene          Scene               `gorm:"foreignKey:SceneID" json:"scene"`
	EpisodeStaffID *uint               `gorm:"index" json:"episode_staff_id"`                  // Autor z przypisanej ekipy (nullable)
	EpisodeStaff   *EpisodeStaff       `gorm:"foreignKey:EpisodeStaffID" json:"episode_staff"` // Autor z przypisanej ekipy (nullable)
	Title          string              `gorm:"size:300;not null" json:"title"`
	Description    string              `gorm:"type:text" json:"description"`
	FilePath       *string             `gorm:"size:1000" json:"file_path"`                    // Ścieżka do pliku (nullable)
	URL            *string             `gorm:"size:1000" json:"url"`                          // URL jeśli zewnętrzne (nullable)
	Duration       int                 `json:"duration"`                                      // Czas trwania w sekundach
	Order          int                 `gorm:"default:0" json:"order"`                        // Kolejność w odcinku
	IsCurrent      bool                `gorm:"default:false" json:"is_current"`               // Czy wczytany w źródło Single
	MediaGroups    []EpisodeMediaGroup `gorm:"foreignKey:EpisodeMediaID" json:"media_groups"` // Przynależność do grup
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// InitDB inicjalizuje bazę danych
func InitDB(db *gorm.DB) error {
	return db.AutoMigrate(
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
		&EpisodeMedia{},
		&EpisodeMediaGroup{},
	)
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

// SetCurrentEpisodeMedia ustawia media jako current (wczytane w źródło Single)
func SetCurrentEpisodeMedia(db *gorm.DB, episodeID uint, mediaID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Pobierz media aby znać scenę
		var media EpisodeMedia
		if err := tx.First(&media, mediaID).Error; err != nil {
			return err
		}

		// Wyłącz wszystkie media tej samej sceny w tym odcinku
		if err := tx.Model(&EpisodeMedia{}).
			Where("episode_id = ? AND scene_id = ? AND is_current = ?", episodeID, media.SceneID, true).
			Update("is_current", false).Error; err != nil {
			return err
		}

		// Włącz wybrane media
		if err := tx.Model(&media).Update("is_current", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// SetCurrentMediaGroup ustawia grupę jako current (wczytana w źródło List)
func SetCurrentMediaGroup(db *gorm.DB, episodeID uint, groupID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Wyłącz wszystkie grupy w tym odcinku
		if err := tx.Model(&MediaGroup{}).
			Where("episode_id = ? AND is_current = ?", episodeID, true).
			Update("is_current", false).Error; err != nil {
			return err
		}

		// Włącz wybraną grupę
		if err := tx.Model(&MediaGroup{}).
			Where("id = ?", groupID).
			Update("is_current", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetNextEpisodeMediaOrder zwraca następny dostępny numer kolejności dla media w odcinku
func GetNextEpisodeMediaOrder(db *gorm.DB, episodeID uint) int {
	var maxMedia EpisodeMedia
	result := db.Where("episode_id = ?", episodeID).Order("\"order\" DESC").First(&maxMedia)
	if result.Error != nil {
		return 0
	}
	return maxMedia.Order + 1
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
