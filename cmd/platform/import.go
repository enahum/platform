// Copyright (c) 2016 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.
package main

import (
	"errors"
	"os"

	"fmt"
	"github.com/mattermost/platform/app"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data.",
}

var slackImportCmd = &cobra.Command{
	Use:     "slack [team] [file]",
	Short:   "Import a team from Slack.",
	Long:    "Import a team from a Slack export zip file.",
	Example: "  import slack myteam slack_export.zip",
	RunE:    slackImportCmdF,
}

var bulkImportCmd = &cobra.Command{
	Use:     "bulk [file]",
	Short:   "Import bulk data.",
	Long:    "Import data from a Mattermost Bulk Import File.",
	Example: "  import bulk bulk_data.json",
	RunE:    bulkImportCmdF,
}

func init() {
	bulkImportCmd.Flags().Bool("apply", false, "Save the import data to the database. Use with caution - this cannot be reverted.")

	importCmd.AddCommand(
		bulkImportCmd,
		slackImportCmd,
	)
}

func slackImportCmdF(cmd *cobra.Command, args []string) error {
	initDBCommandContextCobra(cmd)

	if len(args) != 2 {
		return errors.New("Incorrect number of arguments.")
	}

	team := getTeamFromTeamArg(args[0])
	if team == nil {
		return errors.New("Unable to find team '" + args[0] + "'")
	}

	fileReader, err := os.Open(args[1])
	if err != nil {
		return err
	}
	defer fileReader.Close()

	fileInfo, err := fileReader.Stat()
	if err != nil {
		return err
	}

	CommandPrettyPrintln("Running Slack Import. This may take a long time for large teams or teams with many messages.")

	app.SlackImport(fileReader, fileInfo.Size(), team.Id)

	CommandPrettyPrintln("Finished Slack Import.")

	return nil
}

func bulkImportCmdF(cmd *cobra.Command, args []string) error {
	initDBCommandContextCobra(cmd)

	apply, err := cmd.Flags().GetBool("apply")
	if err != nil {
		return errors.New("Apply flag error")
	}

	if len(args) != 1 {
		return errors.New("Incorrect number of arguments.")
	}

	fileReader, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer fileReader.Close()

	if apply {
		CommandPrettyPrintln("Running Bulk Import. This may take a long time.")
	} else {
		CommandPrettyPrintln("Running Bulk Import Data Validation.")
		CommandPrettyPrintln("** This checks the validity of the entities in the data file, but does not persist any changes **")
		CommandPrettyPrintln("Use the --apply flag to perform the actual data import.")
	}

	CommandPrettyPrintln("")

	if err, lineNumber := app.BulkImport(fileReader, !apply); err != nil {
		CommandPrettyPrintln(err.Error())
		if lineNumber != 0 {
			CommandPrettyPrintln(fmt.Sprintf("Error occurred on data file line %v", lineNumber))
		}
	}

	if apply {
		CommandPrettyPrintln("Finished Bulk Import.")
	} else {
		CommandPrettyPrintln("Validation complete. You can now perform the import by rerunning this command with the --apply flag.")
	}

	return nil
}
