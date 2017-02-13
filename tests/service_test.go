package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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
	race, _ := s.doImport("http://www.nlaa.ca/00-Road-Race.html")
	c.Assert(race.Name, Equals, "Boston Pizza Flat Out 5 km Road Race")
	c.Assert(race.Id, Equals, 1)

	// fetch the race list
	request := gorequest.New()
	resp, body, _ := request.Get(fmt.Sprintf("%s/feed/races", s.host)).End()
	c.Assert(resp.StatusCode, Equals, 200)

	jsonBlob := []byte(body)
	var races api.RaceFeed
	json.Unmarshal(jsonBlob, &races)

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
	err := json.Unmarshal(jsonBlob, &race)

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
	c.Assert(raceResults.Results[0].SexPosition, Equals, 1)
	c.Assert(raceResults.Results[0].AgeCategoryPosition, Equals, 1)
	c.Assert(raceResults.Results[0].Time, Equals, "15:45")
	c.Assert(raceResults.Results[0].AgeCategory, Equals, "20-29")
	c.Assert(raceResults.Results[1].Position, Equals, 2)
	c.Assert(raceResults.Results[1].Club, Equals, "PGNL")
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
	race, _ = s.doImport("http://www.nlaa.ca/01-Road-Race.html")
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

	race, _ := s.doImport("http://www.nlaa.ca/02-Tely.html")
	c.Assert(race.Name, Equals, "88th Annual Tely 10 Mile Road Race")
	c.Assert(race.Id, Equals, 1)

	// fetch the race list
	request := gorequest.New()
	resp, body, _ := request.Get(fmt.Sprintf("%s/feed/races", s.host)).End()
	c.Assert(resp.StatusCode, Equals, 200)

	jsonBlob := []byte(body)
	var races api.RaceFeed
	_ = json.Unmarshal(jsonBlob, &races)

	c.Assert(len(races.Races), Equals, 1)
	c.Assert(races.Races[0].Name, Equals, "88th Annual Tely 10 Mile Road Race")
	c.Assert(races.Races[0].SelfPath, Equals, s.host+"/feed/race/1")
	c.Assert(races.Races[0].ResultsPath, Equals, s.host+"/feed/race/1/results")
	c.Assert(races.Races[0].Date, Equals, "2015-07-26")

	var raceResults api.RaceResults
	s.doRequest(races.Races[0].ResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 40)
	c.Assert(len(raceResults.Racers), Equals, 40)
	c.Assert(raceResults.Results[0].AgeCategoryPosition, Equals, 1)
	c.Assert(raceResults.Results[0].SexPosition, Equals, 1)
	c.Assert(raceResults.Results[0].BibNumber, Equals, "3662")
	c.Assert(raceResults.Results[0].ChipTime, Equals, "49:25")
	c.Assert(raceResults.Results[len(raceResults.Results)-1].ChipTime, Equals, " ")

}

// Import a race from 2008
func (s *TestSuite) Test03ImportRoadRace(c *C) {

	//import a race
	race, _ := s.doImport("http://www.nlaa.ca/03-Road-Race.html")
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")
	c.Assert(race.Id, Equals, 1)
	c.Assert(race.ResultsPath, Equals, s.host+"/feed/race/1/results")

	var raceResults api.RaceResults
	s.doRequest(race.ResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 18)
	c.Assert(len(raceResults.Racers), Equals, 18)
	var result = raceResults.Results[0]
	c.Assert(result.Name, Equals, "PATRICK O'GRADY")
	c.Assert(result.AgeCategory, Equals, "U20")
	c.Assert(result.Time, Equals, "16:51")
}

// Import a race from 2008
func (s *TestSuite) Test04ImportFailed(c *C) {

	//import a race
	_, err := s.doImport("http://www.nlaa.ca/")
	c.Assert(err, Not(Equals), nil)
}

