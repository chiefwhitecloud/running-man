package dataimport

import (
	"bytes"
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/chiefwhitecloud/running-man/model"
	"golang.org/x/net/html"
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

		if err != nil {
			log.Println(err)
		}
	} else {
		return model.RaceDetails{}, errors.New("Could not find race date")
	}

	re := regexp.MustCompile(`^(?P<position>\d+)\s{3,}(?P<bib_number>\d+)\s{1,}(?P<name>.*?)(\s{2,}|\s\((?P<club>[A-Z{2,4}]+\)))\s{2,}(?P<time>[0-9\\:]+)\s{2,}(?P<sex>[MF])\((?P<sex_pos>\d+)(.*)\)\s{2,}(?P<category>U20|70\+|\d\d-\d\d)\s{2,}(?P<category_position>\d+)(.*)`)
	re2 := regexp.MustCompile(`^\s{0,}(?P<position>\d+)\s{2,}(?P<bib_number>\d+)\s{1,}(?P<name>.*?)(\s{2,}|\s\((?P<club>[A-Z{2,4}]+\)))\s{2,}(?P<time>[0-9\:]+)\s{1,}L(?P<sex>M|F|W)(?P<category>-19|80\+|A|\d\d-\d\d)\s{2,}(?P<category_position>\d+)\/[\d]+\s{2,}(?P<sex_pos>\d+)`)
	re3 := regexp.MustCompile(`(?s)^(?P<position>\d+)\s{3,}(?P<bib_number>\d+)\s{1,}(?P<name>.*?)(\s{2,}|\s\((?P<club>[A-Z{2,4}]+)\)).{1}\s{2,}(?P<time>[0-9\:]+)\s{2,}(?P<sex>[MF])\((?P<sex_pos>\d+)(.*)\)\s{2,}(?P<category>U20|70\+|\d\d-\d\d)\s{2,}(?P<category_position>\d+)(.*)`)
	re4 := regexp.MustCompile(`^\s{0,}\d`)
	n1 := re.SubexpNames()
	n2 := re2.SubexpNames()
	n3 := re3.SubexpNames()

	raceRows := strings.Split(results, "\n")

	raceRows = append(raceRows[:0], raceRows[1:]...)

	var racerResults []model.Racer

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
			//multiline race result
			log.Println("Failed to parse result for " + resultsTitle)
			log.Println(raceRows[i])
		} else {
			log.Println("Skipping line in race result for " + resultsTitle)
			log.Println(raceRows[i])
		}

		if len(r2) > 0 {

			p, _ := strconv.Atoi(md["position"])

			sp, _ := strconv.Atoi(md["sex_pos"])

			ap, _ := strconv.Atoi(md["category_position"])

			racerResults = append(racerResults, model.Racer{Position: p,
				Name:                md["name"],
				BibNumber:           md["bib_number"],
				Club:                md["club"],
				Time:                md["time"],
				Sex:                 md["sex"],
				SexPosition:         sp,
				AgeCategory:         md["category"],
				AgeCategoryPosition: ap,
			})
		}
	}

	if len(racerResults) == 0 {
		return model.RaceDetails{}, errors.New("Failed to parse race results")
	}

	race := model.RaceDetails{Racers: racerResults, Name: resultsTitle, Year: raceYear, Month: raceMonth, Day: raceDay}

	return race, nil
}
