package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/chiefwhitecloud/running-man/service"
	"github.com/parnurzeal/gorequest"
	. "gopkg.in/check.v1"
)

var _ = fmt.Print
var _ = log.Print

type TestSuite struct {
	s    service.RunningManService
	host string
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpSuite(c *C) {

	server := service.RunningManService{
		Db:          database.Db{ConnectionString: os.Getenv("DATABASE_URL")},
		Bind:        os.Getenv("PORT"),
		RaceFetcher: &RaceFetcherStub{},
	}

	s.s = server

	s.s.Db.Open()

	go s.s.Run()

	s.host = "http://localhost:" + os.Getenv("PORT")
}

func (s *TestSuite) SetUpTest(c *C) {
	s.s.MigrateDb()
}

func (s *TestSuite) TearDownTest(c *C) {
	s.s.DropAllTables()
}

// Simple import
func (s *TestSuite) Test01Import(c *C) {

	//import a race
	request := gorequest.New()
	resp, body, _ := request.Post(fmt.Sprintf("%s/import", s.host)).
		Send(`{"raceUrl":"http://www.nlaa.ca/00-Road-Race.html"}`).
		End()
	c.Assert(resp.StatusCode, Equals, 201)
	var race api.Race
	jsonBlob := []byte(body)
	err := json.Unmarshal(jsonBlob, &race)
	c.Assert(race.Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(race.Id, Equals, 1)

	// fetch the race list
	resp, body, _ = request.Get(fmt.Sprintf("%s/feed/races", s.host)).End()
	c.Assert(resp.StatusCode, Equals, 200)

	jsonBlob = []byte(body)
	var races api.RaceFeed
	err = json.Unmarshal(jsonBlob, &races)

	c.Assert(len(races.Races), Equals, 1)
	c.Assert(races.Races[0].Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(races.Races[0].SelfPath, Equals, s.host+"/feed/race/1")
	c.Assert(races.Races[0].ResultsPath, Equals, s.host+"/feed/race/1/results")
	c.Assert(races.Races[0].Date, Equals, "2015-04-12")

	raceSelfPath := races.Races[0].SelfPath
	raceResultsPath := races.Races[0].ResultsPath

	resp, body, _ = request.Get(raceSelfPath).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	err = json.Unmarshal(jsonBlob, &race)

	c.Assert(err, Equals, nil)
	c.Assert(race.Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(race.SelfPath, Equals, s.host+"/feed/race/1")
	c.Assert(race.ResultsPath, Equals, s.host+"/feed/race/1/results")
	c.Assert(race.Date, Equals, "2015-04-12")

	//fetch race results
	resp, body, _ = request.Get(raceResultsPath).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	var raceResults api.RaceResults
	err = json.Unmarshal(jsonBlob, &raceResults)
	c.Assert(err, Equals, nil)
	c.Assert(len(raceResults.Results), Equals, 10)

	//results should be order.
	c.Assert(raceResults.Results[0].Position, Equals, 1)
	c.Assert(raceResults.Results[0].AgeCategoryPosition, Equals, 1)
	c.Assert(raceResults.Results[0].Time, Equals, "15:45")
	c.Assert(raceResults.Results[0].AgeCategory, Equals, "20-29")
	c.Assert(raceResults.Results[1].Position, Equals, 2)
	c.Assert(raceResults.Results[2].Position, Equals, 3)
	c.Assert(raceResults.Results[3].Position, Equals, 4)
	c.Assert(raceResults.Results[4].Position, Equals, 5)
	c.Assert(raceResults.Results[4].AgeCategoryPosition, Equals, 4)
	c.Assert(raceResults.Results[5].Position, Equals, 6)
	c.Assert(raceResults.Results[6].Position, Equals, 7)
	c.Assert(raceResults.Results[7].Position, Equals, 8)
	c.Assert(raceResults.Results[8].Position, Equals, 9)
	//check the race and racers map
	c.Assert(raceResults.Results[0].RacerID, Equals, "1")
	c.Assert(raceResults.Results[0].RaceID, Equals, "1")
	c.Assert(raceResults.Results[0].Name, Equals, "JORDAN FEWER")
	c.Assert(len(raceResults.Racers), Equals, 10)
	c.Assert(len(raceResults.Races), Equals, 1)
	c.Assert(raceResults.Races[raceResults.Results[0].RaceID].Name, Equals, "Boston Pizza Flat Out 5 km Road Race")

	andreaSparkesId := raceResults.Racers[raceResults.Results[9].RacerID].Id

	//fetch the racer
	jordanSelfPath := raceResults.Racers[raceResults.Results[0].RacerID].SelfPath
	var jordanRacer api.Racer
	s.doRequest(jordanSelfPath, &jordanRacer)
	c.Assert(jordanRacer.ProfilePath, Equals, s.host+"/feed/racer/1/profile")

	var jordanRacerProfile api.RacerProfile
	var jordanProfilePath = jordanRacer.ProfilePath
	s.doRequest(jordanProfilePath, &jordanRacerProfile)
	c.Assert(jordanRacerProfile.Name, Equals, "JORDAN FEWER")
	c.Assert(jordanRacerProfile.BirthDateLow, Equals, "1985-04-13")
	c.Assert(jordanRacerProfile.BirthDateHigh, Equals, "1995-04-12")

	//fetch the racer results
	jordanResults := jordanRacer.ResultsPath
	s.doRequest(jordanResults, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 1)

	//import another race
	resp, body, _ = request.Post(fmt.Sprintf("%s/import", s.host)).
		Send(`{"raceUrl":"http://www.nlaa.ca/01-Road-Race.html"}`).
		End()
	c.Assert(resp.StatusCode, Equals, 201)
	jsonBlob = []byte(body)
	err = json.Unmarshal(jsonBlob, &race)
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")
	c.Assert(race.ResultsPath, Equals, s.host+"/feed/race/2/results")
	c.Assert(race.Date, Equals, "2015-04-26")
	raceResultsPath = race.ResultsPath

	//fetch race results
	s.doRequest(raceResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 17)

	slowChrisResultsPath := raceResults.Racers[raceResults.Results[10].RacerID].ResultsPath
	andreaWhiteResultsPath := raceResults.Racers[raceResults.Results[11].RacerID].ResultsPath

	//check joe's bday
	c.Assert(raceResults.Results[9].Position, Equals, 10)
	joeDunfordSelfPath := raceResults.Racers[raceResults.Results[9].RacerID].SelfPath
	resp, body, _ = request.Get(joeDunfordSelfPath).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	var joeDunfordRacer api.Racer
	err = json.Unmarshal(jsonBlob, &joeDunfordRacer)
	c.Assert(err, Equals, nil)

	var joeDunfordProfile api.RacerProfile
	s.doRequest(joeDunfordRacer.ProfilePath, &joeDunfordProfile)
	c.Assert(joeDunfordProfile.Name, Equals, "JOE DUNFORD")
	c.Assert(joeDunfordProfile.BirthDateLow, Equals, "1965-04-13")
	c.Assert(joeDunfordProfile.BirthDateHigh, Equals, "1965-04-26")

	//fetch the racer results
	s.doRequest(joeDunfordRacer.ResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 2)

	//jordan should have two results
	s.doRequest(jordanResults, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 2)

	//his birthdate range should be updated.
	s.doRequest(jordanProfilePath, &jordanRacerProfile)
	c.Assert(jordanRacerProfile.BirthDateLow, Equals, "1985-04-27")
	c.Assert(jordanRacerProfile.BirthDateHigh, Equals, "1995-04-12")

	//slow chris should only have one result
	s.doRequest(slowChrisResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 1)

	//andrea should only have one race result
	s.doRequest(andreaWhiteResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 1)
	c.Assert(raceResults.Racers[raceResults.Results[0].RacerID].MergePath, Not(Equals), "")

	//merge andrea white and andrea sparkes
	andreaWhiteMergePath := raceResults.Racers[raceResults.Results[0].RacerID].MergePath
	var merge = api.RacerMerge{RacerId: strconv.Itoa(andreaSparkesId)}
	resp, _, _ = request.Post(andreaWhiteMergePath).
		Send(merge).
		End()

	s.doRequest(andreaWhiteResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 2)

}

// Simple import
func (s *TestSuite) Test02ImportTely(c *C) {

	//import a race

	request := gorequest.New()
	resp, body, _ := request.Post(fmt.Sprintf("%s/import", s.host)).
		Send(`{"raceUrl":"http://www.nlaa.ca/02-Tely.html"}`).
		End()
	c.Assert(resp.StatusCode, Equals, 201)

	var race api.Race
	jsonBlob := []byte(body)
	_ = json.Unmarshal(jsonBlob, &race)
	c.Assert(race.Name, Equals, "88th Annual Tely 10 Mile Road Race")
	c.Assert(race.Id, Equals, 1)

	// fetch the race list
	resp, body, _ = request.Get(fmt.Sprintf("%s/feed/races", s.host)).End()
	c.Assert(resp.StatusCode, Equals, 200)

	jsonBlob = []byte(body)
	var races api.RaceFeed
	_ = json.Unmarshal(jsonBlob, &races)

	c.Assert(len(races.Races), Equals, 1)
	c.Assert(races.Races[0].Name, Equals, "88th Annual Tely 10 Mile Road Race")
	c.Assert(races.Races[0].SelfPath, Equals, s.host+"/feed/race/1")
	c.Assert(races.Races[0].ResultsPath, Equals, s.host+"/feed/race/1/results")
	c.Assert(races.Races[0].Date, Equals, "2015-07-26")

	var raceResults api.RaceResults
	s.doRequest(races.Races[0].ResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 38)
	c.Assert(len(raceResults.Racers), Equals, 38)
	c.Assert(raceResults.Results[0].AgeCategoryPosition, Equals, 1)

}

// Import a race from 2008
func (s *TestSuite) Test03ImportRoadRace(c *C) {

	//import a race
	request := gorequest.New()
	resp, body, _ := request.Post(fmt.Sprintf("%s/import", s.host)).
		Send(`{"raceUrl":"http://www.nlaa.ca/03-Road-Race.html"}`).
		End()
	c.Assert(resp.StatusCode, Equals, 201)

	var race api.Race
	jsonBlob := []byte(body)
	_ = json.Unmarshal(jsonBlob, &race)
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")
	c.Assert(race.Id, Equals, 1)
	c.Assert(race.ResultsPath, Equals, s.host+"/feed/race/1/results")

	var raceResults api.RaceResults
	s.doRequest(race.ResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 18)
	c.Assert(len(raceResults.Racers), Equals, 18)

}

func (s *TestSuite) doRequest(path string, entity interface{}) error {

	request := gorequest.New()

	resp, body, _ := request.Get(path).End()

	if resp.StatusCode != 200 {
		return errors.New("Bad request")
	}

	jsonBlob := []byte(body)
	if errs := json.Unmarshal(jsonBlob, entity); errs != nil {
		return errs
	}

	return nil
}