func (s *TestSuite) Test05ImportDuplicate(c *C) {

	//import a race
	race, _ := s.doImport("http://www.nlaa.ca/03-Road-Race.html")
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")

	//cant import the same race twice
	_, err := s.doImport("http://www.nlaa.ca/03-Road-Race.html")
	c.Assert(err, Not(Equals), nil)
}

func (s *TestSuite) Test06Import(c *C) {

	//import a race
	race, err := s.doImport("http://www.nlaa.ca/04-Road-Race.html")
	c.Assert(race.Name, Equals, "ANE Mile")
	c.Assert(err, Equals, nil)

	var raceResults api.RaceResults
	s.doRequest(race.ResultsPath, &raceResults)
	c.Assert(len(raceResults.Results), Equals, 61)
	c.Assert(len(raceResults.Racers), Equals, 61)
	c.Assert(raceResults.Results[0].SexPosition, Equals, 1)
}

func (s *TestSuite) Test07Import(c *C) {

	//import a race
	race, err := s.doImport("http://www.nlaa.ca/05-Tely.html")
	c.Assert(err, Equals, nil)
	c.Assert(race.Name, Equals, "83rd Annual Tely 10 Mile Road Race")

	//import another race
	var raceResults api.RaceResults
	raceResultsPath := race.ResultsPath
	s.doRequest(raceResultsPath+"?num=10", &raceResults)
	c.Assert(len(raceResults.Results), Equals, 10)
	c.Assert(raceResults.Results[0].Position, Equals, 1)

	s.doRequest(raceResultsPath+"?startPos=2&num=5", &raceResults)
	c.Assert(len(raceResults.Results), Equals, 5)
	c.Assert(raceResults.Results[0].Position, Equals, 2)

}

func (s *TestSuite) Test08ETag(c *C) {

	//import a race
	race, err := s.doImport("http://www.nlaa.ca/05-Tely.html")
	c.Assert(err, Equals, nil)
	c.Assert(race.Name, Equals, "83rd Annual Tely 10 Mile Road Race")

	//check etag on race self path
	request := gorequest.New()
	resp, _, _ := request.Get(race.SelfPath).End()
	raceEtag := resp.Header.Get("ETag")
	c.Assert(raceEtag, Not(Equals), "")
	resp, _, _ = request.Get(race.SelfPath).Set("If-None-Match", raceEtag).End()
	c.Assert(resp.StatusCode, Equals, 304)

	//check etag on raceresults
	request = gorequest.New()
	resp, _, _ = request.Get(race.ResultsPath).End()
	raceEtag = resp.Header.Get("ETag")
	c.Assert(raceEtag, Not(Equals), "")
	resp, _, _ = request.Get(race.ResultsPath).Set("If-None-Match", raceEtag).End()
	c.Assert(resp.StatusCode, Equals, 304)

	//check etag on race list
	request = gorequest.New()
	resp, _, _ = request.Get(fmt.Sprintf("%s/feed/races", s.host)).End()
	raceListEtag := resp.Header.Get("ETag")
	c.Assert(raceEtag, Not(Equals), "")
	resp, _, _ = request.Get(race.ResultsPath).Set("If-None-Match", raceListEtag).End()
	c.Assert(resp.StatusCode, Equals, 304)

}

