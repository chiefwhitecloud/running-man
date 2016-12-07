package feed

import (
	"encoding/json"
	"net/http"
)

func SendNotModifiedIfETagIsValid(res http.ResponseWriter, req *http.Request, etag string) (bool, error) {
	if req.Header.Get("If-None-Match") != "" && req.Header.Get("If-None-Match") == etag {
		res.WriteHeader(http.StatusNotModified)
		return true, nil
	} else {
		return false, nil
	}
}

func SendJsonWithETag(res http.ResponseWriter, entity interface{}, etag string) error {
	res.Header().Set("ETag", etag)
	SendJson(res, entity)
	return nil
}

func SendJson(res http.ResponseWriter, entity interface{}) error {
	if entity != nil {
		b, err := json.Marshal(entity)
		if err != nil {
			return err
		}
		jsonResponse(res)
		res.WriteHeader(http.StatusOK)
		res.Write(b)
	} else {
		res.WriteHeader(http.StatusOK)
	}
	return nil
}

func SendSuccess(res http.ResponseWriter) error {
	b, err := json.Marshal(map[string]interface{}{"success": 1})
	if err != nil {
		return err
	}
	jsonResponse(res)
	res.WriteHeader(http.StatusOK)
	res.Write(b)
	return nil
}

func jsonResponse(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
}
