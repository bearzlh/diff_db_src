// Copyright © 2022 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"diff_db/config"
	"diff_db/log"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "增加配置文件中的数据库连接配置",
	Long:  `增加配置文件中的数据库连接配置`,
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetString("port")
		dbPrefix, _ := cmd.Flags().GetString("db_prefix")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		dbFrom, _ := cmd.Flags().GetInt("db_from")
		dbTo, _ := cmd.Flags().GetInt("db_to")
		section, _ := cmd.Flags().GetString("section")
		dbname, _ := cmd.Flags().GetString("database")
		backup, _ := cmd.Flags().GetBool("backup")
		var cf *viper.Viper
		var fileName string
		if section == "mysql-compare" {
			cf = config.MysqlCompareConfig
			fileName = "mysql-compare.json"
		} else {
			cf = config.MysqlConfig
			fileName = "mysql.json"
		}
		// 单库
		if dbname != "" {
			Db := map[string]string{
				"database": dbname,
				"host":     host,
				"password": password,
				"port":     port,
				"username": username,
			}
			cf.Set(dbname, Db)
		}
		// 分库
		if dbPrefix != "" && dbTo > dbFrom {
			for i := dbFrom; i <= dbTo; i++ {
				database := fmt.Sprintf("%s%d", dbPrefix, i)
				Db := map[string]string{
					"database": database,
					"host":     host,
					"password": password,
					"port":     port,
					"username": username,
				}
				cf.Set(database, Db)
			}
		}

		var err error
		if backup {
			// 备份配置文件
			input, _ := ioutil.ReadFile(fileName)
			err = ioutil.WriteFile(fileName+"."+time.Now().Format("15-04-05"), input, 0644)
			if err != nil {
				log.Fatalf("备份失败:%v", err)
			}
		}

		// 更新配置文件
		err = cf.WriteConfig()
		if err != nil {
			log.Fatalf("更新失败:%v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.PersistentFlags().String("host", "", "主机名或IP")
	configCmd.PersistentFlags().String("port", "3306", "端口")
	configCmd.PersistentFlags().String("database", "", "数据库名称-单库配置")
	configCmd.PersistentFlags().String("username", "", "用户名")
	configCmd.PersistentFlags().String("password", "", "密码")
	configCmd.PersistentFlags().String("section", "mysql", "加入到的配置，可选项：mysql|mysql-compare")
	configCmd.PersistentFlags().String("db_prefix", "", "数据库前缀-分库配置")
	configCmd.PersistentFlags().Int("db_from", 0, "数据库最小索引-分库配置")
	configCmd.PersistentFlags().Int("db_to", 0, "数据库最大索引-分库配置")
	configCmd.PersistentFlags().Bool("backup", false, "是否备份原配置文件")
}
