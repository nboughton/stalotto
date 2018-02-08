package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
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

// NewRecord sets up a new Record struct for use
func NewRecord() Record {
	var rec Record
	rec.Ball = make([]int, BALLS)
	return rec
}

// String satisfies the Stringer interface for Records
func (r Record) String() string {
	return fmt.Sprintf("%s %s:%d %d", r.Date.Format("2006-01-02"), r.Machine, r.Set, r.Ball)
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

	rec := NewRecord()
	return rec, stmt.QueryRow(q.Args...).Scan(&rec.Date, &rec.Set, &rec.Machine, &rec.Ball[0], &rec.Ball[1], &rec.Ball[2], &rec.Ball[3], &rec.Ball[4], &rec.Ball[5], &rec.Ball[6])
}

func groupOR(field string, vals int) string {
	slc := make([]string, vals)
	for i := range slc {
		slc[i] = fmt.Sprintf("%s = ?", field)
	}
	return "(" + strings.Join(slc, " OR ") + ")"
}

// GetRecords returns a channel of records
func (db *AppDB) GetRecords(begin, end time.Time, machines []string, sets []int) <-chan Record {
	c := make(chan Record)

	go func() {
		defer close(c)

		q := qGen.NewQuery().
			Select("results", allFields...).
			Where("date BETWEEN ? AND ?", begin.Format(fmtSqlite), end.Format(fmtSqlite))

		if len(machines) > 0 {
			q.Append(fmt.Sprintf("AND %s", groupOR("bmac", len(machines))))
			for _, m := range machines {
				q.Args = append(q.Args, m)
			}
		}

		if len(sets) > 0 {
			q.Append(fmt.Sprintf("AND %s", groupOR("bset", len(sets))))
			for _, s := range sets {
				q.Args = append(q.Args, s)
			}
		}

		stmt, err := db.Prepare(q.SQL)
		if err != nil {
			log.Println(err)
			return
		}

		rows, err := stmt.Query(q.Args...)
		if err != nil {
			log.Println(err)
			return
		}

		for rows.Next() {
			rec := NewRecord()
			if err := rows.Scan(&rec.Date, &rec.Set, &rec.Machine, &rec.Ball[0], &rec.Ball[1], &rec.Ball[2], &rec.Ball[3], &rec.Ball[4], &rec.Ball[5], &rec.Ball[6]); err != nil {
				log.Println(err)
				continue
			}

			c <- rec
		}
	}()

	return c
}
