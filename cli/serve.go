/*
 * package cli defines the command line interface (CLI) for the Tournabyte identity provider service
 */
package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tournabyte/idp/api"
	"github.com/tournabyte/idp/model"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the IdP web server",
	Run:   doCommand,
}

func init() {
	serveCmd.Flags().Int("port", 8080, "Port for the application to listen on")
	serveCmd.Flags().StringSlice("mongo", []string{"localhost:27017"}, "Comma-separated list of hosts for mongo db persistence functionality")
	serveCmd.Flags().String("dbname", "local", "Database to utilize for persistence layer")
	serveCmd.Flags().String("dbuser", "", "Database identity to access mongo instance")
	serveCmd.Flags().String("dbpass", "", "Database access key for authenticating to mongo instance")

	rootCmd.AddCommand(serveCmd)

	viper.BindPFlag("serve.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("mongo.hosts", serveCmd.Flags().Lookup("mongo"))
	viper.BindPFlag("mongo.database", serveCmd.Flags().Lookup("dbname"))
	viper.BindPFlag("mongo.username", serveCmd.Flags().Lookup("dbuser"))
	viper.BindPFlag("mongo.password", serveCmd.Flags().Lookup("dbpass"))
}

func getCommandOpts() model.CommandOpts {
	var cmdOpts model.CommandOpts

	cmdOpts.Port = viper.GetInt("serve.port")
	cmdOpts.Dbhosts = viper.GetStringSlice("mongo.hosts")
	cmdOpts.Dbname = viper.GetString("mongo.database")
	cmdOpts.Dbuser = viper.GetString("mongo.username")
	cmdOpts.Dbpass = viper.GetString("mongo.password")

	return cmdOpts
}

func doCommand(cmd *cobra.Command, args []string) {

	server, err := api.NewIdentityProviderServer(getCommandOpts())

	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	server.Run()
}
