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

package cmd

import (
	"fmt"
	"time"

	"github.com/nboughton/stalotto/lotto"
	"github.com/spf13/cobra"
)

// dipCmd represents the dip command
var dipCmd = &cobra.Command{
	Use:   "dip",
	Short: "Draw some random balls",
	Long: `Dip is not entirely random, it sorts the results since late 2015 (when the number
of balls was increased to 59) and removes the least drawn half before randomly drawing
a set.`,
	Run: func(cmd *cobra.Command, args []string) {
		begin := time.Date(2015, time.October, 10, 0, 0, 0, 0, time.Local)
		end := time.Now()

		set := lotto.ResultSet{}
		for res := range appDB.Results(begin, end, []string{}, []int{}) {
			set = append(set, res)
		}

		balls, bonus := set.ByDrawFrequency()
		numbers := balls.Prune().Desc().Balls()[:len(balls.Prune().Desc().Balls())/2]
		bonuses := bonus.Prune().Desc().Balls()[:10]

		fmt.Fprintf(tw, "Balls:\t%v\nBonus:\t%v\n", lotto.Draw(numbers, 6), lotto.Draw(bonuses, 1))
		tw.Flush()
	},
}

func init() {
	RootCmd.AddCommand(dipCmd)
}
