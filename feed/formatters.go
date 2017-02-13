package feed

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/chiefwhitecloud/running-man/api"
	"github.com/chiefwhitecloud/running-man/database"
)

func FormatImportTaskLocation(req *http.Request, taskId int) string {
	return fmt.Sprintf("http://%s/import/task/%d", req.Host, taskId)
}

func FormatRaceLocation(req *http.Request, raceId int) string {
	return fmt.Sprintf("http://%s/feed/race/%d", req.Host, raceId)
}

func FormatRaceGroupsForFeed(req *http.Request, raceGroups []database.RaceGroup) api.RaceGroupFeed {

	raceGroupList := make([]api.RaceGroup, len(raceGroups))
	for i, _ := range raceGroups {
		raceGroupList[i] = FormatRaceGroupForFeed(req, raceGroups[i])
	}

	return api.RaceGroupFeed{RaceGroups: raceGroupList}
}

func FormatRaceGroupForFeed(req *http.Request, raceGroup database.RaceGroup) api.RaceGroup {
	return api.RaceGroup{
		Id:        raceGroup.ID,
		Name:      raceGroup.Name,
		Distance:  raceGroup.Distance,
		SelfPath:  fmt.Sprintf("http://%s/feed/racegroup/%d", req.Host, raceGroup.ID),
		RacesPath: fmt.Sprintf("http://%s/feed/racegroup/%d/races", req.Host, raceGroup.ID),
	}
}

func FormatRacesForFeed(req *http.Request, races []database.Race) api.RaceFeed {

	raceList := make([]api.Race, len(races))
	for i, _ := range races {
		raceList[i] = FormatRaceForFeed(req, races[i])
	}

	return api.RaceFeed{Races: raceList}
}

func FormatRaceForFeed(req *http.Request, race database.Race) api.Race {
	raceStruct := api.Race{
		Id:          race.ID,
		Name:        race.Name,
		SelfPath:    fmt.Sprintf("http://%s/feed/race/%d", req.Host, race.ID),
		ResultsPath: fmt.Sprintf("http://%s/feed/race/%d/results", req.Host, race.ID),
		Date:        fmt.Sprintf("%0.4d-%0.2d-%0.2d", race.Date.Year(), race.Date.Month(), race.Date.Day()),
	}

	if race.RaceGroupID > 0 {
		raceStruct.RaceGroupPath = fmt.Sprintf("http://%s/feed/racegroup/%d", req.Host, race.RaceGroupID)
	}

	return raceStruct
}

func FormatRacerForFeed(req *http.Request, racer database.Racer) api.Racer {
	return api.Racer{
		Id:          racer.ID,
		SelfPath:    fmt.Sprintf("http://%s/feed/racer/%d", req.Host, racer.ID),
		ResultsPath: fmt.Sprintf("http://%s/feed/racer/%d/results", req.Host, racer.ID),
		ProfilePath: fmt.Sprintf("http://%s/feed/racer/%d/profile", req.Host, racer.ID),
		MergePath:   fmt.Sprintf("http://%s/feed/racer/%d/merge", req.Host, racer.ID),
	}
}

func FormatRaceResultsForFeed(req *http.Request, raceresults []database.RaceResult, racers []database.Racer, races []database.Race) api.RaceResults {

	ageMap := map[int]string{
		1:  "U20",
		2:  "-19",
		3:  "<20",
		4:  "20-24",
		5:  "25-29",
		6:  "20-29",
		7:  "30-34",
		8:  "35-39",
		9:  "30-39",
		10: "40-44",
		11: "45-49",
		12: "40-49",
		13: "50-54",
		14: "55-59",
		15: "50-59",
		16: "60-64",
		17: "65-69",
		18: "60-69",
		19: "70-74",
		20: "75-79",
		21: "70-79",
		22: "70+",
		23: "80-84",
		24: "85-89",
		25: "80-89",
		26: "80+",
		27: "A",
		28: "NOAGE",
	}

	mapRacers := map[string]api.Racer{}
	for i := range racers {
		mapRacers[strconv.Itoa(racers[i].ID)] = FormatRacerForFeed(req, racers[i])
	}

	mapRaces := map[string]api.Race{}
	for i := range races {
		mapRaces[strconv.Itoa(races[i].ID)] = FormatRaceForFeed(req, races[i])
	}

	rr := make([]api.RaceResult, len(raceresults))
	for i := range raceresults {
		rr[i] = api.RaceResult{
			Name:                raceresults[i].Name,
			Position:            raceresults[i].Position,
			SexPosition:         raceresults[i].SexPosition,
			Sex:                 raceresults[i].Sex,
			AgeCategoryPosition: raceresults[i].AgeCategoryPosition,
			RacerID:             strconv.Itoa(raceresults[i].RacerID),
			RaceID:              strconv.Itoa(raceresults[i].RaceID),
			BibNumber:           raceresults[i].BibNumber,
			Time:                raceresults[i].Time,
			AgeCategory:         ageMap[raceresults[i].AgeCategoryID],
			Club:                raceresults[i].Club,
			ChipTime:            raceresults[i].ChipTime,
		}
	}

	return api.RaceResults{Results: rr, Racers: mapRacers, Races: mapRaces}
}
