package main

import (
	"log"
	"net/http"
	"obs-controller/handlers"
	"obs-controller/models"
	"obs-controller/obsws"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Uruchamianie aplikacji...")

	db, err := gorm.Open(sqlite.Open("obs_controller.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Błąd bazy danych: %v", err)
	}
	log.Println("Baza danych OK")

	if err := models.InitDB(db); err != nil {
		log.Fatalf("Błąd inicjalizacji: %v", err)
	}
	log.Println("Tabele OK")

	// Próbuj połączyć z OBS-WebSocket w pętli
	var obsClient *obsws.Client

	log.Println("Próba połączenia z OBS-WebSocket...")
	for {
		obsClient, err = obsws.NewClient("ws://localhost:4445")
		if err != nil {
			log.Printf("Nie można połączyć z OBS-WebSocket: %v", err)
			log.Println("Ponowna próba za 5 sekund...")
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("OBS-WebSocket OK")
		break
	}
	defer obsClient.Close()

	socketHandler, err := handlers.NewSocketHandler(db, obsClient)
	if err != nil {
		log.Fatalf("Błąd Socket.IO: %v", err)
	}
	log.Println("Socket.IO OK")

	// Inicjalizacja handlerów
	seasonHandler := handlers.NewSeasonHandler(db)
	episodeHandler := handlers.NewEpisodeHandler(db)
	staffTypeHandler := handlers.NewStaffTypeHandler(db)
	staffHandler := handlers.NewStaffHandler(db)
	episodeStaffHandler := handlers.NewEpisodeStaffHandler(db)
	guestTypeHandler := handlers.NewGuestTypeHandler(db)
	guestHandler := handlers.NewGuestHandler(db)
	episodeGuestHandler := handlers.NewEpisodeGuestHandler(db)
	sceneHandler := handlers.NewSceneHandler(db)
	mediaGroupHandler := handlers.NewMediaGroupHandler(db)
	settingsHandler := handlers.NewSettingsHandler(db)

	// Middleware do sprawdzania wymagań i inicjalizacji
	initMiddleware := handlers.NewInitMiddleware(db, obsClient)

	// Ścieżka do mediów
	mediaPath := "./media"
	os.MkdirAll(mediaPath, 0755)
	episodeMediaHandler := handlers.NewEpisodeMediaHandler(db, mediaPath)

	// Routing
	router := mux.NewRouter()

	// Socket.IO
	router.Handle("/socket.io/", socketHandler.Server)

	// Pliki statyczne z poprawnymi MIME types
	fileServer := http.FileServer(http.Dir("./web/static"))
	router.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}
		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	})

	// Strony HTML
	router.HandleFunc("/controller", initMiddleware.CheckRequirements(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/controller.html")
	}))

	router.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/settings.html")
	})

	router.HandleFunc("/seasons", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/seasons.html")
	})

	router.HandleFunc("/episodes", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/episodes.html")
	})

	router.HandleFunc("/staff", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/staff.html")
	})

	router.HandleFunc("/guests", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/guests.html")
	})

	router.HandleFunc("/overlay", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/overlay.html")
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/settings", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})

	// API REST dla Season
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/seasons", seasonHandler.GetSeasons).Methods("GET")
	api.HandleFunc("/seasons/next-number", seasonHandler.GetNextSeasonNumber).Methods("GET")
	api.HandleFunc("/seasons", seasonHandler.CreateSeason).Methods("POST")
	api.HandleFunc("/seasons/{id}", seasonHandler.GetSeason).Methods("GET")
	api.HandleFunc("/seasons/{id}", seasonHandler.UpdateSeason).Methods("PUT")
	api.HandleFunc("/seasons/{id}", seasonHandler.DeleteSeason).Methods("DELETE")
	api.HandleFunc("/seasons/{id}/set-current", seasonHandler.SetCurrentSeason).Methods("POST")

	// API REST dla Episode
	api.HandleFunc("/episodes", episodeHandler.GetEpisodes).Methods("GET")
	api.HandleFunc("/episodes/next-numbers", episodeHandler.GetNextEpisodeNumbers).Methods("GET")
	api.HandleFunc("/episodes", episodeHandler.CreateEpisode).Methods("POST")
	api.HandleFunc("/episodes/{id}", episodeHandler.GetEpisode).Methods("GET")
	api.HandleFunc("/episodes/{id}", episodeHandler.UpdateEpisode).Methods("PUT")
	api.HandleFunc("/episodes/{id}", episodeHandler.DeleteEpisode).Methods("DELETE")
	api.HandleFunc("/episodes/{id}/set-current", episodeHandler.SetCurrentEpisode).Methods("POST")

	// API REST dla Staff
	api.HandleFunc("/staff-types", staffTypeHandler.GetStaffTypes).Methods("GET")
	api.HandleFunc("/staff-types", staffTypeHandler.CreateStaffType).Methods("POST")
	api.HandleFunc("/staff-types/{id}", staffTypeHandler.GetStaffType).Methods("GET")
	api.HandleFunc("/staff-types/{id}", staffTypeHandler.UpdateStaffType).Methods("PUT")
	api.HandleFunc("/staff-types/{id}", staffTypeHandler.DeleteStaffType).Methods("DELETE")

	api.HandleFunc("/staff", staffHandler.GetStaff).Methods("GET")
	api.HandleFunc("/staff", staffHandler.CreateStaff).Methods("POST")
	api.HandleFunc("/staff/{id}", staffHandler.GetStaffMember).Methods("GET")
	api.HandleFunc("/staff/{id}", staffHandler.UpdateStaff).Methods("PUT")
	api.HandleFunc("/staff/{id}", staffHandler.DeleteStaff).Methods("DELETE")

	api.HandleFunc("/episodes/{episode_id}/staff", episodeStaffHandler.GetEpisodeStaff).Methods("GET")
	api.HandleFunc("/episodes/{episode_id}/staff", episodeStaffHandler.AddStaffToEpisode).Methods("POST")
	api.HandleFunc("/episodes/{episode_id}/staff/{id}", episodeStaffHandler.RemoveStaffFromEpisode).Methods("DELETE")
	api.HandleFunc("/episodes/{episode_id}/staff/{id}/types", episodeStaffHandler.UpdateEpisodeStaffTypes).Methods("PUT")

	// API REST dla Guests
	api.HandleFunc("/guest-types", guestTypeHandler.GetGuestTypes).Methods("GET")
	api.HandleFunc("/guest-types", guestTypeHandler.CreateGuestType).Methods("POST")
	api.HandleFunc("/guest-types/{id}", guestTypeHandler.GetGuestType).Methods("GET")
	api.HandleFunc("/guest-types/{id}", guestTypeHandler.UpdateGuestType).Methods("PUT")
	api.HandleFunc("/guest-types/{id}", guestTypeHandler.DeleteGuestType).Methods("DELETE")

	api.HandleFunc("/guests", guestHandler.GetGuests).Methods("GET")
	api.HandleFunc("/guests", guestHandler.CreateGuest).Methods("POST")
	api.HandleFunc("/guests/{id}", guestHandler.GetGuest).Methods("GET")
	api.HandleFunc("/guests/{id}", guestHandler.UpdateGuest).Methods("PUT")
	api.HandleFunc("/guests/{id}", guestHandler.DeleteGuest).Methods("DELETE")

	api.HandleFunc("/episodes/{episode_id}/guests", episodeGuestHandler.GetEpisodeGuests).Methods("GET")
	api.HandleFunc("/episodes/{episode_id}/guests", episodeGuestHandler.AddGuestToEpisode).Methods("POST")
	api.HandleFunc("/episodes/{episode_id}/guests/{id}", episodeGuestHandler.UpdateEpisodeGuest).Methods("PUT")
	api.HandleFunc("/episodes/{episode_id}/guests/{id}", episodeGuestHandler.RemoveGuestFromEpisode).Methods("DELETE")

	// API REST dla EpisodeMedia
	api.HandleFunc("/episodes/{episode_id}/media", episodeMediaHandler.GetEpisodeMedia).Methods("GET")
	api.HandleFunc("/episodes/{episode_id}/media", episodeMediaHandler.CreateEpisodeMedia).Methods("POST")
	api.HandleFunc("/episodes/{episode_id}/media/{id}", episodeMediaHandler.UpdateEpisodeMedia).Methods("PUT")
	api.HandleFunc("/episodes/{episode_id}/media/{id}", episodeMediaHandler.DeleteEpisodeMedia).Methods("DELETE")
	api.HandleFunc("/episodes/{episode_id}/media/{id}/set-current", episodeMediaHandler.SetCurrentMedia).Methods("POST")
	api.HandleFunc("/episodes/{episode_id}/media/{id}/reorder", episodeMediaHandler.ReorderEpisodeMedia).Methods("PUT")
	api.HandleFunc("/episodes/{episode_id}/media/upload", episodeMediaHandler.UploadMedia).Methods("POST")
	api.HandleFunc("/episodes/{episode_id}/media/files", episodeMediaHandler.ListMediaFiles).Methods("GET")

	// API REST dla Scenes
	api.HandleFunc("/scenes", sceneHandler.GetScenes).Methods("GET")
	api.HandleFunc("/scenes/media", sceneHandler.GetMediaScenes).Methods("GET")

	// API REST dla Settings
	api.HandleFunc("/settings/status", settingsHandler.GetStatus).Methods("GET")

	// API REST dla MediaGroup
	api.HandleFunc("/media-groups", mediaGroupHandler.GetMediaGroups).Methods("GET")
	api.HandleFunc("/media-groups", mediaGroupHandler.CreateMediaGroup).Methods("POST")
	api.HandleFunc("/media-groups/{id}", mediaGroupHandler.GetMediaGroup).Methods("GET")
	api.HandleFunc("/media-groups/{id}", mediaGroupHandler.UpdateMediaGroup).Methods("PUT")
	api.HandleFunc("/media-groups/{id}", mediaGroupHandler.DeleteMediaGroup).Methods("DELETE")
	api.HandleFunc("/media-groups/{id}/items", mediaGroupHandler.GetMediaGroupItems).Methods("GET")
	api.HandleFunc("/media-groups/{id}/items", mediaGroupHandler.AddItemToGroup).Methods("POST")
	api.HandleFunc("/media-groups/{id}/reorder", mediaGroupHandler.ReorderMediaGroup).Methods("PUT")
	api.HandleFunc("/media-groups/{group_id}/items/{id}/reorder", mediaGroupHandler.ReorderMediaGroupItem).Methods("PUT")
	api.HandleFunc("/media-groups/{group_id}/media/{media_id}", mediaGroupHandler.AddMediaToGroup).Methods("POST")
	api.HandleFunc("/media-groups/{group_id}/media/{media_id}", mediaGroupHandler.RemoveMediaFromGroup).Methods("DELETE")
	api.HandleFunc("/episodes/{episode_id}/media-groups/{group_id}/set-current", mediaGroupHandler.SetCurrentMediaGroup).Methods("POST")

	log.Println("========================================")
	log.Println("Serwer działa: http://localhost:8080")
	log.Println("Ustawienia: http://localhost:8080/settings")
	log.Println("Kontroler: http://localhost:8080/controller")
	log.Println("Sezony: http://localhost:8080/seasons")
	log.Println("Odcinki: http://localhost:8080/episodes")
	log.Println("Ekipa: http://localhost:8080/staff")
	log.Println("Goście: http://localhost:8080/guests")
	log.Println("Overlay: http://localhost:8080/overlay")
	log.Println("========================================")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Błąd serwera: %v", err)
	}
}
