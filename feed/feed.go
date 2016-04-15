package feed

import (
	"encoding/json"
	"fmt"
	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var _ = log.Print

type FeedResource struct {
	Db database.Db
}

func (r *FeedResource) ListRaces(res http.ResponseWriter, req *http.Request) {

	races, err := r.Db.GetRaces()

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	raceList := make([]api.Race, len(races))
	for i, _ := range races {
		raceList[i] = FormatRaceForFeed(req, races[i])
	}

	feed := api.RaceFeed{Races: raceList}

	raceDetailsFormatted, err := json.Marshal(&feed)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(raceDetailsFormatted))

}

func (r *FeedResource) GetRace(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	raceId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	race, err := r.Db.GetRace(raceId)

	raceFeed := FormatRaceForFeed(req, race)

	raceFeedFormatted, err := json.Marshal(&raceFeed)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(raceFeedFormatted))

}

func (r *FeedResource) GetRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	racer, err := r.Db.GetRacer(racerId)

	racerFeed := r.formatRacerForFeed(req, racer)

	racerFeedFormatted, err := json.Marshal(&racerFeed)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(racerFeedFormatted))

}

func (r *FeedResource) GetRaceResultsForRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	rr, racers, races, err := r.Db.GetRaceResultsForRacer(uint(racerId))

	raceFeed := r.formatRaceResultsForFeed(req, rr, racers, races)

	raceFeedFormatted, err := json.Marshal(&raceFeed)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(raceFeedFormatted))

}

func (r *FeedResource) GetRaceResultsForRace(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	raceId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	rr, racers, races, err := r.Db.GetRaceResultsForRace(uint(raceId))

	raceFeed := r.formatRaceResultsForFeed(req, rr, racers, races)

	raceFeedFormatted, err := json.Marshal(&raceFeed)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(raceFeedFormatted))

}

func FormatRaceForFeed(req *http.Request, race database.Race) api.Race {
	return api.Race{
		Name:        race.Name,
		SelfPath:    fmt.Sprintf("http://%s/race/%d", req.Host, race.ID),
		ResultsPath: fmt.Sprintf("http://%s/race/%d/results", req.Host, race.ID),
		Date:        fmt.Sprintf("%0.4d-%0.2d-%0.2d", race.Year, race.Month, race.Day),
	}
}

func (r *FeedResource) formatRacerForFeed(req *http.Request, racer database.Racer) api.Racer {
	return api.Racer{
		FirstName:   racer.FirstName,
		LastName:    racer.LastName,
		Sex:         racer.Sex,
		SelfPath:    fmt.Sprintf("http://%s/racer/%d", req.Host, racer.ID),
		ResultsPath: fmt.Sprintf("http://%s/racer/%d/results", req.Host, racer.ID),
	}
}

func (r *FeedResource) formatRaceResultsForFeed(req *http.Request, raceresults []database.RaceResult, racers []database.Racer, races []database.Race) api.RaceResults {

	ageMap := map[int]string{
		1:  "U20",
		2:  "20-29",
		3:  "30-39",
		4:  "40-49",
		5:  "50-59",
		6:  "60-69",
		7:  "70-79",
		8:  "80-89",
		9:  "90-99",
		10: "100-109",
	}

	mapRacers := map[string]api.Racer{}
	for i := range racers {
		mapRacers[strconv.Itoa(racers[i].ID)] = r.formatRacerForFeed(req, racers[i])
	}

	mapRaces := map[string]api.Race{}
	for i := range races {
		mapRaces[strconv.Itoa(races[i].ID)] = FormatRaceForFeed(req, races[i])
	}

	rr := make([]api.RaceResult, len(raceresults))
	for i := range raceresults {
		rr[i] = api.RaceResult{
			Position:            raceresults[i].Position,
			SexPosition:         raceresults[i].SexPosition,
			AgeCategoryPosition: raceresults[i].AgeCategoryPosition,
			RacerID:             strconv.Itoa(raceresults[i].RacerID),
			RaceID:              strconv.Itoa(raceresults[i].RaceID),
			BibNumber:           raceresults[i].BibNumber,
			Time:                raceresults[i].Time,
			AgeCategory:         ageMap[raceresults[i].AgeCategoryID],
		}
	}

	return api.RaceResults{Results: rr, Racers: mapRacers, Races: mapRaces}
}
