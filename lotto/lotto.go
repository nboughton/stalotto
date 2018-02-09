// Copyright © 2018 Nick Boughton <nicholasboughton@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package lotto

import (
	"fmt"
	"sort"
	//"sort"
	"time"
)

// Exported constants
const (
	MAXBALLVAL = 59
	BALLS      = 7
)

// Result represents a single Lotto draw result
type Result struct {
	Date    time.Time
	Machine string
	Set     int
	Ball    []int
}

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

// Set represents a collection of Results
type Set []Result

type drawn struct {
	ball      int
	frequency int
}

type byFrequency []drawn

func (f byFrequency) Len() int           { return len(f) }
func (f byFrequency) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f byFrequency) Less(i, j int) bool { return f[i].frequency < f[j].frequency }

// Prune off balls that have never been drawn
func (f byFrequency) Prune() byFrequency {
	out := byFrequency{}
	for _, b := range f {
		if b.frequency > 0 {
			out = append(out, b)
		}
	}
	return out
}

// MostDrawn returns the 6 most drawn balls and the most drawn Bonus number
// from the Set
func (s Set) MostDrawn() ([]int, int) {
	balls, bonus := s.frequencyData()

	sort.Sort(sort.Reverse(balls))
	sort.Sort(sort.Reverse(bonus))

	results := []int{}
	for i := 0; i < BALLS-1; i++ {
		results = append(results, balls[i].ball)
	}

	return results, bonus[0].ball
}

// LeastDrawn returns the 6 lesat drawn balls and the least drawn Bonus number
// from the Set
func (s Set) LeastDrawn() ([]int, int) {
	balls, bonus := s.frequencyData()

	sort.Sort(balls)
	sort.Sort(bonus)

	results := []int{}
	for i := 0; i < BALLS-1; i++ {
		results = append(results, balls[i].ball)
	}

	return results, bonus[0].ball
}

// frequencyData returns the pruned frequency sets for balls and bonus balls
func (s Set) frequencyData() (balls byFrequency, bonus byFrequency) {
	balls = make(byFrequency, MAXBALLVAL+1)
	bonus = make(byFrequency, MAXBALLVAL+1)

	for _, res := range s {
		for i, n := range res.Ball {
			if i == BALLS-1 {
				bonus[n].ball = n
				bonus[n].frequency++
				break // The bonus ball is always last
			}

			balls[n].ball = n
			balls[n].frequency++
		}
	}

	return balls.Prune(), bonus.Prune()
}
