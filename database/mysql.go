package database

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"xorm.io/core"

	"diff_db/config"
	"diff_db/log"
)

var EngineMap map[string]*xorm.Engine
var EngineCompareMap map[string]*xorm.Engine

type Conn struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func InitMysql() {
	if valid := CheckConfigValid(); !valid {
		return
	}

	EngineMap = map[string]*xorm.Engine{}
	EngineCompareMap = map[string]*xorm.Engine{}
	LoadEngine("mysql")
	LoadEngine("mysql-compare")
}

// 初始化连接
func LoadEngine(prefix string) {
	var cf map[string]interface{}
	if prefix == "mysql" {
		cf = config.MysqlConfig.AllSettings()
	} else {
		cf = config.MysqlCompareConfig.AllSettings()
	}
	for k, v := range cf {
		conn := GetDbConnection(v)
		if conn == "" {
			log.Error(fmt.Sprintf("数据库配置有误:%s", k))
		} else {
			engine, err := xorm.NewEngine("mysql", conn)
			if config.Config.GetBool("debug") {
				engine.SetLogger(&log.Log{})
				engine.Logger().SetLevel(core.LOG_DEBUG)
				engine.ShowSQL(true)
				engine.SetMaxOpenConns(100)
			}
			if err != nil {
				log.Error(fmt.Sprintf("数据库链接失败:%v", err))
				os.Exit(1)
			}
			if strings.Contains(prefix, "compare") {
				EngineCompareMap[k] = engine
			} else {
				EngineMap[k] = engine
			}
		}
	}
}

func CheckConfigValid() bool {
	m1 := config.MysqlConfig.AllSettings()
	m2 := config.MysqlCompareConfig.AllSettings()
	err := 0
	for k1, _ := range m1 {
		if _, ok := m2[k1]; !ok {
			log.Errorf("数据库配置缺失:%s", k1)
			err = 1
		}
	}
	if err == 1 {
		log.Fatal("数据库配置错误")
	}
	return true
}

func GetDbConnection(v interface{}) string {
	mapDb := v.(map[string]interface{})
	Con := Conn{
		Host:     mapDb["host"].(string),
		Port:     mapDb["port"].(string),
		Database: mapDb["database"].(string),
		Username: mapDb["username"].(string),
		Password: mapDb["password"].(string),
	}

	return getConString(&Con)
}

func getConString(c *Conn) string {
	if c.Username == "" || c.Password == "" || c.Host == "" || c.Port == "" || c.Database == "" {
		return ""
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", c.Username, c.Password, c.Host, c.Port, c.Database)
	log.Debugf("connection:%s", dsn)
	return dsn
}
