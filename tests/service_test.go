package test

import (
	"encoding/json"
	"fmt"
	"github.com/chiefwhitecloud/running-man/client"
	"github.com/chiefwhitecloud/running-man/database"
	"github.com/chiefwhitecloud/running-man/service"
	. "gopkg.in/check.v1"
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

	// Load Config
	type Configuration struct {
		Database string
		Bind     string
	}

	file, _ := os.Open("../test.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	server := service.RunningManService{
		Db:          database.Db{ConnectionString: configuration.Database},
		Bind:        configuration.Bind,
		RaceFetcher: &RaceFetcherStub{},
	}

	s.s = server

	s.s.Db.Open()

	go s.s.Run()

	s.host = "http://" + configuration.Bind
}

func (s *TestSuite) SetUpTest(c *C) {
	s.s.MigrateDb()
}

func (s *TestSuite) TearDownTest(c *C) {
	s.s.DropAllTables()
}

// Simple import
func (s *TestSuite) Test01Import(c *C) {

	// use the client api
	client := client.Client{Host: s.host}

	race, err := client.AddRace("http://www.nlaa.ca/00-Road-Race.html")
	c.Assert(race.Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(len(race.Racers), Equals, 9)
	c.Assert(err, Equals, nil)

	racelist, err := client.GetRaces()
	c.Assert(err, Equals, nil)
	c.Assert(len(racelist.Races), Equals, 1)
	c.Assert(racelist.Races[0].Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(racelist.Races[0].SelfPath, Equals, "/race/1")
	c.Assert(racelist.Races[0].ResultsPath, Equals, "/race/1/results")
	c.Assert(racelist.Races[0].Date, Equals, "2015-04-12")

	raceDetails, err := client.GetRace()
	c.Assert(err, Equals, nil)
	c.Assert(raceDetails.Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(raceDetails.SelfPath, Equals, "/race/1")
	c.Assert(raceDetails.ResultsPath, Equals, "/race/1/results")
	c.Assert(raceDetails.Date, Equals, "2015-04-12")

	raceResults, err := client.GetRaceResults()
	c.Assert(err, Equals, nil)
	c.Assert(len(raceResults.Results), Equals, 9)

	racerResults, err := client.GetRacerResults()
	c.Assert(err, Equals, nil)
	c.Assert(len(racerResults.Results), Equals, 1)

	race2, err := client.AddRace("http://www.nlaa.ca/01-Road-Race.html")
	c.Assert(race2.Name, Equals, "Nautilus Mundy Pond 5km Road Race")
	c.Assert(len(race2.Racers), Equals, 9)
	c.Assert(err, Equals, nil)

	racerResults, err = client.GetRacerResults()
	c.Assert(err, Equals, nil)
	c.Assert(len(racerResults.Results), Equals, 2)

}
