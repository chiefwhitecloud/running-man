package database

import (
	"github.com/chiefwhitecloud/running-man/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"sort"
	"time"
)

type Db struct {
	orm              gorm.DB
	ConnectionString string
}

type Racer struct {
	ID        int
	FirstName string
	LastName  string
	Sex       string
}

type Race struct {
	ID    int
	Name  string
	Year  int
	Month int
	Day   int
}

type RaceResult struct {
	ID                  int
	Position            int
	SexPosition         int
	AgeCategoryPosition int
	RaceID              int `sql:"index"`
	RacerID             int `sql:"index"`
	AgeCategoryID       int `sql:"index"`
	BibNumber           string
	Time                string
	Racer               Racer
	Race                Race
}

type AgeCategory struct {
	ID   int
	Name string
}

type AgeLookup struct {
	minAge int
	maxAge int
}

type raceResultForTransform struct {
	pos   int
	first string
	last  string
}

type AgeResult struct {
	RaceDate    time.Time
	AgeCategory string
}

type AgeResults []AgeResult

func (slice AgeResults) Len() int {
	return len(slice)
}

func (slice AgeResults) Less(i, j int) bool {
	return slice[i].RaceDate.Before(slice[j].RaceDate)
}

func (slice AgeResults) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (db *Db) Migrate() {
	db.orm.AutoMigrate(&Racer{}, &Race{}, &RaceResult{}, &AgeCategory{})

	cats := []string{"U20", "20-29", "30-39", "40-49", "50-59", "60-69", "70-79", "80-89", "90-99", "100-109"}

	for i := 0; i < len(cats); i++ {
		cat := AgeCategory{Name: cats[i]}
		db.orm.Create(&cat)
	}

}

func (db *Db) Create() {
	db.orm.CreateTable(&Racer{}, &Race{}, &RaceResult{}, &AgeCategory{})

}

func (db *Db) DropAllTables() {
	db.orm.DropTable(&Racer{}, &Race{}, &RaceResult{}, &AgeCategory{})
}

func (db *Db) Open() error {

	gormdb, err := gorm.Open("mysql", db.ConnectionString)
	if err != nil {
		return err
	}
	gormdb.SingularTable(true)

	db.orm = gormdb

	return nil
}

func (db *Db) SaveRace(r *model.RaceDetails) (Race, error) {

	cats := []AgeCategory{}

	db.orm.Find(&cats)

	race := Race{Name: r.Name,
		Year:  r.Year,
		Month: r.Month,
		Day:   r.Day}

	db.orm.Create(&race)

	raceDate := time.Date(r.Year, time.Month(r.Month), r.Day, 0, 0, 0, 0, time.UTC)

	//save the race results information
	for i := range r.Racers {

		mRacer := r.Racers[i]

		var racers []Racer
		var racer Racer

		db.orm.Where(&Racer{FirstName: mRacer.FirstName, LastName: mRacer.LastName, Sex: mRacer.Sex}).Find(&racers)

		if len(racers) == 0 {
			//must be a new racer
			racer = Racer{
				FirstName: mRacer.FirstName,
				LastName:  mRacer.LastName,
				Sex:       mRacer.Sex,
			}
			db.orm.Create(&racer)
		} else if len(racers) > 0 {
			//look at the racers age catgory history... does it look like a match?
			for i := range racers {
				early, late, _ := db.GetRacerBirthDates(racers[i].ID)
				minAge, maxAge, _ := db.GetAgeRangeOnDate(early, late, raceDate)

				if db.isAgeRangeWithinCatgory(maxAge, minAge, mRacer.AgeCategory) {
					//existing racer is found
					racer = racers[i]
					break
				}
			}

			if racer == (Racer{}) {
				racer = Racer{
					FirstName: mRacer.FirstName,
					LastName:  mRacer.LastName,
					Sex:       mRacer.Sex,
				}
				db.orm.Create(&racer)
			}

			//XXX Fix me: handle two runners with the same name, sex, and age category in the same race.

		}

		//find the agecategory id.
		catId := 0
		for i := range cats {
			if cats[i].Name == mRacer.AgeCategory {
				catId = cats[i].ID
			}
		}

		result := RaceResult{
			RaceID:              race.ID,
			RacerID:             racer.ID,
			Position:            mRacer.Position,
			BibNumber:           mRacer.BibNumber,
			SexPosition:         mRacer.SexPosition,
			AgeCategoryPosition: mRacer.AgeCategoryPosition,
			AgeCategoryID:       catId,
			Time:                mRacer.Time,
		}

		db.orm.Create(&result)

	}

	return race, nil
}

