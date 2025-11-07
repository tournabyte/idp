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
	port       int
	mongo_host string
	mongo_user string
	mongo_pass string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the IdP web server",
	Run:   run,
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVarP(&mongo_host, "dbhost", "n", "", "Database hostname")
	serveCmd.Flags().StringVarP(&mongo_user, "dbuser", "u", "", "Database username")
	serveCmd.Flags().StringVarP(&mongo_pass, "dbpass", "w", "", "Database password")
	rootCmd.AddCommand(serveCmd)
}

func run(cmd *cobra.Command, args []string) {
	log.Printf("Starting service on port %d\n", port)
	server, err := api.NewIdentityProviderServer(
		mongo_host,
		mongo_user,
		mongo_pass,
	)

	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	server.ConfigureServer()
	server.RunServer(port)
}
