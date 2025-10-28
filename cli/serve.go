/*
* package cli defines the command line interface (CLI) for the Tournabyte identity provider service
 */
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	port int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the IdP web server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting service on port %d\n", port)
		//TODO: Instatiate a server instance and listen
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}
