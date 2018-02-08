package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import for sqlite db driver
	qGen "github.com/nboughton/go-sqgenlite"
)

var (
	sqlSchema = "CREATE TABLE IF NOT EXISTS results (id INTEGER PRIMARY KEY AUTOINCREMENT, date DATETIME, bset INT, bmac TEXT,	ball1 INT, ball2 INT, ball3 INT, ball4 INT, ball5 INT, ball6 INT, bonus INT)"
	allFields = []string{"date", "bset", "bmac", "ball1", "ball2", "ball3", "ball4", "ball5", "ball6", "bonus"}
	fmtSqlite = "2006-01-02 15:04:05-07:00"
)

// Exported constants
const (
	MAXBALLVAL = 59
	BALLS      = 7
)

// AppDB is a wrapper for *sql.DB so I can extend it by adding my own methods
type AppDB struct {
	*sql.DB
}

// Record wraps a single record from the database
type Record struct {
	Date    time.Time
	Machine string
	Set     int
	Ball    []int
}

// Connect returns a DB connection wrapper
func Connect(path string) *AppDB {
	// Connect to the database
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	// Create DB schema if it doesn't exist
	if _, err := db.Exec(sqlSchema); err != nil {
		log.Fatal(err)
	}

	return &AppDB{db}
}

// Update scrapes the archive site and adds newer records until
// an existing record is found.
func (db *AppDB) Update() error {
	q := qGen.NewQuery().
		Insert("results", allFields)
	stmt, err := db.Prepare(q.SQL)
	if err != nil {
		return err
	}

	for rec := range Scrape() {
		if db.Exists(rec.Date) {
			return fmt.Errorf("%+v exists, stopping update", rec)
		}

		if _, err := stmt.Exec(rec.Date, rec.Set, rec.Machine, rec.Ball[0], rec.Ball[1], rec.Ball[2], rec.Ball[3], rec.Ball[4], rec.Ball[5], rec.Ball[6]); err != nil {
			return err
		}
		log.Printf("Inserted: %+v \n", rec)
	}

	return nil
}

// Exists returns true if a record with t timestamp exists
func (db *AppDB) Exists(t time.Time) bool {
	if _, err := db.GetRecord(t); err != nil {
		return false
	}

	return true
}

// GetRecord retrieves a single record
func (db *AppDB) GetRecord(t time.Time) (Record, error) {
	q := qGen.NewQuery().
		Select("results", allFields...).
		Where("date = ?", t.Format(fmtSqlite))

	stmt, err := db.Prepare(q.SQL)
	if err != nil {
		return Record{}, err
	}

	var rec Record
	rec.Ball = make([]int, BALLS)
	return rec, stmt.QueryRow(q.Args...).Scan(&rec.Date, &rec.Set, &rec.Machine, &rec.Ball[0], &rec.Ball[1], &rec.Ball[2], &rec.Ball[3], &rec.Ball[4], &rec.Ball[5], &rec.Ball[6])
}

