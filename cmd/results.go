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
	"os"
	"time"

	"github.com/nboughton/stalotto/lotto"
	"github.com/spf13/cobra"
)

const (
	flBegin   = "begin"
	flEnd     = "end"
	flMachine = "machine"
	flSet     = "set"
)

var fmtDate = "2006-01-02"

// resultsCmd represents the results command
var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Retrieve/Print/Export a result set",
	Long:  `--begin and --end dates must be formatted as YYYY-MM-DD`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(tw, "DATE\tSET\tMACHINE\tB1\tB2\tB3\tB4\tB5\tB6\tBONUS")
		for _, r := range resultsQuery(cmd) {
			fmt.Fprintf(tw, "%s\t%d\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n", r.Date.Format("06/01/02"), r.Set, r.Machine, r.Balls[0], r.Balls[1], r.Balls[2], r.Balls[3], r.Balls[4], r.Balls[5], r.Bonus)
		}
		tw.Flush()
	},
}

func resultsQuery(cmd *cobra.Command) lotto.ResultSet {
	set := lotto.ResultSet{}
	for res := range appDB.Results(parseQueryFlags(cmd)) {
		set = append(set, res)
	}

	return set
}

func parseQueryFlags(cmd *cobra.Command) (time.Time, time.Time, []string, []int) {
	bStr, _ := cmd.Flags().GetString(flBegin)
	begin, err := time.Parse(fmtDate, bStr)
	chkDateErr(err)

	eStr, _ := cmd.Flags().GetString(flEnd)
	end, err := time.Parse(fmtDate, eStr)
	chkDateErr(err)

	machines, _ := cmd.Flags().GetStringArray(flMachine)
	sets, _ := cmd.Flags().GetIntSlice(flSet)

	return begin, end, machines, sets
}

func chkDateErr(e error) {
	if e != nil {
		fmt.Printf("Invalid date (%s). Ensure format is YYYY-MM-DD\n", e)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(resultsCmd)
	resultsCmd.PersistentFlags().String(flBegin, "2015-10-10", "Set beginning date for query")
	resultsCmd.PersistentFlags().String(flEnd, time.Now().Format(fmtDate), "Set end date for query")
	resultsCmd.PersistentFlags().StringArrayP(flMachine, "m", []string{}, "Constrain results by machine")
	resultsCmd.PersistentFlags().IntSliceP(flSet, "s", []int{}, "Constrain results by Set")
}
