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

	raceLastUpdated, err := r.Db.GetLastUpdatedRace()

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	if ok, _ := IsETagValid(res, req, raceLastUpdated.ETag); ok {
		return
	}

	races, err := r.Db.GetRaces()

	if err != nil {
		http.Error(res, err.Error(), 500)
	}
	SendJsonWithETag(res, FormatRacesForFeed(req, races), raceLastUpdated.ETag)
}

func (r *FeedResource) GetRace(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	raceId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	race, err := r.Db.GetRace(raceId)

	if ok, err := IsETagValid(res, req, race.ETag); err == nil && !ok {
		SendJsonWithETag(res, FormatRaceForFeed(req, race), race.ETag)
	}
}

func (r *FeedResource) GetRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	racer, err := r.Db.GetRacer(racerId)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJson(res, FormatRacerForFeed(req, racer))
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

	SendJson(res, racerProfile)

}

func (r *FeedResource) GetRaceResultsForRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	rr, racers, races, err := r.Db.GetRaceResultsForRacer(uint(racerId))

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJson(res, FormatRaceResultsForFeed(req, rr, racers, races))
}

func (r *FeedResource) GetRacesForRaceGroup(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	raceGroupId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	races, err := r.Db.GetRacesForRaceGroup(int(raceGroupId))

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJson(res, FormatRacesForFeed(req, races))

}

func (r *FeedResource) AddRaceToRaceGroup(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	var addRaceGroup api.RaceGroupAddRace

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&addRaceGroup)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	raceId, err := strconv.Atoi(addRaceGroup.RaceId)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	race, err := r.Db.GetRace(raceId)

	raceGroupId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	raceGroup, err := r.Db.GetRaceGroup(raceGroupId)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	r.Db.AddRaceToRaceGroup(raceGroup, race)

	res.WriteHeader(http.StatusOK)
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

func (r *FeedResource) CreateRaceGroup(res http.ResponseWriter, req *http.Request) {
	v := req.Header.Get("Content-Type")
	if v != "application/json" {
		http.Error(res, "Invalid Request", http.StatusBadRequest)
		return
	}

	var raceGroup api.RaceGroupCreate

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&raceGroup)

	if err != nil {
		http.Error(res, "Bad Parameters", http.StatusBadRequest)
		return
	}

	raceGroupDB, _ := r.Db.CreateRaceGroup(raceGroup.Name, raceGroup.Distance)

	raceGroupFeed := FormatRaceGroupForFeed(req, raceGroupDB)

	raceGroupFeedFormatted, _ := json.Marshal(&raceGroupFeed)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(raceGroupFeedFormatted))

}

func (r *FeedResource) ListRaceGroups(res http.ResponseWriter, req *http.Request) {

	raceGroupLastUpdated, err := r.Db.GetLastUpdatedRaceGroup()

	if req.Header.Get("If-None-Match") == raceGroupLastUpdated.ETag {
		res.WriteHeader(http.StatusNotModified)
		return
	}

	raceGroups, err := r.Db.GetRaceGroups()

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	raceGroupList := make([]api.RaceGroup, len(raceGroups))
	for i, _ := range raceGroups {
		raceGroupList[i] = FormatRaceGroupForFeed(req, raceGroups[i])
	}

	feed := api.RaceGroupFeed{RaceGroups: raceGroupList}

	raceGroupFormatted, err := json.Marshal(&feed)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("ETag", raceGroupLastUpdated.ETag)
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(raceGroupFormatted))

}

func (r *FeedResource) GetRaceGroup(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	raceGroupId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	raceGroupDB, _ := r.Db.GetRaceGroup(int(raceGroupId))

	raceGroupFeed := FormatRaceGroupForFeed(req, raceGroupDB)

	raceGroupFeedFormatted, _ := json.Marshal(&raceGroupFeed)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(raceGroupFeedFormatted))
}

func (r *FeedResource) GetRaceResultsForRace(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	raceId, err := strconv.Atoi(vars["id"])

	startPlace := 0

	recCount := 0

	//optional querystring parameters
	place := req.URL.Query().Get("startPos")

	if len(place) != 0 {
		startPlace, err = strconv.Atoi(place)
		if err != nil {
			http.Error(res, err.Error(), 400)
		}
	}

	numOfRecords := req.URL.Query().Get("num")

	if len(numOfRecords) != 0 {
		recCount, err = strconv.Atoi(numOfRecords)
		if err != nil {
			http.Error(res, err.Error(), 400)
		}
	}

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	race, err := r.Db.GetRace(raceId)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	if ok, _ := IsETagValid(res, req, race.ETag); ok {
		return
	}

	rr, racers, races, err := r.Db.GetRaceResultsForRace(raceId, startPlace, recCount)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJsonWithETag(res, FormatRaceResultsForFeed(req, rr, racers, races), race.ETag)

}
