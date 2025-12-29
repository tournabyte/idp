/*
 * package cli defines the command line interface (CLI) for the Tournabyte identity provider service
 */
package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tournabyte/idp/api"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the IdP web server",
	Run:   doCommand,
}

func init() {
	serveCmd.Flags().Int("port", 8080, "Port for the application to listen on")
	serveCmd.Flags().StringSlice("dbhosts", []string{"localhost:27017"}, "Comma-separated list of hosts for mongo db persistence functionality")
	serveCmd.Flags().String("dbuser", "", "Database identity to access mongo instance")
	serveCmd.Flags().String("dbpass", "", "Database access key for authenticating to mongo instance")

	rootCmd.AddCommand(serveCmd)

	optionsToFlags := map[string]string{
		"serve.port":         "port",
		"datastore.hosts":    "dbhosts",
		"datastore.username": "dbuser",
		"datastore.password": "dbpass",
	}

	if err := appConf.ApplyFlags(serveCmd.Flags(), optionsToFlags); err != nil {
		log.Fatalf("Error applying flag overrides: %v", err)
	}

	log.Printf("Application configuration after applying plags")
	log.Printf("\tserve.port: %v", appConf.GetValue("serve.port"))
	log.Printf("\tserve.jwt.key: %v", appConf.GetValue("serve.jwt.key"))
	log.Printf("\tserve.jwt.leeway: %v", appConf.GetValue("serve.jwt.leeway"))
	log.Printf("\tdatastore.hosts: %v", appConf.GetValue("datastore.hosts"))
	log.Printf("\tdatastore.username: %v", appConf.GetValue("datastore.username"))
	log.Printf("\tdatastore.password: %v", appConf.GetValue("datastore.password"))

}

func doCommand(cmd *cobra.Command, args []string) {

	if opts, err := appConf.GetOptions(); err != nil {
		log.Fatalf("Application options could not be retrieved: %v", err)
	} else {
		log.Printf("Application Options:\n")
		log.Printf("\tServe.Port = %d", opts.Serve.Port)
		log.Printf("\tServe.WebToken.Key = %s", opts.Serve.WebToken.Key)
		log.Printf("\tServe.WebToken.Leeway = %s", opts.Serve.WebToken.Leeway.String())
		log.Printf("\tDatastore.Hosts = %v", opts.Datastore.Hosts)
		log.Printf("\tDatastore.Username = %v", opts.Datastore.Username)
		log.Printf("\tDatastore.Password = %v", opts.Datastore.Password)
		server, err := api.NewIdentityProviderServer(opts)

		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		server.Run()
	}
}
