/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// repoCmd represents the repo command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Work with GitHub repositories.",
	Long: `The repo command groups repository-oriented operations such as listing
repositories and transferring them between organizations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("repo called")
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// repoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// repoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
