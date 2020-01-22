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
	"math/rand"
	"sort"
	"time"
)

// Exported constants
const (
	MAXBALLVAL = 59
	BALLS      = 6
)

// Result represents a single Lotto draw result
type Result struct {
	Date    time.Time
	Machine string
	Set     int
	Balls   []int
	Bonus   int
}

// NewResult sets up a new Result struct for use
func NewResult() Result {
	var res Result
	res.Balls = make([]int, BALLS)
	return res
}

// String satisfies the Stringer interface for Result
func (r Result) String() string {
	return fmt.Sprintf("%s %s:%d %d %d", r.Date.Format("2006-01-02"), r.Machine, r.Set, r.Balls, r.Bonus)
}

// ResultSet represents a collection of Results
type ResultSet []Result

// ByDrawFrequency returns the frequency sets for balls and bonus balls
func (s ResultSet) ByDrawFrequency() (balls FrequencySet, bonus FrequencySet) {
	balls = make(FrequencySet, MAXBALLVAL+1)
	bonus = make(FrequencySet, MAXBALLVAL+1)

	for _, res := range s {
		for _, n := range res.Balls {
			balls[n].Ball = n
			balls[n].Frequency++
		}
		bonus[res.Bonus].Ball = res.Bonus
		bonus[res.Bonus].Frequency++
	}

	return balls[1:], bonus[1:]
}

// Drawn represents a record of a ball number and how often it has been drawn
type drawn struct {
	Ball      int
	Frequency int
}

// FrequencySet represents a collection of balls that can be ordered
// by draw frequency. FrequencySet also satisfies the Sort interface
type FrequencySet []drawn

// Len, Swap and Less satisfy the Sort interface for FrequencySet
func (f FrequencySet) Len() int           { return len(f) }
func (f FrequencySet) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f FrequencySet) Less(i, j int) bool { return f[i].Frequency < f[j].Frequency }

// Prune off balls that have never been drawn
func (f FrequencySet) Prune() FrequencySet {
	var out FrequencySet
	for _, b := range f {
		if b.Frequency > 0 {
			out = append(out, b)
		}
	}
	return out
}

// Balls returns numbers in whatever order the set is currently in
func (f FrequencySet) Balls() []int {
	var b []int
	for _, n := range f {
		b = append(b, n.Ball)
	}
	return b
}

// Asc orders balls by least to most frequently drawn
func (f FrequencySet) Asc() FrequencySet {
	sort.Sort(f)

	return f
}

// Desc orders balls by most to least frequently drawn
func (f FrequencySet) Desc() FrequencySet {
	sort.Sort(sort.Reverse(f))

	return f
}

// Draw returns n numbers at random from set
func Draw(set []int, n int) []int {
	// ResultSet a rand seed
	rand.Seed(time.Now().UnixNano())

	var out []int
	for i := 0; i < n && len(set) > 0; i++ {
		// Select index for this draw
		idx := rand.Intn(len(set))
		// Append pick to the output
		out = append(out, set[idx])
		// Remove item from set for next draw
		set = append(set[:idx], set[idx+1:]...)
	}
	sort.Ints(out)

	return out
}
