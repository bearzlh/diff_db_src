package log

import (
	"fmt"
	"os"

	. "github.com/logrusorgru/aurora/v3"
	"xorm.io/core"

	"diff_db/config"
)

type Log struct {
	level   core.LogLevel
	showSql bool
}

func (L *Log) Debug(v ...interface{}) {
	Debug(Sprintf("%s", v))
}
func (L *Log) Debugf(format string, v ...interface{}) {
	Debugf(format, v...)
}
func (L *Log) Error(v ...interface{}) {
	Error(Sprintf("%s", v))
}
func (L *Log) Errorf(format string, v ...interface{}) {
	Errorf(format, v...)
}
func (L *Log) Info(v ...interface{}) {
	Info(Sprintf("%s", v))
}
func (L *Log) Infof(format string, v ...interface{}) {
	Infof(format, v...)
}
func (L *Log) Warn(v ...interface{}) {
	Error(Sprintf("%s", v))
}
func (L *Log) Warnf(format string, v ...interface{}) {
	Errorf(format, v...)
}
func (L *Log) Level() core.LogLevel {
	return L.level
}
func (L *Log) SetLevel(l core.LogLevel) {
	L.level = l
}
func (L *Log) ShowSQL(show ...bool) {
	L.showSql = show[0]
}
func (L *Log) IsShowSQL() bool {
	return L.showSql
}

func Info(input string) {
	WriteLog(fmt.Sprintln(input))
}

func Debug(input string) {
	if config.Config.GetBool("debug") {
		WriteLog(fmt.Sprintln(input))
	}
}

func Error(input string) {
	WriteLog(fmt.Sprintln(Magenta(input)))
}

func Fatal(input string) {
	WriteLog(fmt.Sprintln(Red(input)))
	fmt.Println(Red(input))
	os.Exit(1)
}

func Infof(format string, a ...interface{}) {
	Info(Sprintf(format, a...))
}

func Debugf(format string, a ...interface{}) {
	Debug(Sprintf(format, a...))
}

func Errorf(format string, a ...interface{}) {
	Error(Sprintf(format, a...))
}

func Fatalf(format string, a ...interface{}) {
	Fatal(Sprintf(format, a...))
}

func WriteLog(s string) {
	f, err := os.OpenFile("./debug.log", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("文件打开失败:%v", err)
	}
	defer f.Close()
	n, _ := f.Seek(0, 2)
	_, err = f.WriteAt([]byte(s), n)
	if err != nil {
		fmt.Printf("文件写入失败:%v", err)
	}
}
