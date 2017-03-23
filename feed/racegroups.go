package feed

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/gorilla/mux"
)

// CreateRaceGroup Create a new race group
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

	raceGroupDB, _ := r.Db.CreateRaceGroup(raceGroup.Name, raceGroup.Distance, raceGroup.DistanceUnit)

	raceGroupFeed := FormatRaceGroupForFeed(req, raceGroupDB)

	raceGroupFeedFormatted, _ := json.Marshal(&raceGroupFeed)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(raceGroupFeedFormatted))

}

// UpdateRaceGroup Update the race group
func (r *FeedResource) UpdateRaceGroup(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	raceGroupID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 400)
	}

	raceGroupDB, err := r.Db.GetRaceGroup(int(raceGroupID))

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	var raceGroup api.RaceGroupCreate

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&raceGroup)

	if err != nil {
		http.Error(res, "Bad Parameters", http.StatusBadRequest)
		return
	}

	var raceGroupUpdated database.RaceGroup

	if raceGroupUpdated, err = r.Db.UpdateRaceGroup(raceGroupDB.ID, raceGroup.Name, raceGroup.Distance, raceGroup.DistanceUnit); err != nil {
		http.Error(res, err.Error(), 500)
	}

	SendJson(res, FormatRaceGroupForFeed(req, raceGroupUpdated))

}

// DeleteRaceGroup Delete the race group
func (r *FeedResource) DeleteRaceGroup(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	raceGroupID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	if _, err := r.Db.DeleteRaceGroup(int(raceGroupID)); err != nil {
		http.Error(res, err.Error(), 500)
	}

	res.WriteHeader(http.StatusOK)

}

// ListRaceGroups Get a list of the all the race groups
func (r *FeedResource) ListRaceGroups(res http.ResponseWriter, req *http.Request) {

	var etag string

	if raceGroupLastUpdated, err := r.Db.GetLastUpdatedRaceGroup(); err == nil {
		if ok, _ := SendNotModifiedIfETagIsValid(res, req, raceGroupLastUpdated.ETag); ok {
			return
		}
		etag = raceGroupLastUpdated.ETag
	}

	raceGroups, err := r.Db.GetRaceGroups()

	if err != nil {
		http.Error(res, err.Error(), 500)
	}

	if len(etag) > 0 {
		SendJsonWithETag(res, FormatRaceGroupsForFeed(req, raceGroups), etag)
	} else {
		SendJson(res, FormatRaceGroupsForFeed(req, raceGroups))
	}
}

// GetRaceGroup Fetch the requested race group and return its data
func (r *FeedResource) GetRaceGroup(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	raceGroupID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	raceGroupDB, _ := r.Db.GetRaceGroup(int(raceGroupID))

	SendJson(res, FormatRaceGroupForFeed(req, raceGroupDB))

}
