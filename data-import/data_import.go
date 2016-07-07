package dataimport

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/chiefwhitecloud/running-man/feed"
	"github.com/gorilla/mux"
)

var _ = log.Print

type DataImportResource struct {
	Db          database.Db
	RaceFetcher RaceFetcher
}

func (r *DataImportResource) DoImport(res http.ResponseWriter, req *http.Request) {

	var dataimport api.DataImport

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&dataimport)

	if err != nil {
		http.Error(res, "Invalid parameter", 400)
		return
	}

	pendingRaceId, _ := r.Db.CreatePendingRace()
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Location", feed.FormatImportTaskLocation(req, pendingRaceId))
	res.WriteHeader(http.StatusAccepted)

	r.ImportResults(pendingRaceId, dataimport.RaceUrl)
}

func (r *DataImportResource) ImportResults(pendingRaceId int, url string) {

	results, err := r.RaceFetcher.GetRawResults(url)

	if err != nil {
		return
	}

	//parse the race results from the html string
	raceDetails, err := parseResults(results)

	if err != nil {
		log.Println(err.Error())
		//http.Error(res, err.Error(), 500)
		return
	}

	_, err = r.Db.SaveRace(pendingRaceId, &raceDetails)

	//raceFeed := feed.FormatRaceForFeed(req, race)

	//raceFeedFormatted, err := json.Marshal(&raceFeed)

	//if err != nil {
	//	http.Error(res, err.Error(), 500)
	//}

}

func (r *DataImportResource) CheckImportStatus(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	raceId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	race, _ := r.Db.GetRace(raceId)

	if race.ImportStatus == "pending" {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
	} else if race.ImportStatus == "completed" {
		//redirect to the new race resource
		http.Redirect(res, req, feed.FormatRaceLocation(req, race.ID), http.StatusSeeOther)
	}
}
