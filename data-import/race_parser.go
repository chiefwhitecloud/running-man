package dataimport

import (
	"bytes"
	"errors"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/chiefwhitecloud/running-man/model"
	"golang.org/x/net/html"
)

var _ = log.Print

func parseResults(htmlresult []byte) (model.RaceDetails, error) {

	z := html.NewTokenizer(bytes.NewReader(htmlresult))
	found := false
	preMode := false
	foundTitle := false
	foundAddress := false
	var results string
	var resultsTitle string
	var resultsAddress string
	var raceDay, raceYear, raceMonth int
	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			break
		} else if preMode && tt == html.StartTagToken {

		} else if tt == html.EndTagToken {
			t := z.Token()
			if t.Data == "address" {
				foundAddress = false
			} else if t.Data == "pre" {
				found = false
				preMode = false
			}
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "pre" {
				found = true
				preMode = true
			} else if t.Data == "title" {
				foundTitle = true
			} else if t.Data == "address" {
				foundAddress = true
			}
		} else if tt == html.TextToken {
			if found {
				results = results + string(z.Text())
			}

			if foundTitle {
				resultsTitle = string(z.Text())
				foundTitle = false
			}

			if foundAddress {
				resultsAddress = resultsAddress + string(z.Text())
				foundAddress = false
			}
		}
	}

	if results == "" {
		return model.RaceDetails{}, errors.New("Results not found in HTML")
	}

	raceRows := strings.Split(results, "\n")
	raceRows = append(raceRows[:0], raceRows[1:]...)

	if resultsTitle == "" {
		return model.RaceDetails{}, errors.New("Title tag not found in HTML")
	}

	raceTitles := strings.Split(resultsTitle, ":")
	if len(raceTitles) > 0 {
		resultsTitle = strings.Trim(raceTitles[len(raceTitles)-1], " ")
	}

	if resultsAddress == "" {
		//maybe the date it is in the first line of the pre tag..
		resultsAddress = raceRows[0]
	} else {
		resultsAddress = strings.Replace(resultsAddress, "\n", "", -1)
	}

	dateReg := regexp.MustCompile(`(?P<month>January|February|March|April|May|June|July|August|September|October|November|December)[ ](?P<day>0?[1-9]|[1-2][0-9]|3[0-1])(st|nd|rd|th)?[,][ ](?P<year>20[0-9]{2})`)
	r3 := dateReg.FindAllStringSubmatch(resultsAddress, -1)
	var r4 []string
	if r3 == nil {
		log.Println(resultsAddress)
		return model.RaceDetails{}, errors.New("Date not found in " + resultsTitle)
	} else {
		r4 = r3[0]
	}

	monthMap := map[string]int{
		"JANUARY":   1,
		"FEBRUARY":  2,
		"MARCH":     3,
		"APRIL":     4,
		"MAY":       5,
		"JUNE":      6,
		"JULY":      7,
		"AUGUST":    8,
		"SEPTEMBER": 9,
		"OCTOBER":   10,
		"NOVEMBER":  11,
		"DECEMBER":  12,
	}

	if len(r4) > 0 {
		raceMonth = monthMap[strings.ToUpper(r4[1])]
		raceDay, _ = strconv.Atoi(r4[2])
		raceYear, _ = strconv.Atoi(r4[4])
	} else {
		return model.RaceDetails{}, errors.New("Could not find race date")
	}

	const position = `^[ ]*(?P<position>\d+)`
	const bibName = `[ ]+(?P<bib_number>\d+)[ ]+(?P<name>[\D\(\)]+)`
	const time = `(?P<time>[\:\d]+)`
	const chiptime = `(?P<chiptime>[\:\d]+|[ ]+)`
	const pace = `(?P<pace>[\:\d]+|[ ]+)`
	const spaceOrMore = `[ ]+`
	const sexPosition = `[\(]?(?P<sex_pos>\d+)[\)]?`
	const sex = `(?P<sex>M|F|W|P)`
	const category = `(?P<category>A|-19|U20|\<20|NOAGE|70\+|80\+|\d\d-\d\d|)`
	const categoryPosition = `((?P<category_position>\d+)(\/[\d]+)?)`

	const clubNameRegEx = `[^\(]*\((?P<club>\w{2,4})\)`

	mainRegEx := position + bibName + time + spaceOrMore + sex + sexPosition + spaceOrMore + category + spaceOrMore + categoryPosition

	re := regexp.MustCompile(mainRegEx)
	//Tely 10 Result
	re2 := regexp.MustCompile(position + bibName + time + spaceOrMore + `L?` + sex + category + spaceOrMore + categoryPosition + spaceOrMore + sexPosition + spaceOrMore + pace + spaceOrMore + chiptime)
	//Other format
	re3 := regexp.MustCompile(position + bibName + time + `.{1}` + spaceOrMore + sex + sexPosition + spaceOrMore + category + spaceOrMore + categoryPosition)
	re4 := regexp.MustCompile(position + bibName + time + spaceOrMore + sex + spaceOrMore + categoryPosition + spaceOrMore + sexPosition)
	re5 := regexp.MustCompile(`^\s{0,}\d{1,} `)

	clubName := regexp.MustCompile(clubNameRegEx)

	n1 := re.SubexpNames()
	n2 := re2.SubexpNames()
	n3 := re3.SubexpNames()

	raceResultsMap := make(map[int]model.Racer)

	for i := 0; i < len(raceRows); i++ {
		var r2 [][]string

		md := map[string]string{}

		if re.MatchString(raceRows[i]) {
			r2 = re.FindAllStringSubmatch(raceRows[i], -1)
			for i, n := range r2[0] {
				md[n1[i]] = n
			}
		} else if re2.MatchString(raceRows[i]) {
			r2 = re2.FindAllStringSubmatch(raceRows[i], -1)
			for i, n := range r2[0] {
				md[n2[i]] = n
			}
		} else if (i+1) < (len(raceRows)-1) && re3.MatchString(raceRows[i]+raceRows[i+1]) {
			//multiline race result
			r2 = re3.FindAllStringSubmatch(raceRows[i]+raceRows[i+1], -1)
			i++
			for i, n := range r2[0] {
				md[n3[i]] = n
			}
		} else if re4.MatchString(raceRows[i]) {
			r2 = re4.FindAllStringSubmatch(raceRows[i], -1)
			for i, n := range r2[0] {
				md[n2[i]] = n
			}
		} else if re5.MatchString(raceRows[i]) {
			log.Println("Failed to parse result with " + re4.String())
			return model.RaceDetails{}, errors.New("Failed to parse line : " + raceRows[i])
		} else {
			//log.Println("Skipping line in race result for " + resultsTitle)
			//log.Println(raceRows[i])
		}

		if len(r2) > 0 {

			p, _ := strconv.Atoi(md["position"])

			sp, _ := strconv.Atoi(md["sex_pos"])

			ap, _ := strconv.Atoi(md["category_position"])

			var club []string

			var runnersClubName string

			if clubName.MatchString(md["name"]) {
				club = clubName.FindStringSubmatch(md["name"])
				//extract club name and remove it from the racers name
				runnersClubName = club[1]
				md["name"] = strings.Replace(md["name"], "("+runnersClubName+")", "", 1)
			}

			raceResultsMap[p] = model.Racer{
				Position:            p,
				Name:                strings.TrimSpace(md["name"]),
				BibNumber:           md["bib_number"],
				Club:                runnersClubName,
				Time:                md["time"],
				Sex:                 md["sex"],
				SexPosition:         sp,
				AgeCategory:         md["category"],
				AgeCategoryPosition: ap,
				ChipTime:            md["chiptime"],
			}
		}
	}

	if len(raceResultsMap) == 0 {
		return model.RaceDetails{}, errors.New("Failed to parse race results")
	}

	var keys []int
	var racerResults []model.Racer

	for k := range raceResultsMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		racerResults = append(racerResults, raceResultsMap[k])
	}

	race := model.RaceDetails{Racers: racerResults, Name: resultsTitle, Year: raceYear, Month: raceMonth, Day: raceDay}

	return race, nil
}
