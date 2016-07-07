package feed

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/gorilla/mux"
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

func (r *FeedResource) GetRacerProfile(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	racer, err := r.Db.GetRacer(racerId)

	lowBirthDate, highBirthDate, _ := r.Db.GetRacerBirthDates(racerId)
	names, _ := r.Db.GetRacerNames(racerId)

	racerProfile := api.RacerProfile{
		Name:          names[0],
		NameList:      names,
		SelfPath:      fmt.Sprintf("http://%s/feed/racer/%d/profile", req.Host, racer.ID),
		BirthDateLow:  fmt.Sprintf("%0.4d-%0.2d-%0.2d", lowBirthDate.Year(), lowBirthDate.Month(), lowBirthDate.Day()),
		BirthDateHigh: fmt.Sprintf("%0.4d-%0.2d-%0.2d", highBirthDate.Year(), highBirthDate.Month(), highBirthDate.Day()),
	}

	racerFeedFormatted, err := json.Marshal(&racerProfile)

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

func (r *FeedResource) MergeRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	parentRacerId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	parentRacer, err := r.Db.GetRacer(parentRacerId)

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	var racerMerge api.RacerMerge

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&racerMerge)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	racerId, err := strconv.Atoi(racerMerge.RacerId)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	racer, err := r.Db.GetRacer(racerId)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	_, err = r.Db.MergeRacers(parentRacer, racer)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	//res.Header().Set("Content-Type", "application/json")
	//res.WriteHeader(http.StatusOK)
	//res.Write([]byte(raceFeedFormatted))

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

func FormatImportTaskLocation(req *http.Request, taskId int) string {
	return fmt.Sprintf("http://%s/import/task/%d", req.Host, taskId)
}

func FormatRaceLocation(req *http.Request, raceId int) string {
	return fmt.Sprintf("http://%s/feed/race/%d", req.Host, raceId)
}

func FormatRaceForFeed(req *http.Request, race database.Race) api.Race {
	return api.Race{
		Id:          race.ID,
		Name:        race.Name,
		SelfPath:    fmt.Sprintf("http://%s/feed/race/%d", req.Host, race.ID),
		ResultsPath: fmt.Sprintf("http://%s/feed/race/%d/results", req.Host, race.ID),
		Date:        fmt.Sprintf("%0.4d-%0.2d-%0.2d", race.Date.Year(), race.Date.Month(), race.Date.Day()),
	}
}

func (r *FeedResource) formatRacerForFeed(req *http.Request, racer database.Racer) api.Racer {
	return api.Racer{
		Id:          racer.ID,
		SelfPath:    fmt.Sprintf("http://%s/feed/racer/%d", req.Host, racer.ID),
		ResultsPath: fmt.Sprintf("http://%s/feed/racer/%d/results", req.Host, racer.ID),
		ProfilePath: fmt.Sprintf("http://%s/feed/racer/%d/profile", req.Host, racer.ID),
		MergePath:   fmt.Sprintf("http://%s/feed/racer/%d/merge", req.Host, racer.ID),
	}
}

func (r *FeedResource) formatRaceResultsForFeed(req *http.Request, raceresults []database.RaceResult, racers []database.Racer, races []database.Race) api.RaceResults {

	ageMap := map[int]string{
		1:  "U20",
		2:  "-19",
		3:  "20-24",
		4:  "25-29",
		5:  "20-29",
		6:  "30-34",
		7:  "35-39",
		8:  "30-39",
		9:  "40-44",
		10: "45-49",
		11: "40-49",
		12: "50-54",
		13: "55-59",
		14: "50-59",
		15: "60-64",
		16: "65-69",
		17: "60-69",
		18: "70-74",
		19: "75-79",
		20: "70-79",
		21: "70+",
		22: "80-84",
		23: "85-89",
		24: "80-89",
		25: "80+",
		26: "A",
	}

	//	"80-84", "85-89", "80-89",
	//	"90-94", "95-99", "90-99",
	//		"100-104", "105-109", "100-109",
	//	}

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
			Name:                raceresults[i].Name,
			Position:            raceresults[i].Position,
			SexPosition:         raceresults[i].SexPosition,
			Sex:                 raceresults[i].Sex,
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
