package database

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chiefwhitecloud/running-man/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Db struct {
	orm              gorm.DB
	ConnectionString string
}

type ImportTask struct {
	ID        int
	RaceID    int
	Status    string
	SrcUrl    string
	ErrorText string
}

type Racer struct {
	ID      int
	Created time.Time
}

type RaceGroup struct {
	ID           int
	Name         string
	Distance     string `gorm:"size:10"`
	DistanceUnit string `gorm:"size:1"`
	ETag         string
	LastUpdated  time.Time
}

type Race struct {
	ID           int
	Name         string
	Date         time.Time
	RaceGroupID  int `sql:"index"`
	ImportStatus string
	SrcUrl       string
	ETag         string
	LastUpdated  time.Time
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
	ChipTime            string
	Racer               Racer
	Race                Race
	Sex                 string
	Club                string
}

type AgeCategory struct {
	ID   int
	Name string
}

type AgeLookup struct {
	minAge int
	maxAge int
}

type AgeResult struct {
	RaceDate    time.Time
	AgeCategory string
}

// ErrRecordNotFoundError is an error implementation that includes the table name
var ErrRecordNotFoundError = errors.New("Record not found")
var ErrNoRecordsAvailable = errors.New("No records available")

func (db *Db) Migrate() {
	db.orm.AutoMigrate(&Racer{}, &Race{}, &RaceResult{}, &AgeCategory{}, &ImportTask{}, &RaceGroup{})

	cats := []string{
		"U20", "-19", "<20",
		"20-24", "25-29", "20-29",
		"30-34", "35-39", "30-39",
		"40-44", "45-49", "40-49",
		"50-54", "55-59", "50-59",
		"60-64", "65-69", "60-69",
		"70-74", "75-79", "70-79",
		"70+", "80-84", "85-89",
		"80-89", "80+", "A", "NOAGE",
	}

	for i := 0; i < len(cats); i++ {
		cat := AgeCategory{Name: cats[i]}
		db.orm.Create(&cat)
	}

}

func (db *Db) Create() {
	db.orm.CreateTable(&Racer{}, &Race{}, &RaceResult{}, &AgeCategory{}, &ImportTask{}, &RaceGroup{})
}

func (db *Db) DropAllTables() {
	db.orm.DropTable(&Racer{}, &Race{}, &RaceResult{}, &AgeCategory{}, &ImportTask{}, &RaceGroup{})
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

//CreateImportTask creates an import task and returns the new task
func (db *Db) CreateImportTask(url string) (ImportTask, error) {
	race := Race{Name: "Pending", ImportStatus: "pending", SrcUrl: url, Date: time.Now(), LastUpdated: time.Now()}

	if err := db.orm.Create(&race).Error; err != nil {
		return ImportTask{}, err
	}

	task := ImportTask{RaceID: race.ID, Status: "pending", SrcUrl: url}

	if err := db.orm.Create(&task).Error; err != nil {
		return task, err
	}

	return task, nil
}

//FailedImport cleans up after a failed import
func (db *Db) FailedImport(task ImportTask, err error) {
	task.Status = "failed"
	task.ErrorText = err.Error()
	if err := db.orm.Save(&task).Error; err != nil {

	}

	db.orm.Delete(&Race{}, task.RaceID)

	db.orm.Delete(&RaceResult{}, "race_id = ?", task.RaceID)

}

//CreateRaceGroup creates a new race group and returns the race group
func (db *Db) CreateRaceGroup(name string, distance string, distanceunit string) (RaceGroup, error) {
	etag, lastUpdated := db.CreateEtagAndLastUpdated(name)
	raceGroup := RaceGroup{Name: name, Distance: distance, DistanceUnit: distanceunit, LastUpdated: lastUpdated, ETag: etag}
	if err := db.orm.Save(&raceGroup).Error; err != nil {
		return raceGroup, err
	}
	return raceGroup, nil
}

func (db *Db) UpdateRaceGroup(id int, name string, distance string, distanceunit string) (RaceGroup, error) {
	raceGroup, _ := db.GetRaceGroup(id)
	etag, lastUpdated := db.CreateEtagAndLastUpdated(name)
	raceGroup.Name = name
	raceGroup.Distance = distance
	raceGroup.DistanceUnit = distanceunit
	raceGroup.LastUpdated = lastUpdated
	raceGroup.ETag = etag
	db.orm.Save(&raceGroup)
	return raceGroup, nil
}

func (db *Db) CreateEtagAndLastUpdated(name string) (string, time.Time) {
	t := time.Now()
	h := sha1.New()
	h.Write([]byte(name + t.String()))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs), t
}

