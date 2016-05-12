package api

type DataImport struct {
	RaceUrl string `json:"raceUrl"`
}

type Race struct {
	Id    			int `json:"id"`
	Name        string `json:"name"`
	SelfPath    string `json:"selfPath"`
	ResultsPath string `json:"resultsPath"`
	Date        string `json:"date"`
}

type Racer struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Sex         string `json:"sex"`
	SelfPath    string `json:"selfPath"`
	ResultsPath string `json:"resultsPath"`
	BirthDateLow string `json:"birthDateLow,omitempty"`
	BirthDateHigh string `json:"birthDateHigh,omitempty"`	
}

type RaceResults struct {
	Racers  map[string]Racer `json:"racers"`
	Races   map[string]Race  `json:"races"`
	Results []RaceResult     `json:"results"`
}

type RaceResult struct {
	Time                string `json:"time"`
	Position            int    `json:"position"`
	SexPosition         int    `json:"sexPosition"`
	AgeCategoryPosition int    `json:"ageCategoryPosition"`
	RacerID             string `json:"racerId"`
	RaceID              string `json:"raceId"`
	BibNumber           string `json:"bibNumber"`
	AgeCategory         string `json:"ageCategory"`
}

type RaceFeed struct {
	Races []Race `json:"races"`
}
