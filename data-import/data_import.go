package dataimport

import (
	"encoding/json"
	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/chiefwhitecloud/running-man/feed"
	"log"
	"net/http"
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
		http.Error(res, err.Error(), 500)
	}

	results, err := r.RaceFetcher.GetRawResults(dataimport.RaceUrl)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	//parse the race results from the html string
	raceDetails, err := parseResults(results)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	race, err := r.Db.SaveRace(&raceDetails)

	raceFeed := feed.FormatRaceForFeed(req, race)

	raceFeedFormatted, err := json.Marshal(&raceFeed)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(raceFeedFormatted))

}
