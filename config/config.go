package config

import (
	_ "fmt"
	"log"
	"os"

	"github.com/theherk/viper"
)

var (
	Config *viper.Viper
)

func init() {

	// set file path
	file_name := "config\\config.toml"
	viper.SetConfigFile(file_name)
	viper.AddConfigPath(".")

	// if file not existing then create it
	f, err := os.OpenFile(file_name, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("Fatal error when opening config file: %s \n", err)
	}
	f.Close()

	// Read config file
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalln("Error to read config file: %s \n", err)
	}
	Config = viper.GetViper()
}
