package database

import (
	"github.com/chiefwhitecloud/running-man/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
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
	ID        int
	Position  int
	RaceID    int `sql:"index"`
	RacerID   int `sql:"index"`
	BibNumber string
	Racer     Racer
	Race      Race
}

type raceResultForTransform struct {
	pos   int
	first string
	last  string
}

func (db *Db) Migrate() {
	db.orm.AutoMigrate(&Racer{}, &Race{}, &RaceResult{})
}

func (db *Db) Create() {
	db.orm.CreateTable(&Racer{}, &Race{}, &RaceResult{})
}

func (db *Db) DropAllTables() {
	db.orm.DropTable(&Racer{}, &Race{}, &RaceResult{})
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

func (db *Db) SaveRace(r *model.RaceDetails) error {

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
				FirstName: mRacer.FirstName,
				LastName:  mRacer.LastName,
				Sex:       mRacer.Sex,
			}
			db.orm.Create(&racer)
		}

		result := RaceResult{
			RaceID:    race.ID,
			RacerID:   racer.ID,
			Position:  mRacer.Position,
			BibNumber: mRacer.BibNumber,
		}

		db.orm.Create(&result)

	}

	return nil
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

	rows, err := db.orm.Table("race_result").Select("race_result.position, race_result.bib_number, racer.first_name, racer.last_name, racer.id, race_result.id, racer.sex").Joins("join racer on race_result.racer_id = racer.id").Where("race_result.race_id = ?", r.ID).Rows()

	if err != nil {
		log.Println(err)
	}

	var (
		position     int
		bibnumber    string
		firstname    string
		lastname     string
		racerid      int
		raceresultid int
		sex          string
	)

	var results []RaceResult
	var races []Race
	var racers []Racer
	races = append(races, r)

	for rows.Next() {
		err := rows.Scan(&position, &bibnumber, &firstname, &lastname, &racerid, &raceresultid, &sex)
		if err != nil {
			log.Fatal(err)
		}

		xx := RaceResult{
			ID:        raceresultid,
			Position:  position,
			RaceID:    r.ID,
			RacerID:   racerid,
			BibNumber: bibnumber,
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

	rows, err := db.orm.Table("race_result").Select("race_result.position, race_result.bib_number, race.name,  race.id, race_result.id, race.year, race.month, race.day").Joins("join race on race_result.race_id = race.id").Where("race_result.racer_id = ?", r.ID).Rows()

	if err != nil {
		log.Println(err)
	}

	var (
		position     int
		bibnumber    string
		name         string
		raceid       int
		raceresultid int
		year         int
		month        int
		day          int
	)

	var results []RaceResult
	var races []Race
	var racers []Racer

	racers = append(racers, r)

	for rows.Next() {
		err := rows.Scan(&position, &bibnumber, &name, &raceid, &raceresultid, &year, &month, &day)
		if err != nil {
			log.Fatal(err)
		}

		xx := RaceResult{
			ID:        raceresultid,
			Position:  position,
			RaceID:    raceid,
			RacerID:   r.ID,
			BibNumber: bibnumber,
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
