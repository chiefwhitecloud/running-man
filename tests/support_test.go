package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"
)

var _ = fmt.Print
var _ = log.Print

func Test(t *testing.T) { TestingT(t) }

type RaceFetcherStub struct {
}

func (c *RaceFetcherStub) GetRawResults(resultsurl string) ([]byte, error) {

	u, err := url.Parse(resultsurl)
	if err != nil {
		log.Fatal(err)
	}

	if u.Path == "/00-Road-Race.html" {
		absPath, _ := filepath.Abs("test-data/00-Road-Race.html")
		byes, _ := ioutil.ReadFile(absPath)
		return byes, nil
	} else if u.Path == "/01-Road-Race.html" {
		absPath, _ := filepath.Abs("test-data/01-Road-Race.html")
		byes, _ := ioutil.ReadFile(absPath)
		return byes, nil
	} else if u.Path == "/02-Tely.html" {
		absPath, _ := filepath.Abs("test-data/02-Tely.html")
		byes, _ := ioutil.ReadFile(absPath)
		return byes, nil
	} else if u.Path == "/03-Road-Race.html" {
		absPath, _ := filepath.Abs("test-data/03-Road-Race.html")
		byes, _ := ioutil.ReadFile(absPath)
		return byes, nil
	} else if u.Path == "/04-Road-Race.html" {
		absPath, _ := filepath.Abs("test-data/04-Road-Race.html")
		byes, _ := ioutil.ReadFile(absPath)
		return byes, nil
	} else if u.Path == "/05-Tely.html" {
		absPath, _ := filepath.Abs("test-data/05-Tely.html")
		byes, _ := ioutil.ReadFile(absPath)
		return byes, nil
	} else {
		return []byte(`{"raceUrl": "Hello"}`), nil
	}

}