func (db *Db) GetRaces() ([]Race, error) {
	races := []Race{}
	db.orm.Find(&races)
	return races, nil
}

func (db *Db) GetRace(id int) (Race, error) {
	race := Race{}
	db.orm.First(&race, id)
	return race, nil
}

func (db *Db) GetRacer(id int) (Racer, error) {
	racer := Racer{}
	db.orm.First(&racer, id)
	return racer, nil
}

func (db *Db) isAgeRangeWithinCatgory(minAge int, maxAge int, ageCategory string) bool {
	catMinAge, catMaxAge, _ := db.GetMinMaxAgeForCategory(ageCategory)
	return minAge >= catMinAge && maxAge <= catMaxAge
}

func (db *Db) GetMinMaxAgeForCategory(ageCategory string) (int, int, error) {
	ageCategoryMap := map[string]*AgeLookup{
		"U20":     &AgeLookup{1, 19},
		"20-29":   &AgeLookup{20, 29},
		"30-39":   &AgeLookup{30, 39},
		"40-49":   &AgeLookup{40, 49},
		"50-59":   &AgeLookup{50, 59},
		"60-69":   &AgeLookup{60, 69},
		"70-79":   &AgeLookup{70, 79},
		"80-89":   &AgeLookup{80, 89},
		"90-99":   &AgeLookup{90, 99},
		"100-109": &AgeLookup{100, 109},
	}

	age := ageCategoryMap[ageCategory]

	return age.minAge, age.maxAge, nil
}

func (db *Db) GetAgeRangeOnDate(earlyBirthDate time.Time, lateBirthDate time.Time, raceDate time.Time) (int, int, error) {
	earlyDate := raceDate.Sub(earlyBirthDate)
	years := earlyDate / time.Hour / 24 / 365

	lateDate := raceDate.Sub(lateBirthDate)
	minyears := lateDate / time.Hour / 24 / 365

	return int(minyears), int(years), nil
}

func (db *Db) GetBirthDateRangeForCategory(raceDate time.Time, ageCategory string) (time.Time, time.Time, error) {
	var earlyDate time.Time
	var lateDate time.Time

	catMinAge, catMaxAge, _ := db.GetMinMaxAgeForCategory(ageCategory)

	earlyDate = raceDate.AddDate(-catMaxAge-1, 0, 1)
	lateDate = earlyDate.AddDate(catMaxAge-catMinAge+1, 0, -1)

	return earlyDate, lateDate, nil

}

func (db *Db) GetRacerBirthDates(id int) (time.Time, time.Time, error) {

	rows, err := db.orm.Table("race_result").
		Select("race.year, race.month, race.day, age_category.name").
		Joins("join race on race_result.race_id = race.id join racer on race_result.racer_id = racer.id join age_category on age_category.id = race_result.age_category_id").
		Where("race_result.racer_id = ?", id).
		Rows()

	if err != nil {
		log.Println(err)
	}

	var (
		year   int
		month  int
		day    int
		agecat string
	)

	var results AgeResults

	for rows.Next() {
		err := rows.Scan(&year, &month, &day, &agecat)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, AgeResult{
			RaceDate:    time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC),
			AgeCategory: agecat,
		})
	}

	//sort the results from earliest to latest
	sort.Sort(results)

	var high time.Time
	var low time.Time
	var ageCat string
	var lastRaceDate time.Time

	for i := range results {

		if ageCat == "" && lastRaceDate.IsZero() {
			ageCat = results[i].AgeCategory
			lastRaceDate = results[i].RaceDate
			low, high, _ = db.GetBirthDateRangeForCategory(results[i].RaceDate, results[i].AgeCategory)
		}

		if ageCat != results[i].AgeCategory {

			lowforCat, highforCat, _ := db.GetBirthDateRangeForCategory(results[i].RaceDate, results[i].AgeCategory)

			if lowforCat.After(low) {
				low = lowforCat
			}

			if highforCat.Before(high) {
				high = highforCat
			}
		}

	}

	return low, high, nil
}

