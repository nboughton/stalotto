// Copyright © 2019 Nick Boughton <nicholasboughton@gmail.com>
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

	"github.com/spf13/cobra"
)

// setFrequencyCmd represents the setFrequency command
var setFrequencyCmd = &cobra.Command{
	Use:   "setFrequency",
	Short: "Show frequency of machine/set combinations in a date constrained data set",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		start, end, _, _ := parseQueryFlags(cmd)

		freqSets, err := appDB.MachineSetFreq(start, end)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Fprintf(tw, "Machine\tSet\tDraws\n")
		for _, s := range freqSets {
			fmt.Fprintf(tw, "%s\t%d\t%d\n", s.Machine, s.Set, s.Freq)
		}
		tw.Flush()
	},
}

func init() {
	resultsCmd.AddCommand(setFrequencyCmd)
}
