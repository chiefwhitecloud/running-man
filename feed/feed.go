package feed

import (
	"errors"
	"log"
	"net/http"

	"github.com/chiefwhitecloud/running-man/database"
)

var _ = log.Print

var ErrNotFound = errors.New("Record not found")

var ErrBadRequest = errors.New("Bad Request")

type FeedResource struct {
	Db database.Db
}

func handleError(err error, w http.ResponseWriter) {

	if err == database.ErrRecordNotFoundError {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if err == ErrBadRequest {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
