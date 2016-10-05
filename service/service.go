package service

import (
	"log"
	"net/http"
	"os"

	_ "github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/data-import"
	"github.com/chiefwhitecloud/running-man/data-import/fetcher"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/chiefwhitecloud/running-man/feed"
	"github.com/chiefwhitecloud/running-man/ui"
	"github.com/gorilla/mux"
)

var _ = log.Printf

type RunningManService struct {
	Bind        string
	Db          database.Db
	RaceFetcher dataimport.RaceFetcher
}

func NewRunningManService(bind string, dbStr string) (*RunningManService, error) {

	s := RunningManService{
		Db:          database.Db{ConnectionString: dbStr},
		Bind:        bind,
		RaceFetcher: &fetcher.RaceFetcher{},
	}

	s.Db.Open()

	return &s, nil
}

func (s *RunningManService) MigrateDb() error {
	s.Db.Migrate()
	return nil
}

func (s *RunningManService) Create() error {
	s.Db.Create()
	return nil
}

func (s *RunningManService) DropAllTables() error {
	s.Db.DropAllTables()
	return nil
}

func (s *RunningManService) Run() error {

	importer := &dataimport.DataImportResource{
		Db:          s.Db,
		RaceFetcher: s.RaceFetcher,
	}

	feeds := &feed.FeedResource{
		Db: s.Db,
	}

	cwd, _ := os.Getwd()

	ui := &ui.UI{
		BaseDir: cwd,
	}

	// route handlers
	r := mux.NewRouter()

	var importRouter = r.PathPrefix("/feed/").Subrouter()
	importRouter.HandleFunc("/", importer.DoImport).Methods("POST")
	importRouter.HandleFunc("/task/{id}", importer.CheckImportStatus).Methods("GET")
	var feedRouter = r.PathPrefix("/feed/").Subrouter()
	feedRouter.HandleFunc("/races", feeds.ListRaces).Methods("GET")
	feedRouter.HandleFunc("/race/{id}", feeds.GetRace).Methods("GET")
	feedRouter.HandleFunc("/race/{id}/results", feeds.GetRaceResultsForRace).Methods("GET")
	feedRouter.HandleFunc("/racer/{id}", feeds.GetRacer).Methods("GET")
	feedRouter.HandleFunc("/racer/{id}/results", feeds.GetRaceResultsForRacer).Methods("GET")
	feedRouter.HandleFunc("/racer/{id}/profile", feeds.GetRacerProfile).Methods("GET")
	feedRouter.HandleFunc("/racer/{id}/merge", feeds.MergeRacer).Methods("POST")

	r.PathPrefix("/").Handler(ui)

	http.Handle("/", r)

	// Start HTTP Server
	return http.ListenAndServe(":"+s.Bind, nil)
}
