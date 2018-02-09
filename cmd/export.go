// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"

	"github.com/nboughton/go-utils/json/file"
	"github.com/nboughton/stalotto/db"
	"github.com/nboughton/stalotto/lotto"
	"github.com/spf13/cobra"
)

const (
	flExportFormat = "format" // I might add different export formats at some point
	flExportFile   = "output-file"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a record set as a json file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString(flDBPath)
		outputFile, _ := cmd.Flags().GetString(flExportFile)

		appDB := db.Connect(dbPath)

		set := lotto.Set{}
		for res := range appDB.GetRecords(parseRecordsQueryFlags(cmd)) {
			set = append(set, res)
		}

		if err := file.Write(outputFile, set); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	recordsCmd.AddCommand(exportCmd)
	exportCmd.Flags().String(flExportFile, "stalotto-export.json", "Set output file path/name")
}