func (s *TestSuite) Test09CreateRaceGroup(c *C) {

	var raceGroup api.RaceGroup
	//create a new race group
	request := gorequest.New()
	data := api.RaceGroupCreate{Name: "Marathon", Distance: "42km"}
	resp, body, _ := request.Post(fmt.Sprintf("%s/feed/racegroup", s.host)).
		Send(data).
		End()
	c.Assert(resp.StatusCode, Equals, 201)
	jsonBlob := []byte(body)
	json.Unmarshal(jsonBlob, &raceGroup)
	c.Assert(raceGroup.Name, Equals, "Marathon")
	c.Assert(raceGroup.Distance, Equals, "42km")
	c.Assert(raceGroup.SelfPath, Equals, s.host+"/feed/racegroup/1")

	//fetch the race group via its self path
	request = gorequest.New()
	resp, body, _ = request.Get(raceGroup.SelfPath).End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	json.Unmarshal(jsonBlob, &raceGroup)
	c.Assert(raceGroup.Name, Equals, "Marathon")
	c.Assert(raceGroup.Distance, Equals, "42km")
	c.Assert(raceGroup.SelfPath, Equals, s.host+"/feed/racegroup/1")

	race, _ := s.doImport("http://www.nlaa.ca/03-Road-Race.html")
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")
	c.Assert(race.Id, Equals, 1)

	//get a list of race groups
	var raceGroups api.RaceGroupFeed
	request = gorequest.New()
	resp, body, _ = request.Get(s.host + "/feed/racegroups").End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	json.Unmarshal(jsonBlob, &raceGroups)
	c.Assert(raceGroups.RaceGroups[0].Name, Equals, "Marathon")
	c.Assert(raceGroups.RaceGroups[0].Distance, Equals, "42km")
	c.Assert(raceGroups.RaceGroups[0].SelfPath, Equals, s.host+"/feed/racegroup/1")
	c.Assert(raceGroups.RaceGroups[0].RacesPath, Equals, s.host+"/feed/racegroup/1/races")

	//add the race to the group
	request = gorequest.New()
	addRace := api.RaceGroupAddRace{RaceId: strconv.Itoa(race.Id)}
	resp, _, _ = request.Post(raceGroups.RaceGroups[0].RacesPath).
		Send(addRace).
		End()
	c.Assert(resp.StatusCode, Equals, 200)

	//race should be returned on the races path now..
	request = gorequest.New()
	resp, body, _ = request.Get(raceGroups.RaceGroups[0].RacesPath).End()
	c.Assert(resp.StatusCode, Equals, 200)

	jsonBlob = []byte(body)
	var races api.RaceFeed
	json.Unmarshal(jsonBlob, &races)
	c.Assert(len(races.Races), Equals, 1)
	c.Assert(races.Races[0].RaceGroupPath, Equals, s.host+"/feed/racegroup/1")

}

func (s *TestSuite) Test10DeleteRaceGroup(c *C) {

	var raceGroups api.RaceGroupFeed
	request := gorequest.New()
	resp, body, _ := request.Get(s.host + "/feed/racegroups").End()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(resp.Header.Get("ETag"), Equals, "")

	var raceGroup api.RaceGroup
	//create a new race group
	request = gorequest.New()
	data := api.RaceGroupCreate{Name: "Marathon", Distance: "42km"}
	resp, body, _ = request.Post(fmt.Sprintf("%s/feed/racegroup", s.host)).
		Send(data).
		End()

	c.Assert(resp.StatusCode, Equals, 201)
	jsonBlob := []byte(body)
	json.Unmarshal(jsonBlob, &raceGroup)
	c.Assert(raceGroup.Name, Equals, "Marathon")
	c.Assert(raceGroup.Distance, Equals, "42km")
	c.Assert(raceGroup.SelfPath, Equals, s.host+"/feed/racegroup/1")

	request = gorequest.New()
	resp, body, _ = request.Get(s.host + "/feed/racegroups").End()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(resp.Header.Get("ETag"), Not(Equals), "")

	jsonBlob = []byte(body)
	json.Unmarshal(jsonBlob, &raceGroups)
	c.Assert(raceGroups.RaceGroups[0].Name, Equals, "Marathon")

	race, _ := s.doImport("http://www.nlaa.ca/03-Road-Race.html")
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")
	c.Assert(race.Id, Equals, 1)

	request = gorequest.New()
	resp, _, _ = request.Get(race.SelfPath).End()
	originalRaceEtag := resp.Header.Get("ETag")
	c.Assert(originalRaceEtag, Not(Equals), "")

	request = gorequest.New()
	addRace := api.RaceGroupAddRace{RaceId: strconv.Itoa(race.Id)}
	resp, _, _ = request.Post(raceGroups.RaceGroups[0].RacesPath).
		Send(addRace).
		End()
	c.Assert(resp.StatusCode, Equals, 200)

	request = gorequest.New()
	resp, _, _ = request.Get(race.SelfPath).End()
	updatedRaceEtag := resp.Header.Get("ETag")
	c.Assert(updatedRaceEtag, Not(Equals), "")
	c.Assert(updatedRaceEtag, Not(Equals), originalRaceEtag)

	deleteRequest := gorequest.New()
	resp, _, _ = deleteRequest.Delete(raceGroup.SelfPath).End()
	c.Assert(resp.StatusCode, Equals, 200)

	request = gorequest.New()
	resp, body, _ = request.Get(s.host + "/feed/racegroups").End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	json.Unmarshal(jsonBlob, &raceGroups)
	c.Assert(resp.Header.Get("ETag"), Equals, "")
	c.Assert(len(raceGroups.RaceGroups), Equals, 0)

	request = gorequest.New()
	resp, _, _ = request.Get(race.SelfPath).End()
	updatedAfterDeleteRaceEtag := resp.Header.Get("ETag")
	c.Assert(updatedAfterDeleteRaceEtag, Not(Equals), "")
	c.Assert(updatedAfterDeleteRaceEtag, Not(Equals), updatedRaceEtag)

}

