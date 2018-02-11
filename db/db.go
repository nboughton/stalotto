package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import for sqlite db driver
	qGen "github.com/nboughton/go-sqgenlite"
	"github.com/nboughton/stalotto/lotto"
)

var (
	sqlPragmas = "PRAGMA journal_mode=WAL;	PRAGMA busy_timeout=5000"
	sqlSchema  = "CREATE TABLE IF NOT EXISTS results (id INTEGER PRIMARY KEY AUTOINCREMENT, date DATETIME, bset INT, bmac TEXT,	ball1 INT, ball2 INT, ball3 INT, ball4 INT, ball5 INT, ball6 INT, bonus INT)"
	allFields  = []string{"date", "bset", "bmac", "ball1", "ball2", "ball3", "ball4", "ball5", "ball6", "bonus"}
	fmtSqlite  = "2006-01-02 15:04:05-07:00"
)

// AppDB is a wrapper for *sql.DB so I can extend it by adding my own methods
type AppDB struct {
	*sql.DB
}

// Connect returns a DB connection wrapper
func Connect(path string) *AppDB {
	// I don't care where you want your database. I'm going to ensure that it's there
	dir, _ := filepath.Split(path)
	if err := os.MkdirAll(dir, 0770); err != nil {
		log.Fatal(err)
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	// Disable connection pooling
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	// Set PRAGMAs
	if _, err := db.Exec(sqlPragmas); err != nil {
		log.Fatal(err.Error())
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

	for res := range Scrape() {
		if db.Exists(res.Date) {
			return fmt.Errorf("update done")
		}

		if _, err := stmt.Exec(res.Date, res.Set, res.Machine, res.Balls[0], res.Balls[1], res.Balls[2], res.Balls[3], res.Balls[4], res.Balls[5], res.Bonus); err != nil {
			return err
		}
		log.Printf("Inserted: %+v \n", res)
	}

	return nil
}

// Exists returns true if a record with t timestamp exists
func (db *AppDB) Exists(t time.Time) bool {
	if _, err := db.Result(t); err != nil {
		return false
	}

	return true
}

// Result retrieves a single record
func (db *AppDB) Result(t time.Time) (lotto.Result, error) {
	q := qGen.NewQuery().
		Select("results", allFields...).
		Where("date = ?", t.Format(fmtSqlite))

	stmt, err := db.Prepare(q.SQL)
	if err != nil {
		return lotto.Result{}, err
	}

	res := lotto.NewResult()
	return res, stmt.QueryRow(q.Args...).Scan(&res.Date, &res.Set, &res.Machine, &res.Balls[0], &res.Balls[1], &res.Balls[2], &res.Balls[3], &res.Balls[4], &res.Balls[5], &res.Bonus)
}

func groupOR(field string, vals int) string {
	slc := make([]string, vals)
	for i := range slc {
		slc[i] = fmt.Sprintf("%s = ?", field)
	}
	return "(" + strings.Join(slc, " OR ") + ")"
}

// Results returns a channel of records
func (db *AppDB) Results(begin, end time.Time, machines []string, sets []int) <-chan lotto.Result {
	c := make(chan lotto.Result)

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

		q.Order("date").Append("DESC")

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
			res := lotto.NewResult()
			if err := rows.Scan(&res.Date, &res.Set, &res.Machine, &res.Balls[0], &res.Balls[1], &res.Balls[2], &res.Balls[3], &res.Balls[4], &res.Balls[5], &res.Bonus); err != nil {
				log.Println(err)
				continue
			}

			c <- res
		}
	}()

	return c
}

// Machines returns the distinct machine names constrained by date and sets
func (db *AppDB) Machines(begin time.Time, end time.Time, sets []int) ([]string, error) {
	q := qGen.NewQuery().Select("results", "DISTINCT(bmac)")
	if len(sets) > 0 {
		q.Append(fmt.Sprintf("AND %s", groupOR("bset", len(sets))))
		for _, s := range sets {
			q.Args = append(q.Args, s)
		}
	}
	q.Order("bmac")

	stmt, err := db.Prepare(q.SQL)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(q.Args...)
	if err != nil {
		return nil, err
	}

	r := []string{}
	for rows.Next() {
		m := ""
		rows.Scan(&m)
		r = append(r, m)
	}

	return r, nil
}

// Sets returns the distinct sets constrained by date and machines
func (db *AppDB) Sets(begin time.Time, end time.Time, machines []string) ([]int, error) {
	q := qGen.NewQuery().Select("results", "DISTINCT(bset)")
	if len(machines) > 0 {
		q.Append(fmt.Sprintf("AND %s", groupOR("bmac", len(machines))))
		for _, m := range machines {
			q.Args = append(q.Args, m)
		}
	}
	q.Order("bset")

	stmt, err := db.Prepare(q.SQL)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(q.Args...)
	if err != nil {
		return nil, err
	}

	r := []int{}
	for rows.Next() {
		s := 0
		rows.Scan(&s)
		r = append(r, s)
	}

	return r, nil
}

// LastDraw retrieves the most recent set of results
func (db *AppDB) LastDraw() (lotto.Result, error) {
	q := qGen.NewQuery().
		Select("results", allFields...).
		Order("date").
		Append("DESC LIMIT 1")

	res := lotto.Result{}
	stmt, err := db.Prepare(q.SQL)
	if err != nil {
		return res, err
	}

	return res, stmt.QueryRow().Scan(&res.Date, &res.Set, &res.Machine, &res.Balls[0], &res.Balls[1], &res.Balls[2], &res.Balls[3], &res.Balls[4], &res.Balls[5], &res.Bonus)
}

// DataRange retrieves the first and last record dates
func (db *AppDB) DataRange() (time.Time, time.Time, error) {
	q := qGen.NewQuery().Select("results", "MIN(date)", "MAX(date)")
	stmt, err := db.Prepare(q.SQL)
	if err != nil {
		return time.Now(), time.Now(), err
	}

	first, last := "", ""
	if err := stmt.QueryRow().Scan(&first, &last); err != nil {
		return time.Now(), time.Now(), err
	}

	f, _ := time.Parse(fmtSqlite, first)
	l, _ := time.Parse(fmtSqlite, last)
	return f, l, nil
}
