package feed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chiefwhitecloud/running-man/api"
	"github.com/gorilla/mux"
)

//GetRacer Fetch Racer
func (r *FeedResource) GetRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	racer, err := r.Db.GetRacer(racerID)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJson(res, FormatRacerForFeed(req, racer))
}

//GetRacerProfile Fetch racer profile
func (r *FeedResource) GetRacerProfile(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	racer, err := r.Db.GetRacer(racerID)

	lowBirthDate, highBirthDate, _ := r.Db.GetRacerBirthDates(racerID)
	names, _ := r.Db.GetRacerNames(racerID)

	racerProfile := api.RacerProfile{
		Name:          names[0],
		NameList:      names,
		SelfPath:      fmt.Sprintf("http://%s/feed/racer/%d/profile", req.Host, racer.ID),
		BirthDateLow:  fmt.Sprintf("%0.4d-%0.2d-%0.2d", lowBirthDate.Year(), lowBirthDate.Month(), lowBirthDate.Day()),
		BirthDateHigh: fmt.Sprintf("%0.4d-%0.2d-%0.2d", highBirthDate.Year(), highBirthDate.Month(), highBirthDate.Day()),
	}

	SendJson(res, racerProfile)

}

//GetRaceResultsForRacer Fetch results for the racer
func (r *FeedResource) GetRaceResultsForRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	racerID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	rr, racers, races, err := r.Db.GetRaceResultsForRacer(uint(racerID))

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJson(res, FormatRaceResultsForFeed(req, rr, racers, races))
}

//MergeRacer Merge to racers results
func (r *FeedResource) MergeRacer(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	parentRacerID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	parentRacer, err := r.Db.GetRacer(parentRacerID)

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	var racerMerge api.RacerMerge

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&racerMerge)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	racerID, err := strconv.Atoi(racerMerge.RacerId)

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	racer, err := r.Db.GetRacer(racerID)

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
