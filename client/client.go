package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/model"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var _ = log.Print

type Client struct {
	Host string
}

func (c *Client) AddRace(raceUrl string) (model.RaceDetails, error) {

	var raceDetails model.RaceDetails

	addRace := api.DataImport{
		RaceUrl: raceUrl,
	}

	url := fmt.Sprintf("%s/import", c.Host)
	r, err := MakeRequest("POST", url, addRace)
	if err != nil {
		return raceDetails, err
	}
	err = ProcessResponseEntity(r, &raceDetails, http.StatusCreated)
	return raceDetails, err
}

func (c *Client) GetRaces() (api.RaceFeed, error) {

	var races api.RaceFeed

	url := fmt.Sprintf("%s/races", c.Host)
	r, err := MakeRequest("GET", url, nil)
	if err != nil {
		return api.RaceFeed{}, err
	}
	err = ProcessResponseEntity(r, &races, http.StatusOK)
	return races, err
}

func (c *Client) GetRace() (api.Race, error) {

	var race api.Race

	url := fmt.Sprintf("%s/race/1", c.Host)
	r, err := MakeRequest("GET", url, nil)
	if err != nil {
		return api.Race{}, err
	}
	err = ProcessResponseEntity(r, &race, http.StatusOK)
	return race, err
}

func (c *Client) GetRaceResults() (api.RaceResults, error) {

	var rr api.RaceResults

	url := fmt.Sprintf("%s/race/1/results", c.Host)
	r, err := MakeRequest("GET", url, nil)
	if err != nil {
		return api.RaceResults{}, err
	}
	err = ProcessResponseEntity(r, &rr, http.StatusOK)
	return rr, err
}

func (c *Client) GetRacerResults() (api.RaceResults, error) {

	var rr api.RaceResults

	url := fmt.Sprintf("%s/racer/1/results", c.Host)
	r, err := MakeRequest("GET", url, nil)
	if err != nil {
		return api.RaceResults{}, err
	}
	err = ProcessResponseEntity(r, &rr, http.StatusOK)
	return rr, err
}

var (
	ErrStatusConflict            error = errors.New(fmt.Sprintf("%d: Conflict", http.StatusConflict))
	ErrStatusBadRequest          error = errors.New(fmt.Sprintf("%d: Bad Request", http.StatusBadRequest))
	ErrStatusInternalServerError error = errors.New(fmt.Sprintf("%d: Internal Server Error", http.StatusInternalServerError))
	ErrStatusNotFound            error = errors.New(fmt.Sprintf("%d: Not Found", http.StatusNotFound))
)

func NewRequestH(method string, url string, headers map[string]interface{}, entity interface{}) (*http.Response, error) {
	req, err := buildRequest(method, url, headers, entity)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func MakeRequest(method string, url string, entity interface{}) (*http.Response, error) {
	headers := map[string]interface{}{}
	req, err := buildRequest(method, url, headers, entity)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func buildRequest(method string, url string, headers map[string]interface{}, entity interface{}) (*http.Request, error) {
	body, err := encodeEntity(entity)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}

	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	for header, value := range headers {
		switch value.(type) {
		case int:
			req.Header.Set(header, string(value.(int)))
		case string:
			req.Header.Set(header, value.(string))
		}
	}
	return req, err
}

func encodeEntity(entity interface{}) (io.Reader, error) {
	if entity == nil {
		return nil, nil
	} else {
		b, err := json.Marshal(entity)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(b), nil
	}
}

func ProcessResponseBytes(r *http.Response, expectedStatus int) ([]byte, error) {
	if err := processResponse(r, expectedStatus); err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(r.Body)
	return respBody, err
}
func ProcessResponseEntity(r *http.Response, entity interface{}, expectedStatus int) error {
	if err := processResponse(r, expectedStatus); err != nil {
		return err
	}
	return ForceProcessResponseEntity(r, entity)
}
func ForceProcessResponseEntity(r *http.Response, entity interface{}) error {
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if entity != nil {
		if err = json.Unmarshal(respBody, entity); err != nil {
			return err
		}
	}
	return nil
}
func processResponse(r *http.Response, expectedStatus int) error {
	if r == nil {
		return errors.New("response is nil")
	}
	if r.StatusCode != expectedStatus {

		switch r.StatusCode {
		case http.StatusConflict:
			return ErrStatusConflict
		case http.StatusBadRequest:
			return ErrStatusBadRequest
		case http.StatusInternalServerError:
			return ErrStatusInternalServerError
		case http.StatusNotFound:
			return ErrStatusNotFound
		default:
			return errors.New("response status of " + r.Status)
		}

	}

	return nil
}