func (db *Db) DeleteRaceGroup(id int) (RaceGroup, error) {
	raceGroup := RaceGroup{}
	if db.orm.First(&raceGroup, id).RecordNotFound() {
		return raceGroup, ErrRecordNotFoundError
	}

	races, _ := db.GetRacesForRaceGroup(raceGroup.ID)

	for i, _ := range races {
		etag, lastUpdated := db.CreateEtagAndLastUpdated(races[i].Name)
		races[i].LastUpdated = lastUpdated
		races[i].ETag = etag
		db.orm.Save(&races[i])
	}

	if err := db.orm.Delete(&raceGroup).Error; err != nil {
		return raceGroup, err
	}

	//update the etag for the newest item... this is the etag used to the list
	if raceGroupLastUpdated, err := db.GetLastUpdatedRaceGroup(); err == nil {
		etag, lastUpdated := db.CreateEtagAndLastUpdated(raceGroupLastUpdated.Name)
		raceGroupLastUpdated.LastUpdated = lastUpdated
		raceGroupLastUpdated.ETag = etag
		db.orm.Save(&raceGroupLastUpdated)
	}

	return raceGroup, nil
}

func (db *Db) GetRaceGroup(id int) (RaceGroup, error) {
	raceGroup := RaceGroup{}
	if err := db.orm.First(&raceGroup, id).Error; err != nil {
		return raceGroup, err
	}
	return raceGroup, nil
}

func (db *Db) GetPendingImportTasks() []ImportTask {
	tasks := []ImportTask{}
	db.orm.Where("status = ?", "pending").Find(&tasks)
	return tasks
}

func (db *Db) HasRaceBeenImported(url string) bool {
	races := []Race{}
	db.orm.Where("src_url = ? AND import_status = ?", url, "completed").Find(&races)
	if len(races) > 0 {
		return true
	} else {
		return false
	}
}

func (db *Db) SaveRace(task ImportTask, r *model.RaceDetails) (Race, error) {

	cats := []AgeCategory{}

	db.orm.Find(&cats)

	raceDate := time.Date(r.Year, time.Month(r.Month), r.Day, 0, 0, 0, 0, time.UTC)

	race := Race{ID: task.RaceID}
	db.orm.First(&race)
	race.Name = r.Name
	race.Date = raceDate
	db.orm.Save(&race)

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
				rows, _ := db.orm.Raw("SELECT race_result.racer_id FROM race_result WHERE race_result.name = ? AND race_result.race_id = ?", mRacer.Name, race.ID).Rows()
				defer rows.Close()

				mustBeNewRacer := false

				for rows.Next() {
					var uid int
					_ = rows.Scan(&uid)
					mustBeNewRacer = true
				}

				if mustBeNewRacer {
					break
				} else {
					//look at the racers age catgory history... does it look like a match?
					early, late, _ := db.GetRacerBirthDates(raceResults[i].RacerID)
					minAge, maxAge, _ := db.GetAgeRangeOnDate(early, late, raceDate)

					//check to see if the race is within the same age category
					ok, err := db.isAgeRangeWithinCatgory(maxAge, minAge, mRacer.AgeCategory)

					if err != nil {
						return race, err
					}
					if ok {
						//existing racer is found
						db.orm.Where(&Racer{ID: raceResults[i].RacerID}).Find(&racer)
						break
					}
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
		if len(mRacer.Club) > 0 {
			result.Club = mRacer.Club
		}

		if len(mRacer.ChipTime) > 0 {
			result.ChipTime = mRacer.ChipTime
		}

		db.orm.Create(&result)

	}

	t := time.Now()

	h := sha1.New()
	h.Write([]byte(race.Name + race.Date.String() + t.String()))
	bs := h.Sum(nil)

	race.ImportStatus = "completed"
	race.LastUpdated = time.Now()
	race.ETag = hex.EncodeToString(bs)
	db.orm.Save(&race)

	task.Status = "completed"
	db.orm.Save(&task)

	return race, nil
}

