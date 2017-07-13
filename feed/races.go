package feed

import (
	"net/http"
	"strconv"

	"github.com/chiefwhitecloud/running-man/database"
	"github.com/gorilla/mux"
)

// ListRaces Return a list of race tags. Check for etag
func (r *FeedResource) ListRaces(w http.ResponseWriter, req *http.Request) {

	var etag string

	raceLastUpdated, err := r.Db.GetLastUpdatedRace()

	if err != nil && err != database.ErrNoRecordsAvailable {
		handleError(err, w)
		return
	}

	if err != database.ErrNoRecordsAvailable {
		etag = raceLastUpdated.ETag

		sent, error := SendNotModifiedIfETagIsValid(w, req, etag)

		if error != nil {
			handleError(error, w)
			return
		}

		if sent {
			return
		}
	}

	races, err := r.Db.GetRaces()

	if err != nil {
		handleError(err, w)
		return
	}

	if len(etag) > 0 {
		SendJsonWithETag(w, FormatRacesForFeed(req, races), etag)
	} else {
		SendJson(w, FormatRacesForFeed(req, races))
	}

}

// GetRace Returns individual race info
func (r *FeedResource) GetRace(w http.ResponseWriter, req *http.Request) {

	race := r.GetRaceOrSendError(w, req)

	if race == nil {
		return
	}

	ok, err := SendNotModifiedIfETagIsValid(w, req, race.ETag)

	if err != nil {
		handleError(err, w)
		return
	}

	if !ok {
		SendJsonWithETag(w, FormatRaceForFeed(req, *race), race.ETag)
	}

	return
}

//DeleteRace Delete race
func (r *FeedResource) DeleteRace(w http.ResponseWriter, req *http.Request) {

	race := r.GetRaceOrSendError(w, req)

	if race == nil {
		return
	}

	if _, err := r.Db.DeleteRace(race.ID); err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)

}

//GetRaceResultsForRace Handler for fetching race results for a race
func (r *FeedResource) GetRaceResultsForRace(res http.ResponseWriter, req *http.Request) {

	race := r.GetRaceOrSendError(res, req)

	if race == nil {
		return
	}

	var startPlace int
	var recCount int
	var err error

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

	if ok, _ := SendNotModifiedIfETagIsValid(res, req, race.ETag); ok {
		return
	}

	rr, racers, races, err := r.Db.GetRaceResultsForRace(race.ID, startPlace, recCount)

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJsonWithETag(res, FormatRaceResultsForFeed(req, rr, racers, races), race.ETag)

}

func (r *FeedResource) GetRaceOrSendError(w http.ResponseWriter, req *http.Request) *database.Race {
	vars := mux.Vars(req)

	raceID, err := strconv.Atoi(vars["id"])

	if err != nil {
		handleError(ErrNotFound, w)
		return nil
	}

	race, err := r.Db.GetRace(raceID)
	if err != nil {
		handleError(err, w)
		return nil
	}

	return &race

}
