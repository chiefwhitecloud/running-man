package database

import (
	_ "github.com/chiefwhitecloud/running-man/Godeps/_workspace/src/github.com/go-sql-driver/mysql"
	"github.com/chiefwhitecloud/running-man/Godeps/_workspace/src/github.com/jinzhu/gorm"
	"github.com/chiefwhitecloud/running-man/model"
	"log"
	"time"
)

type Db struct {
	orm              gorm.DB
	ConnectionString string
}

type Racer struct {
	ID           int
	FirstName    string
	LastName     string
	Sex          string
	LowBirthday  time.Time
	HighBirthday time.Time
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

type raceResultForTransform struct {
	pos   int
	first string
	last  string
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

	//save the race results information
	for i := range r.Racers {

		mRacer := r.Racers[i]

		var racer Racer

		if (db.orm.Where(&Racer{FirstName: mRacer.FirstName, LastName: mRacer.LastName}).First(&racer).RecordNotFound()) {
			racer = Racer{
				FirstName:    mRacer.FirstName,
				LastName:     mRacer.LastName,
				Sex:          mRacer.Sex,
				LowBirthday:  mRacer.LowBirthdayDate,
				HighBirthday: mRacer.HighBirthdayDate,
			}
			db.orm.Create(&racer)
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
