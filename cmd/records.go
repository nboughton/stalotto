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
	"github.com/nboughton/stalotto/lotto"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	flBegin   = "begin"
	flEnd     = "end"
	flMachine = "machine"
	flSet     = "set"
)

var fmtDate = "2006-01-02"

// recordsCmd represents the records command
var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Retrieve and print a record set",
	Long:  `--begin and --end dates must be formatted as YYYY-MM-DD`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, rec := range resultsQuery(cmd) {
			fmt.Println(rec)
		}
	},
}

func resultsQuery(cmd *cobra.Command) lotto.Set {
	set := lotto.Set{}
	for res := range appDB.GetRecords(parseRecordsQueryFlags(cmd)) {
		set = append(set, res)
	}

	return set
}

func parseRecordsQueryFlags(cmd *cobra.Command) (time.Time, time.Time, []string, []int) {
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
	RootCmd.AddCommand(recordsCmd)
	recordsCmd.PersistentFlags().String(flBegin, "2015-09-10", "Set beginning date for query")
	recordsCmd.PersistentFlags().String(flEnd, time.Now().Format(fmtDate), "Set end date for query")
	recordsCmd.PersistentFlags().StringArrayP(flMachine, "m", []string{}, "Constrain results by machine")
	recordsCmd.PersistentFlags().IntSliceP(flSet, "s", []int{}, "Constrain results by Set")
}