func (db *Db) GetLastUpdatedRace() (Race, error) {
	race := Race{}
	db.orm.Order("last_updated desc").First(&race)

	if race.Name == "" {
		return race, ErrNoRecordsAvailable
	}

	return race, nil
}

func (db *Db) GetLastUpdatedRaceGroup() (RaceGroup, error) {
	raceGroup := RaceGroup{}

	db.orm.Order("last_updated desc").First(&raceGroup)

	if raceGroup.Name == "" {
		return raceGroup, errors.New("No race groups found")
	}

	return raceGroup, nil
}

func (db *Db) GetRaceGroups() ([]RaceGroup, error) {
	raceGroups := []RaceGroup{}
	db.orm.Order("name asc").Find(&raceGroups)
	return raceGroups, nil
}

func (db *Db) GetRacesForRaceGroup(raceGroupId int) ([]Race, error) {
	races := []Race{}
	db.orm.Where("race_group_id = ?", raceGroupId).Find(&races)
	return races, nil
}

func (db *Db) GetRaces() ([]Race, error) {
	races := []Race{}
	db.orm.Order("date desc").Find(&races)
	return races, nil
}

//GetRace
func (db *Db) GetRace(id int) (Race, error) {
	race := Race{}
	if db.orm.First(&race, id).RecordNotFound() {
		return race, ErrRecordNotFoundError
	}
	return race, nil
}

//DeleteRace
func (db *Db) DeleteRace(id int) (Race, error) {
	race := Race{}
	if db.orm.First(&race, id).RecordNotFound() {
		return race, ErrRecordNotFoundError
	}

	if err := db.orm.Delete(&race).Error; err != nil {
		return race, err
	}

	//update the etag for the newest item... this is the etag used to the list
	if raceLastUpdated, err := db.GetLastUpdatedRace(); err == nil {
		etag, lastUpdated := db.CreateEtagAndLastUpdated(raceLastUpdated.Name)
		raceLastUpdated.LastUpdated = lastUpdated
		raceLastUpdated.ETag = etag
		db.orm.Save(&raceLastUpdated)
	}

	return race, nil
}

//GetImportTask
func (db *Db) GetImportTask(id int) (ImportTask, error) {
	task := ImportTask{}
	if db.orm.First(&task, id).RecordNotFound() {
		return task, ErrRecordNotFoundError
	}
	return task, nil
}

func (db *Db) GetRacer(id int) (Racer, error) {
	racer := Racer{}
	if db.orm.First(&racer, id).RecordNotFound() {
		return racer, ErrRecordNotFoundError
	}
	return racer, nil
}

func (db *Db) MergeRacers(parentRacer Racer, racer Racer) (Racer, error) {
	//update all race results with the new id
	db.orm.Exec("UPDATE race_result SET racer_id=? WHERE racer_id =?", parentRacer.ID, racer.ID)
	return parentRacer, nil
}

func (db *Db) AddRaceToRaceGroup(raceGroup RaceGroup, race Race) (RaceGroup, error) {

	etag, lastUpdated := db.CreateEtagAndLastUpdated(race.Name)
	race.LastUpdated = lastUpdated
	race.ETag = etag
	race.RaceGroupID = raceGroup.ID
	db.orm.Save(&race)

	etag, lastUpdated = db.CreateEtagAndLastUpdated(raceGroup.Name)
	raceGroup.LastUpdated = lastUpdated
	raceGroup.ETag = etag
	db.orm.Save(&raceGroup)

	return raceGroup, nil
}

func (db *Db) isAgeRangeWithinCatgory(minAge int, maxAge int, ageCategory string) (bool, error) {
	catMinAge, catMaxAge, err := db.GetMinMaxAgeForCategory(ageCategory)

	if err != nil {
		return false, err
	} else {
		return minAge >= catMinAge && maxAge <= catMaxAge, nil
	}

}

