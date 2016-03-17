package test

import (
	"encoding/json"
	"fmt"
	"github.com/chiefwhitecloud/running-man/Godeps/_workspace/src/github.com/parnurzeal/gorequest"
	. "github.com/chiefwhitecloud/running-man/Godeps/_workspace/src/gopkg.in/check.v1"
	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/chiefwhitecloud/running-man/service"
	"log"
	"os"
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

	// fetch the race list
	resp, body, _ = request.Get(fmt.Sprintf("%s/races", s.host)).End()
	c.Assert(resp.StatusCode, Equals, 200)

	jsonBlob = []byte(body)
	var races api.RaceFeed
	err = json.Unmarshal(jsonBlob, &races)

	c.Assert(len(races.Races), Equals, 1)
	c.Assert(races.Races[0].Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(races.Races[0].SelfPath, Equals, "/race/1")
	c.Assert(races.Races[0].ResultsPath, Equals, "/race/1/results")
	c.Assert(races.Races[0].Date, Equals, "2015-04-12")

	raceSelfPath := races.Races[0].SelfPath
	raceResultsPath := races.Races[0].ResultsPath

	resp, body, _ = request.Get(s.host + raceSelfPath).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	err = json.Unmarshal(jsonBlob, &race)

	c.Assert(err, Equals, nil)
	c.Assert(race.Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(race.SelfPath, Equals, "/race/1")
	c.Assert(race.ResultsPath, Equals, "/race/1/results")
	c.Assert(race.Date, Equals, "2015-04-12")

	//fetch race results
	resp, body, _ = request.Get(s.host + raceResultsPath).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	var raceResults api.RaceResults
	err = json.Unmarshal(jsonBlob, &raceResults)
	c.Assert(err, Equals, nil)
	c.Assert(len(raceResults.Results), Equals, 9)

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
	c.Assert(raceResults.Racers["1"].FirstName, Equals, "JORDAN")
	c.Assert(raceResults.Racers["1"].LastName, Equals, "FEWER")
	c.Assert(len(raceResults.Racers), Equals, 9)
	c.Assert(len(raceResults.Races), Equals, 1)
	c.Assert(raceResults.Races["1"].Name, Equals, "Boston Pizza Flat Out 5 km Road Race")

	//fetch the racer
	jordanSelfPath := raceResults.Racers["1"].SelfPath
	resp, body, _ = request.Get(s.host + jordanSelfPath).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	var jordanRacer api.Racer
	err = json.Unmarshal(jsonBlob, &jordanRacer)
	c.Assert(err, Equals, nil)
	c.Assert(jordanRacer.FirstName, Equals, "JORDAN")
	c.Assert(jordanRacer.LastName, Equals, "FEWER")
	c.Assert(jordanRacer.Sex, Equals, "M")

	//fetch the racer results
	jordanResults := jordanRacer.ResultsPath
	resp, body, _ = request.Get(s.host + jordanResults).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	err = json.Unmarshal(jsonBlob, &raceResults)
	c.Assert(err, Equals, nil)
	c.Assert(len(raceResults.Results), Equals, 1)

	//import another race
	resp, body, _ = request.Post(fmt.Sprintf("%s/import", s.host)).
		Send(`{"raceUrl":"http://www.nlaa.ca/01-Road-Race.html"}`).
		End()
	c.Assert(resp.StatusCode, Equals, 201)
	jsonBlob = []byte(body)
	err = json.Unmarshal(jsonBlob, &race)
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")

	//jordan should have two results
	resp, body, _ = request.Get(s.host + jordanResults).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	err = json.Unmarshal(jsonBlob, &raceResults)
	c.Assert(err, Equals, nil)
	c.Assert(len(raceResults.Results), Equals, 2)

}
