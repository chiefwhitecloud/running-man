package dataimport

import (
	"bytes"
	"errors"
	"github.com/chiefwhitecloud/running-man/Godeps/_workspace/src/golang.org/x/net/html"
	"github.com/chiefwhitecloud/running-man/model"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var _ = log.Print

type AgeCategory struct {
	MinAge int
	MaxAge int
}

func (age *AgeCategory) GetBirthdayDateRange(givenDate time.Time) (time.Time, time.Time, error) {
	//based on the given date what are the high and low range for when their birthday could be
	lowDate := givenDate.AddDate((-1 * age.MinAge), 0, 0)
	highDate := givenDate.AddDate(0, 0, -1)
	return lowDate, highDate, nil
}

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

	ageCategoryMap := map[string]*AgeCategory{
		"U20":     &AgeCategory{1, 19},
		"20-29":   &AgeCategory{20, 29},
		"30-39":   &AgeCategory{30, 39},
		"40-49":   &AgeCategory{40, 49},
		"50-59":   &AgeCategory{50, 59},
		"60-69":   &AgeCategory{60, 69},
		"70-79":   &AgeCategory{70, 79},
		"80-89":   &AgeCategory{80, 89},
		"90-99":   &AgeCategory{90, 99},
		"100-109": &AgeCategory{100, 109},
	}

	if len(r3) > 0 {
		raceMonth = monthMap[strings.ToUpper(r3[1])]
		raceDay, err = strconv.Atoi(r3[2])
		raceYear, err = strconv.Atoi(r3[4])
	} else {
		return model.RaceDetails{}, errors.New("Could not find race date")
	}

	raceDate := time.Date(raceYear, time.Month(raceMonth), raceDay, 0, 0, 0, 0, time.UTC)

	re, err := regexp.Compile("^(?P<position>\\d+)\\s{3,}(?P<bib_number>\\d+)\\s{1,}(?P<first_name>[a-zA-Z0-9]+)\\s(?P<last_name>[a-zA-Z0-9]+)(\\s(\\((?P<club>[A-Z]+)\\))?)\\s{2,}(?P<time>[0-9\\:]+)\\s{2,}(?P<sex>[MF])\\((?P<sex_pos>\\d+)(.*)\\)\\s{2,}(?P<category>U20|\\d\\d-\\d\\d)\\s{2,}(?P<category_position>\\d+)(.*)")

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

			sp, _ := strconv.Atoi(md["sex_pos"])

			ap, _ := strconv.Atoi(md["category_position"])

			low, high, _ := ageCategoryMap[md["category"]].GetBirthdayDateRange(raceDate)

			racerResults = append(racerResults, model.Racer{Position: p,
				FirstName:           md["first_name"],
				LastName:            md["last_name"],
				BibNumber:           md["bib_number"],
				Club:                md["club"],
				Time:                md["time"],
				Sex:                 md["sex"],
				SexPosition:         sp,
				AgeCategory:         md["category"],
				AgeCategoryPosition: ap,
				LowBirthdayDate:     low,
				HighBirthdayDate:    high,
			})
		}
	}

	race := model.RaceDetails{Racers: racerResults, Name: resultsTitle, Year: raceYear, Month: raceMonth, Day: raceDay}

	return race, nil
}