/*
// Results retrieves all records matching parameters in p
func (db *AppDB) Results(p QueryParams) <-chan Record {
	c := make(chan Record)

	go func() {
		q := qGen.NewQuery().Select("results", "date", "ball_machine", "ball_set", "ball1", "ball2", "ball3", "ball4", "ball5", "ball6", "bonus")

		applyFilters(q, p)
		stmt, _ := db.Prepare(q.Order("DATE(date)").SQL)
		rows, err := stmt.Query(q.Args...)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			var r Record
			r.Ball = make([]int, BALLS)
			if err := rows.Scan(&r.Date, &r.Machine, &r.Set, &r.Ball[0], &r.Ball[1], &r.Ball[2], &r.Ball[3], &r.Ball[4], &r.Ball[5], &r.Ball[6]); err != nil {
				log.Println(err)
			}

			c <- r
		}

		close(c)
	}()

	return c
}

// LastDraw retrieves the most recent set of results
func (db *AppDB) LastDraw() ([]int, error) {
	r, q := make([]int, BALLS), qGen.NewQuery().
		Select("results", "ball1", "ball2", "ball3", "ball4", "ball5", "ball6", "bonus").
		Order("DATE(date)").
		Append("DESC LIMIT 1")
	stmt, _ := db.Prepare(q.SQL)

	err := stmt.QueryRow().Scan(&r[0], &r[1], &r[2], &r[3], &r[4], &r[5], &r[6])

	return r, err
}

// Machines retrieves the available machines for the query parameters, constrained by date
func (db *AppDB) Machines(p QueryParams) ([]string, error) {
	r, q := []string{}, qGen.NewQuery().Select("results", "DISTINCT(ball_machine)")

	p.Machine = "all" // Ensure QueryParams are right for this
	applyFilters(q, p)
	stmt, _ := db.Prepare(q.Order("ball_machine").SQL)
	rows, err := stmt.Query(q.Args...)
	if err != nil {
		return r, err
	}
	defer rows.Close()

	for rows.Next() {
		var m string
		rows.Scan(&m)
		r = append(r, m)
	}

	return r, nil
}

// Sets retrieves the available ball sets for the query parameters, constrained by date
func (db *AppDB) Sets(p QueryParams) ([]int, error) {
	r, q := []int{}, qGen.NewQuery().Select("results", "DISTINCT(ball_set)")

	p.Set = 0 // Ensure QueryParams are right for this
	applyFilters(q, p)
	stmt, _ := db.Prepare(q.Order("ball_set").SQL)
	rows, err := stmt.Query(q.Args...)
	if err != nil {
		return r, err
	}
	defer rows.Close()

	for rows.Next() {
		var s int
		rows.Scan(&s)
		r = append(r, s)
	}

	return r, nil
}

// RowCount retrieves a count(*) of all rows in the db
func (db *AppDB) RowCount() (c int, err error) {
	err = db.QueryRow(qGen.NewQuery().Select("results", "COUNT(*)").SQL).Scan(&c)
	return c, err
}

// DataRange retrieves the first and last record dates
func (db *AppDB) DataRange() (time.Time, time.Time, error) {
	var first, last string
	q := qGen.NewQuery().Select("results", "MIN(date)", "MAX(date)")
	stmt, _ := db.Prepare(q.SQL)

	if err := stmt.QueryRow().Scan(&first, &last); err != nil {
		return time.Now(), time.Now(), err
	}

	f, _ := time.Parse(fmtSqlite, first)
	l, _ := time.Parse(fmtSqlite, last)
	return f, l, nil
}

// Update updates the database with the 4 most recent records
func (db *AppDB) Update() error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare queries
	qSel := qGen.NewQuery().Select("results", "COUNT(*)").Where("date = ?")
	sel, err := tx.Prepare(qSel.SQL)
	if err != nil {
		return err
	}

	qIns := qGen.NewQuery().Insert("results", []string{"date", "ball_set", "ball_machine", "ball1", "ball2", "ball3", "ball4", "ball5", "ball6", "bonus"})
	ins, err := tx.Prepare(qIns.SQL)
	if err != nil {
		return err
	}

	// Iterate scrape data
	for d := range ScrapeMostRecent() {
		// Check I don't already have this record
		var i int

		// Set args for queries
		qSel.Args = []interface{}{d.Date.Format(fmtSqlite)}
		qIns.Args = []interface{}{d.Date, d.Set, d.Machine, d.Ball[0], d.Ball[1], d.Ball[2], d.Ball[3], d.Ball[4], d.Ball[5], d.Ball[6]}

		if err := sel.QueryRow(qSel.Args...).Scan(&i); err != nil {
			tx.Rollback()
			return err
		}
		if i != 0 {
			log.Println(d.Date.String(), "already in DB, skipping")
			continue
		}

		// Insert new record
		if _, err := ins.Exec(qIns.Args...); err != nil {
			tx.Rollback()
			return err
		}
		log.Println(d, " inserted")
	}
	tx.Commit()

	return nil
}

// Populate retrieves and inserts the entire archive.
func (db *AppDB) Populate() error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare statement
	qIns := qGen.NewQuery().Insert("results", []string{"date", "ball_set", "ball_machine", "ball1", "ball2", "ball3", "ball4", "ball5", "ball6", "bonus"})
	insert, err := tx.Prepare(qIns.SQL)
	if err != nil {
		return err
	}

	// Iterate scrape data
	for d := range ScrapeFullArchive() {
		// Set args
		qIns.Args = []interface{}{d.Date, d.Set, d.Machine, d.Ball[0], d.Ball[1], d.Ball[2], d.Ball[3], d.Ball[4], d.Ball[5], d.Ball[6]}

		// Exec
		if _, err := insert.Exec(qIns.Args...); err != nil {
			tx.Rollback()
			return err
		}
		log.Println(d)
	}
	tx.Commit()

	return nil
}
*/
