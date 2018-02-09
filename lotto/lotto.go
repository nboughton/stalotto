package lotto

import (
	"fmt"
	"time"
)

// Result represents a single Lotto draw result
type Result struct {
	Date    time.Time
	Machine string
	Set     int
	Ball    []int
}

// Exported constants
const (
	MAXBALLVAL = 59
	BALLS      = 7
)

// NewResult sets up a new Result struct for use
func NewResult() Result {
	var res Result
	res.Ball = make([]int, BALLS)
	return res
}

// String satisfies the Stringer interface for Result
func (r Result) String() string {
	return fmt.Sprintf("%s %s:%d %d", r.Date.Format("2006-01-02"), r.Machine, r.Set, r.Ball)
}
