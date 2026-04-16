/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dev-soprasteriano/gheese/internal/github"
	"github.com/spf13/cobra"
)

// costcenterCmd represents the costcenter command
var costcenterCmd = &cobra.Command{
	Use:   "costcenter",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := github.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		users, err := github.GetUsersMissingCC(c, "soprasteriasca")
		if err != nil {
			fmt.Println(err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "USER\tCOST CENTER")
		for _, user := range users {
			fmt.Fprintf(w, "%s\t%s\n", user.Name, user.CostCenter)
		}
		_ = w.Flush()
	},
}

func init() {
	enterpriseCmd.AddCommand(costcenterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// costcenterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// costcenterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
