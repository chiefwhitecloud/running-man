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

	r.HandleFunc("/import", importer.DoImport).Methods("POST")
	r.HandleFunc("/import/task/{id}", importer.CheckImportStatus).Methods("GET")

	var feedRouter = r.PathPrefix("/feed/").Subrouter()
	feedRouter.HandleFunc("/racegroup", feeds.CreateRaceGroup).Methods("POST")
	feedRouter.HandleFunc("/racegroups", feeds.ListRaceGroups).Methods("GET")
	feedRouter.HandleFunc("/racegroup/{id}", feeds.DeleteRaceGroup).Methods("DELETE")
	feedRouter.HandleFunc("/racegroup/{id}", feeds.UpdateRaceGroup).Methods("PUT")
	feedRouter.HandleFunc("/racegroup/{id}", feeds.GetRaceGroup).Methods("GET")
	feedRouter.HandleFunc("/racegroup/{id}/races", feeds.AddRaceToRaceGroup).Methods("POST")
	feedRouter.HandleFunc("/racegroup/{id}/races", feeds.GetRacesForRaceGroup).Methods("GET")
	feedRouter.HandleFunc("/races", feeds.ListRaces).Methods("GET")
	feedRouter.HandleFunc("/race/{id}", feeds.GetRace).Methods("GET")
	feedRouter.HandleFunc("/race/{id}", feeds.DeleteRace).Methods("DELETE")
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