func (db *Db) GetMinMaxAgeForCategory(ageCategory string) (int, int, error) {
	ageCategoryMap := map[string]*AgeLookup{
		"U20":   &AgeLookup{5, 19},
		"-19":   &AgeLookup{5, 19},
		"<20":   &AgeLookup{5, 19},
		"20-24": &AgeLookup{20, 24},
		"25-29": &AgeLookup{25, 29},
		"20-29": &AgeLookup{20, 29},
		"30-34": &AgeLookup{30, 34},
		"35-39": &AgeLookup{35, 39},
		"30-39": &AgeLookup{30, 39},
		"40-44": &AgeLookup{40, 44},
		"45-49": &AgeLookup{45, 49},
		"40-49": &AgeLookup{40, 49},
		"50-54": &AgeLookup{50, 54},
		"55-59": &AgeLookup{55, 59},
		"50-59": &AgeLookup{50, 59},
		"60-64": &AgeLookup{60, 64},
		"65-69": &AgeLookup{65, 69},
		"60-69": &AgeLookup{60, 69},
		"70-74": &AgeLookup{70, 74},
		"75-79": &AgeLookup{75, 79},
		"70-79": &AgeLookup{70, 79},
		"70+":   &AgeLookup{70, 100},
		"80-84": &AgeLookup{80, 84},
		"85-89": &AgeLookup{85, 89},
		"80-89": &AgeLookup{80, 89},
		"80+":   &AgeLookup{80, 100},
		"A":     &AgeLookup{5, 100},
		"NOAGE": &AgeLookup{5, 100},
	}
	if age, ok := ageCategoryMap[ageCategory]; ok {
		return age.minAge, age.maxAge, nil
	} else {
		return 0, 0, errors.New("Failed to find age category " + ageCategory)
	}
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

func (db *Db) GetRacerNames(id int) ([]string, error) {

	rows, err := db.orm.Raw(fmt.Sprintf("SELECT race_result.name FROM race_result LEFT JOIN (race) ON (race.id = race_result.race_id) WHERE race_result.racer_id = %d GROUP BY race_result.name", id)).Rows()

	if err != nil {
		log.Println(err)
	}

	var name string

	var results []string

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, name)
	}

	return results, nil
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

	defer rows.Close()

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

func (db *Db) GetRaceResultsForRace(raceid int, startPosition int, numOfRecords int) ([]RaceResult, []Racer, []Race, error) {

	// XXX: Maybe a better way to do this using the ORM.  Couldn't figure it out.
	// For now doing a manual join and populating the struct to return.  Seems like the
	// ORM should be doing some more work here.

	r := Race{}

	db.orm.Find(&r, raceid)

	rows, err := db.orm.Table("race_result").
		Select("race_result.time, race_result.position, race_result.sex_position, race_result.age_category_position, race_result.bib_number, race_result.name, racer.id, race_result.id, race_result.sex, race_result.age_category_id, race_result.club, race_result.chip_time").
		Joins("join racer on race_result.racer_id = racer.id").
		Where("race_result.race_id = ?", r.ID).
		Order("race_result.position ASC").
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
		club                string
		chiptime            string
	)

	var results []RaceResult
	var races []Race
	var racers []Racer
	var numOfRows = 0
	races = append(races, r)

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&time, &position, &sexposition, &agecategoryposition, &bibnumber, &racername, &racerid, &raceresultid, &sex, &agecat, &club, &chiptime)
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
			Club:                club,
			ChipTime:            chiptime,
		}

		if startPosition > 0 {

			if startPosition <= xx.Position {
				results = append(results, xx)

				racers = append(racers, Racer{
					ID: racerid,
				})

				numOfRows++
			}

		} else {
			results = append(results, xx)

			racers = append(racers, Racer{
				ID: racerid,
			})

			numOfRows++
		}

		if numOfRows == numOfRecords {
			break
		}
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
		Select("race_result.time, race_result.position, race_result.sex_position, race_result.age_category_position, race_result.bib_number, race_result.name, race.name,  race.id, race.race_group_id, race_result.id,  race_result.sex, race.date, race_result.age_category_id").
		Joins("join race on race_result.race_id = race.id").
		Where("race_result.racer_id = ?", r.ID).
		Order("race.date DESC").
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
		sex                 string
		raceGroupId         int
	)

	var results []RaceResult
	var races []Race
	var racers []Racer

	racers = append(racers, r)

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&raceresulttime, &position, &sexposition, &agecategoryposition, &bibnumber, &racername, &racename, &raceid, &raceGroupId, &raceresultid, &sex, &raceDate, &agecat)
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
			Sex:                 sex,
		}

		results = append(results, xx)

		races = append(races, Race{
			ID:          raceid,
			Name:        racename,
			Date:        raceDate,
			RaceGroupID: raceGroupId,
		})
	}

	return results, racers, races, nil
}
