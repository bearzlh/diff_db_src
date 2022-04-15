package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	Config             *viper.Viper
	MysqlConfig        *viper.Viper
	MysqlCompareConfig *viper.Viper
)

func init() {
	var err error
	Config = viper.New()
	Config.SetConfigFile("config.json") // name of config file (without extension)
	err = Config.ReadInConfig()         // Find and read the config file
	if err != nil {                          // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error mysql config file: %w \n", err))
	}

	MysqlConfig = viper.New()
	MysqlConfig.SetConfigFile("mysql.json") // name of config file (without extension)
	err = MysqlConfig.ReadInConfig()        // Find and read the config file
	if err != nil {                         // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error mysql config file: %w \n", err))
	}

	MysqlCompareConfig = viper.New()
	MysqlCompareConfig.SetConfigFile("mysql-compare.json") // name of config file (without extension)
	err = MysqlCompareConfig.ReadInConfig()        // Find and read the config file
	if err != nil {                                // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error mysql-compare config file: %w \n", err))
	}
}
