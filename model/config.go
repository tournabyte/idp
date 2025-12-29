/*
 * package model describes the data types utilized by the idp service
 */

package model

import (
	"log"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ApplicationOptions struct {
	Serve struct {
		Port     int `mapstructure:"port"`
		WebToken struct {
			Key    string        `mapstructure:"key"`
			Leeway time.Duration `mapstructure:"leeway"`
		} `mapstructure:"jwt"`
	} `mapstructure:"serve"`
	Datastore struct {
		Hosts    []string
		Username string
		Password string
	} `mapstructure:"datastore"`
}

type ApplicationConfiguration struct {
	config *viper.Viper
}

func NewApplicationConfiguration(cfgType string, cfgName string, cfgPaths []string) *ApplicationConfiguration {
	appConfig := ApplicationConfiguration{config: viper.New()}

	appConfig.config.SetConfigName(cfgName)
	appConfig.config.SetConfigType(cfgType)

	for _, p := range cfgPaths {
		appConfig.config.AddConfigPath(p)
	}

	return &appConfig
}

func (appconf ApplicationConfiguration) PopulateConfiguration() error {
	return appconf.config.ReadInConfig()
}

func (appconf ApplicationConfiguration) ApplyFlags(flags *pflag.FlagSet, optionBindings map[string]string) error {
	for optionName, flagName := range optionBindings {
		err := appconf.config.BindPFlag(optionName, flags.Lookup(flagName))
		if err != nil {
			return err
		}
	}
	return nil
}

func (appconf ApplicationConfiguration) GetOptions() (*ApplicationOptions, error) {
	var options ApplicationOptions

	log.Printf("Settings: %v", appconf.config.AllSettings())

	err := appconf.config.Unmarshal(&options)
	if err != nil {
		return nil, err
	}
	return &options, nil
}

func (appconf ApplicationConfiguration) GetValue(key string) any {
	return appconf.config.Get(key)
}
