/*
 * package cli defines the command line interface (CLI) for the Tournabyte identity provider service
 */
package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tournabyte/idp/api"
)

var (
	port int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the IdP web server",
	Run:   run,
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}

func run(cmd *cobra.Command, args []string) {
	log.Printf("Starting service on port %d\n", port)
	server, err := api.NewIdentityProviderServer()

	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	server.ConfigureServer()
	server.RunServer(port)
}
