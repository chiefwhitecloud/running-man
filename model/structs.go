package model

type Racer struct {
	Position  int
	FirstName string
	LastName  string
	BibNumber string
	Club      string
	Time      string
	Sex       string
}

type RaceDetails struct {
	Racers []Racer
	Name   string
	Year   int
	Month  int
	Day    int
}
