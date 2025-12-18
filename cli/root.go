/*
 * package cli defines the command line interface (CLI) for the Tournabyte identity provider service
 */
package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tournabyte/idp/model"
)

var rootCmd = &cobra.Command{
	Use:              "tbyte-idp",
	Short:            "tbyte-idp CLI controls the identity provider service for Tournabyte",
	PersistentPreRun: getConfig,
}

var appConf *model.ApplicationConfiguration = model.NewApplicationConfiguration("json", "appconf", []string{"/etc/tournabyte/idp", "$HOME/.local/tournabyte/idp", "."})

func Execute() {
	rootCmd.Execute()
}

func getConfig(cmd *cobra.Command, args []string) {
	log.Println("In `getConfig` persistent pre-run hook")

	if err := appConf.PopulateConfiguration(); err != nil {
		log.Fatalf("Error reading config from file: %v", err)
	}
	log.Printf("Application configuration after reading config files")
	log.Printf("\tserve.port: %v", appConf.GetValue("serve.port"))
	log.Printf("\tserve.jwt.key: %v", appConf.GetValue("serve.jwt.key"))
	log.Printf("\tserve.jwt.leeway: %v", appConf.GetValue("serve.jwt.leeway"))
	log.Printf("\tdatastore.hosts: %v", appConf.GetValue("datastore.hosts"))
	log.Printf("\tdatastore.username: %v", appConf.GetValue("datastore.username"))
	log.Printf("\tdatastore.password: %v", appConf.GetValue("datastore.password"))
}
