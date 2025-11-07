/*
 * package cli defines the command line interface (CLI) for the Tournabyte identity provider service
 */
package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tbyte-idp",
	Short: "tbyte-idp CLI controls the identity provider service for Tournabyte",
}

func Execute() {
	rootCmd.Execute()
}
