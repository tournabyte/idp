/*
 * package cli defines the command line interface (CLI) for the Tournabyte identity provider service
 */
package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:              "tbyte-idp",
	Short:            "tbyte-idp CLI controls the identity provider service for Tournabyte",
	PersistentPreRun: getConfig,
}

func Execute() {
	rootCmd.Execute()
}

func getConfig(cmd *cobra.Command, args []string) {
	log.Println("In `getConfig` persistene pre-run hook")

	viper.AddConfigPath("/etc/tournabyte/idp")
	viper.AddConfigPath("$HOME/.local/tournabyte/idp")
	viper.AddConfigPath(".")

	viper.SetConfigName("appconf")
	viper.SetConfigType("json")

	if configErr := viper.ReadInConfig(); configErr != nil {
		log.Fatalf("Error reading config from file: %v", configErr)
	}

	log.Printf("Serve config from file:\n{%v}\n", viper.Get("serve"))
	log.Printf("Mongo config from file:\n{%v}\n", viper.Get("mongo"))
}
