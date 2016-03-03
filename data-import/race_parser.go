package dataimport

import (
	"bytes"
	"errors"
	"github.com/chiefwhitecloud/running-man/model"
	"golang.org/x/net/html"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var _ = log.Print

func parseResults(htmlresult []byte) (model.RaceDetails, error) {

	z := html.NewTokenizer(bytes.NewReader(htmlresult))
	found := false
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
		} else if tt == html.EndTagToken {
			t := z.Token()
			if t.Data == "address" {
				foundAddress = false
			}
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "pre" {
				found = true
			} else if t.Data == "title" {
				foundTitle = true
			} else if t.Data == "address" {
				foundAddress = true
			}
		} else if tt == html.TextToken {
			if found {
				results = string(z.Text())
				found = false
			}

			if foundTitle {
				resultsTitle = string(z.Text())
				foundTitle = false
			}

			if foundAddress {
				resultsAddress = resultsAddress + string(z.Text())
			}
		}
	}

	raceTitles := strings.Split(resultsTitle, ":")
	if len(raceTitles) > 0 {
		resultsTitle = strings.Trim(raceTitles[1], " ")
	}

	resultsAddress = strings.Replace(resultsAddress, "\n", "", -1)

	if results == "" {
		return model.RaceDetails{}, errors.New("Results not found in HTML")
	}

	dateReg, err := regexp.Compile("(?P<month>January|February|March|April|May|June|July|August|September|October|November|December)[ ](?P<day>0[1-9]|[1-2][0-9]|3[0-1])(st|nd|rd|th)?[,][ ](?P<year>20[0-9]{2})")
	r3 := dateReg.FindAllStringSubmatch(resultsAddress, -1)[0]

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

	if len(r3) > 0 {
		raceMonth = monthMap[strings.ToUpper(r3[1])]
		raceDay, err = strconv.Atoi(r3[2])
		raceYear, err = strconv.Atoi(r3[4])
	} else {
		return model.RaceDetails{}, errors.New("Could not find race date")
	}

	re, err := regexp.Compile("^(?P<position>\\d+)\\s{3,}(?P<bib_number>\\d+)\\s{1,}(?P<first_name>[a-zA-Z0-9]+)\\s(?P<last_name>[a-zA-Z0-9]+)(\\s(\\((?P<club>[A-Z]+)\\))?)\\s{2,}(?P<time>[0-9\\:]+)\\s{2,}(?P<sex>[MF])(.*)")
	n1 := re.SubexpNames()

	if err != nil {
		log.Println(err)
	}

	raceRows := strings.Split(results, "\n")

	raceRows = append(raceRows[:0], raceRows[1:]...)

	var racerResults []model.Racer

	for i := range raceRows {
		r2 := re.FindAllStringSubmatch(raceRows[i], -1)
		if len(r2) > 0 {
			md := map[string]string{}
			for i, n := range r2[0] {
				md[n1[i]] = n
			}

			p, _ := strconv.Atoi(md["position"])

			racerResults = append(racerResults, model.Racer{Position: p, FirstName: md["first_name"], LastName: md["last_name"], BibNumber: md["bib_number"], Club: md["club"], Time: md["time"], Sex: md["sex"]})
		}
	}

	race := model.RaceDetails{Racers: racerResults, Name: resultsTitle, Year: raceYear, Month: raceMonth, Day: raceDay}

	return race, nil
}
