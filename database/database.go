package database

import (
	"fmt"
	"github.com/chiefwhitecloud/running-man/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"time"
)

type Db struct {
	orm              gorm.DB
	ConnectionString string
}

type Racer struct {
	ID      int
	Created time.Time
}

type Race struct {
	ID   int
	Name string
	Date time.Time
}

type RaceResult struct {
	ID                  int
	Name                string
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
	Sex                 string
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

	raceDate := time.Date(r.Year, time.Month(r.Month), r.Day, 0, 0, 0, 0, time.UTC)

	race := Race{Name: r.Name, Date: raceDate}

	db.orm.Create(&race)

	//save the race results information
	for i := range r.Racers {

		mRacer := r.Racers[i]

		var raceResults []RaceResult
		var racer Racer

		db.orm.Where(&RaceResult{Name: mRacer.Name}).Find(&raceResults)

		//find the agecategory id for the current race result
		catId := 0
		for i := range cats {
			if cats[i].Name == mRacer.AgeCategory {
				catId = cats[i].ID
			}
		}

		if len(raceResults) == 0 {
			//must be a new racer.. no racer with that name
			racer = Racer{Created: time.Now()}
			db.orm.Create(&racer)
		} else if len(raceResults) > 0 {
			//We have some Racer records with the same name, etc... Time to match the race result with an existing Racer in the database.

			for i := range raceResults {
				//did we already save a racer with same name and age group to this race?

				//a runnner with same name already ran this race.
				// FIX ME:  Needs to check their aliases too
				rows, _ := db.orm.Raw("SELECT race_result.racer_id FROM race_result WHERE race_result.name = ? AND race_result.race_id = ? AND race_result.age_category_id = ?", mRacer.Name, race.ID, catId).Rows()
				defer rows.Close()

				found := false

				for rows.Next() {
					var uid int
					_ = rows.Scan(&uid)
					found = true
				}

				//look at the racers age catgory history... does it look like a match?
				early, late, _ := db.GetRacerBirthDates(raceResults[i].RacerID)
				minAge, maxAge, _ := db.GetAgeRangeOnDate(early, late, raceDate)

				//check to see if the race is within the same age category
				if db.isAgeRangeWithinCatgory(maxAge, minAge, mRacer.AgeCategory) && !found {
					//existing racer is found
					db.orm.Where(&Racer{ID: raceResults[i].RacerID}).Find(&racer)
					break
				}
			}

			//if no match found... create a new Racer
			if racer == (Racer{}) {
				racer = Racer{Created: time.Now()}
				db.orm.Create(&racer)
			}

		}

		result := RaceResult{
			RaceID:              race.ID,
			RacerID:             racer.ID,
			Name:                mRacer.Name,
			Position:            mRacer.Position,
			BibNumber:           mRacer.BibNumber,
			SexPosition:         mRacer.SexPosition,
			AgeCategoryPosition: mRacer.AgeCategoryPosition,
			AgeCategoryID:       catId,
			Time:                mRacer.Time,
			Sex:                 mRacer.Sex,
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

	rows, err := db.orm.Raw(fmt.Sprintf("SELECT race.date, age_category.name FROM race LEFT JOIN (race_result, age_category) ON (race_result.race_id = race.id AND age_category.id = race_result.age_category_id) WHERE race_result.racer_id = %d ORDER BY race.date ASC", id)).Rows()

	if err != nil {
		log.Println(err)
	}

	var (
		raceDate        time.Time
		ageCategoryName string
	)

	var results []AgeResult

	for rows.Next() {
		err := rows.Scan(&raceDate, &ageCategoryName)
		if err != nil {
			log.Fatal(err)
		}
		xx := AgeResult{
			RaceDate:    raceDate,
			AgeCategory: ageCategoryName,
		}

		results = append(results, xx)
	}

	var high time.Time
	var low time.Time
	var ageCat string
	var lastRaceDate time.Time

	for i := range results {

		if ageCat == "" && lastRaceDate.IsZero() {
			ageCat = results[i].AgeCategory
			lastRaceDate = results[i].RaceDate
			low, high, _ = db.GetBirthDateRangeForCategory(results[i].RaceDate, results[i].AgeCategory)
		} else {
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
		Select("race_result.time, race_result.position, race_result.sex_position, race_result.age_category_position, race_result.bib_number, race_result.name, racer.id, race_result.id, race_result.sex, race_result.age_category_id").
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
		racerid             int
		raceresultid        int
		sex                 string
		agecat              int
		racername           string
	)

	var results []RaceResult
	var races []Race
	var racers []Racer
	races = append(races, r)

	for rows.Next() {
		err := rows.Scan(&time, &position, &sexposition, &agecategoryposition, &bibnumber, &racername, &racerid, &raceresultid, &sex, &agecat)
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
			Name:                racername,
			Sex:                 sex,
		}

		results = append(results, xx)

		racers = append(racers, Racer{
			ID: racerid,
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
		Select("race_result.time, race_result.position, race_result.sex_position, race_result.age_category_position, race_result.bib_number, race_result.name, race.name,  race.id, race_result.id, race.date, race_result.age_category_id").
		Joins("join race on race_result.race_id = race.id").
		Where("race_result.racer_id = ?", r.ID).
		Rows()

	if err != nil {
		log.Println(err)
	}

	var (
		raceresulttime      string
		position            int
		sexposition         int
		agecategoryposition int
		bibnumber           string
		racename            string
		racername           string
		raceid              int
		raceresultid        int
		agecat              int
		raceDate            time.Time
	)

	var results []RaceResult
	var races []Race
	var racers []Racer

	racers = append(racers, r)

	for rows.Next() {
		err := rows.Scan(&raceresulttime, &position, &sexposition, &agecategoryposition, &bibnumber, &racername, &racename, &raceid, &raceresultid, &raceDate, &agecat)
		if err != nil {
			log.Fatal(err)
		}

		xx := RaceResult{
			ID:                  raceresultid,
			Time:                raceresulttime,
			Position:            position,
			SexPosition:         sexposition,
			AgeCategoryPosition: agecategoryposition,
			RaceID:              raceid,
			RacerID:             r.ID,
			BibNumber:           bibnumber,
			AgeCategoryID:       agecat,
			Name:                racername,
		}

		results = append(results, xx)

		races = append(races, Race{
			ID:   raceid,
			Name: racename,
			Date: raceDate,
		})
	}

	return results, racers, races, nil
}