func (db *Db) GetRaceResultsForRace(raceid uint) ([]RaceResult, []Racer, []Race, error) {

	// XXX: Maybe a better way to do this using the ORM.  Couldn't figure it out.
	// For now doing a manual join and populating the struct to return.  Seems like the
	// ORM should be doing some more work here.

	r := Race{}

	db.orm.Find(&r, raceid)

	rows, err := db.orm.Table("race_result").
		Select("race_result.time, race_result.position, race_result.sex_position, race_result.age_category_position, race_result.bib_number, racer.first_name, racer.last_name, racer.id, race_result.id, racer.sex, race_result.age_category_id").
		Joins("join racer on race_result.racer_id = racer.id").
		Where("race_result.race_id = ?", r.ID).
		Rows()

	if err != nil {
		log.Println(err)
	}

	var (
		time                string
		position            int
		sexposition         int
		agecategoryposition int
		bibnumber           string
		firstname           string
		lastname            string
		racerid             int
		raceresultid        int
		sex                 string
		agecat              int
	)

	var results []RaceResult
	var races []Race
	var racers []Racer
	races = append(races, r)

	for rows.Next() {
		err := rows.Scan(&time, &position, &sexposition, &agecategoryposition, &bibnumber, &firstname, &lastname, &racerid, &raceresultid, &sex, &agecat)
		if err != nil {
			log.Fatal(err)
		}

		xx := RaceResult{
			ID:                  raceresultid,
			Time:                time,
			Position:            position,
			SexPosition:         sexposition,
			AgeCategoryPosition: agecategoryposition,
			RaceID:              r.ID,
			RacerID:             racerid,
			BibNumber:           bibnumber,
			AgeCategoryID:       agecat,
		}

		results = append(results, xx)

		racers = append(racers, Racer{
			ID:        racerid,
			FirstName: firstname,
			LastName:  lastname,
			Sex:       sex,
		})
	}

	return results, racers, races, nil
}

func (db *Db) GetRaceResultsForRacer(racerid uint) ([]RaceResult, []Racer, []Race, error) {

	// XXX: Maybe a better way to do this using the ORM.  Couldn't figure it out.
	// For now doing a manual join and populating the struct to return.  Seems like the
	// ORM should be doing some more work here.

	r := Racer{}

	db.orm.Find(&r, racerid)

	rows, err := db.orm.Table("race_result").
		Select("race_result.time, race_result.position, race_result.sex_position, race_result.age_category_position, race_result.bib_number, race.name,  race.id, race_result.id, race.year, race.month, race.day, race_result.age_category_id").
		Joins("join race on race_result.race_id = race.id").
		Where("race_result.racer_id = ?", r.ID).
		Rows()

	if err != nil {
		log.Println(err)
	}

	var (
		time                string
		position            int
		sexposition         int
		agecategoryposition int
		bibnumber           string
		name                string
		raceid              int
		raceresultid        int
		year                int
		month               int
		day                 int
		agecat              int
	)

	var results []RaceResult
	var races []Race
	var racers []Racer

	racers = append(racers, r)

	for rows.Next() {
		err := rows.Scan(&time, &position, &sexposition, &agecategoryposition, &bibnumber, &name, &raceid, &raceresultid, &year, &month, &day, &agecat)
		if err != nil {
			log.Fatal(err)
		}

		xx := RaceResult{
			ID:                  raceresultid,
			Time:                time,
			Position:            position,
			SexPosition:         sexposition,
			AgeCategoryPosition: agecategoryposition,
			RaceID:              raceid,
			RacerID:             r.ID,
			BibNumber:           bibnumber,
			AgeCategoryID:       agecat,
		}

		results = append(results, xx)

		races = append(races, Race{
			ID:    raceid,
			Name:  name,
			Year:  year,
			Month: month,
			Day:   day,
		})
	}

	return results, racers, races, nil
}
