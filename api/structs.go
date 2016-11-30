package api

type DataImport struct {
	RaceUrl string `json:"raceUrl"`
}

type RacerMerge struct {
	RacerId string `json:"racerId"`
}

type RaceGroupAddRace struct {
	RaceId string `json:"raceId"`
}

type RaceGroupCreate struct {
	Name     string `json:"name"`
	Distance string `json:"distance"`
}

type RaceGroup struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Distance  string `json:"distance"`
	SelfPath  string `json:"self"`
	RacesPath string `json:"races"`
}

type Race struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	SelfPath      string `json:"self"`
	ResultsPath   string `json:"results"`
	Date          string `json:"date"`
	RaceGroupPath string `json:"raceGroup,omitempty"`
}

type Racer struct {
	Id          int    `json:"id"`
	SelfPath    string `json:"self"`
	ResultsPath string `json:"results"`
	ProfilePath string `json:"profile"`
	MergePath   string `json:"merge"`
}

type RacerProfile struct {
	Name          string   `json:"name"`
	NameList      []string `json:"nameList"`
	SelfPath      string   `json:"self"`
	BirthDateLow  string   `json:"birthDateLow"`
	BirthDateHigh string   `json:"birthDateHigh"`
}

type RaceResults struct {
	Racers  map[string]Racer `json:"racers"`
	Races   map[string]Race  `json:"races"`
	Results []RaceResult     `json:"results"`
}

type RaceResult struct {
	Name                string `json:"name"`
	Time                string `json:"time"`
	Position            int    `json:"position"`
	SexPosition         int    `json:"sexPosition"`
	AgeCategoryPosition int    `json:"ageCategoryPosition"`
	RacerID             string `json:"racerId"`
	RaceID              string `json:"raceId"`
	BibNumber           string `json:"bibNumber"`
	AgeCategory         string `json:"ageCategory"`
	Sex                 string `json:"sex"`
	Club                string `json:"club,omitempty"`
	ChipTime            string `json:"chipTime,omitempty"`
}

type RaceFeed struct {
	Races []Race `json:"races"`
}

type RaceGroupFeed struct {
	RaceGroups []RaceGroup `json:"raceGroups"`
}