func (s *TestSuite) Test11DeleteRace(c *C) {

	var races api.RaceFeed
	request := gorequest.New()
	resp, body, _ := request.Get(s.host + "/feed/races").End()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(resp.Header.Get("ETag"), Equals, "")
	jsonBlob := []byte(body)
	json.Unmarshal(jsonBlob, &races)
	c.Assert(len(races.Races), Equals, 0)

	//create a new race
	race, _ := s.doImport("http://www.nlaa.ca/03-Road-Race.html")
	c.Assert(race.Name, Equals, "Nautilus Mundy Pond 5km Road Race")
	c.Assert(race.Id, Equals, 1)

	request = gorequest.New()
	resp, body, _ = request.Get(s.host + "/feed/races").End()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(resp.Header.Get("ETag"), Not(Equals), "")
	jsonBlob = []byte(body)
	json.Unmarshal(jsonBlob, &races)
	c.Assert(len(races.Races), Equals, 1)

	deleteRequest := gorequest.New()
	resp, _, _ = deleteRequest.Delete(race.SelfPath).End()
	c.Assert(resp.StatusCode, Equals, 200)

	request = gorequest.New()
	resp, body, _ = request.Get(s.host + "/feed/races").End()
	c.Assert(resp.StatusCode, Equals, 200)
	jsonBlob = []byte(body)
	json.Unmarshal(jsonBlob, &races)
	c.Assert(resp.Header.Get("ETag"), Equals, "")
	c.Assert(len(races.Races), Equals, 0)

}

func (s *TestSuite) doImport(path string) (api.Race, error) {

	var race api.Race

	request := gorequest.New()
	url := api.DataImport{RaceUrl: path}
	resp, _, _ := request.Post(fmt.Sprintf("%s/import", s.host)).
		Send(url).
		End()

	//initial status code should be 202
	if resp.StatusCode == 202 {

		//get the location header
		var taskLocation string
		taskLocation = resp.Header.Get("Location")
		var resp gorequest.Response
		var body string

		//ping the task until it is completed
		errRetry := retry(5, func() error {
			resp, body, _ = request.Get(taskLocation).End()
			if resp.StatusCode == 200 {
				if resp.Request.URL.String() == taskLocation {
					return errors.New("Still pending")
				} else {
					return nil
				}
			} else if resp.StatusCode == 500 {
				log.Println(resp.Body)
				return errors.New("Import Failed")
			} else {
				return errors.New("Unknown status")
			}
		})
		if errRetry != nil {
			return race, errRetry
		}

		jsonBlob := []byte(body)
		err := json.Unmarshal(jsonBlob, &race)

		if err != nil {
			return race, err
		}

		return race, nil

	} else {
		return race, errors.New("nannan")
	}
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

func retry(attempts int, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return nil
		}

		if err != nil && err.Error() == "Import Failed" {
			break
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(2 * time.Second)

	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
