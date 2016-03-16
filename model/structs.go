package model

import (
	"time"
)

type Racer struct {
	Position            int
	FirstName           string
	LastName            string
	BibNumber           string
	Club                string
	Time                string
	Sex                 string
	SexPosition         int
	AgeCategory         string
	AgeCategoryPosition int
	LowBirthdayDate     time.Time
	HighBirthdayDate    time.Time
}

type RaceDetails struct {
	Racers []Racer
	Name   string
	Year   int
	Month  int
	Day    int
}
