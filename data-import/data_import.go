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

	v := req.Header.Get("Content-Type")
	if v != "application/json" {
		http.Error(res, "Invalid Request", 400)
		return
	}

	var dataimport api.DataImport

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&dataimport)

	if err != nil {
		http.Error(res, "Invalid Request", 400)
		return
	}

	importTask, _ := r.Db.CreateImportTask(dataimport.RaceUrl)
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Location", feed.FormatImportTaskLocation(req, importTask.ID))
	res.WriteHeader(http.StatusAccepted)

	r.ImportResults(importTask)
}

func (r *DataImportResource) ImportResults(task database.ImportTask) {

	results, err := r.RaceFetcher.GetRawResults(task.SrcUrl)

	if err != nil {
		r.Db.FailedImport(task, err)
		return
	}

	//parse the race results from the html string
	raceDetails, err := parseResults(results)

	if err != nil {
		r.Db.FailedImport(task, err)
		return
	}

	_, err = r.Db.SaveRace(task, &raceDetails)

	if err != nil {
		r.Db.FailedImport(task, err)
		return
	}

}

func (r *DataImportResource) CheckImportStatus(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	taskId, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(res, err.Error(), 404)
	}

	task, _ := r.Db.GetImportTask(taskId)

	if task.Status == "pending" {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
	} else if task.Status == "failed" {
		http.Error(res, task.ErrorText, http.StatusInternalServerError)
	} else if task.Status == "completed" {
		//redirect to the new race resource
		http.Redirect(res, req, feed.FormatRaceLocation(req, task.RaceID), http.StatusSeeOther)
	}
}
