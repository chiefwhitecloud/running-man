package model

type Racer struct {
	Position            int
	Name                string
	BibNumber           string
	Club                string
	Time                string
	ChipTime            string
	Sex                 string
	SexPosition         int
	AgeCategory         string
	AgeCategoryPosition int
}

type RaceDetails struct {
	Racers []Racer
	Name   string
	Year   int
	Month  int
	Day    int
}
