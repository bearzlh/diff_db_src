/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"

	"diff_db/config"
	"diff_db/database"
	"diff_db/helper"
	"diff_db/log"
)

var Lk sync.RWMutex

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "mysql",
	Short: "数据库对比",
	Long:  `可对比数据表，数据字段，数据量。可以选择某个数据表进行对比`,
	Run: func(cmd *cobra.Command, args []string) {
		dbString, _ := cmd.Flags().GetString("dbs")
		typeString, _ := cmd.Flags().GetString("types")
		tableString, _ := cmd.Flags().GetString("tables")
		parallel, _ := cmd.Flags().GetInt("parallel")
		var dbList []string
		var tableList []string
		if dbString != "" {
			dbList = strings.Split(dbString, ",")
			for _, s := range dbList {
				c := config.MysqlConfig.Get(s)
				if c == nil {
					log.Fatalf("配置不存在:%s", s)
				}
			}
		} else {
			for k, _ := range config.MysqlConfig.AllSettings() {
				dbList = append(dbList, k)
			}
		}
		if tableString != "" {
			tableList = strings.Split(tableString, ",")
		}
		if tableString != "" && len(dbList) > 1 {
			log.Fatalf("过滤数据表时数据库只能选择一个")
		}
		debug, _ := cmd.Flags().GetBool("debug")
		if debug {
			config.Config.Set("debug", true)
		}
		database.InitMysql()
		uiprogress.Start()
		GetTotalTask(typeString, tableList, dbList, parallel)
		log.Infof("任务数:%d", database.TotalTask)
		if database.TotalTask == 0 {
			log.Fatalf("无任务需要执行")
		}
		database.CmdChannel = make(chan int, parallel)
		database.TaskChannel = make(chan int, database.TotalTask)
		go func() {
			for _, db := range dbList {
				if strings.Contains(typeString, "table") {
					go func(db string) {
						database.TableCheck(db)
					}(db)
				}
				var tl []string
				// 过滤数据表时不进行数据表对比
				if len(tableList) == 0 {
					tl = database.GetDbTables(db)
				} else {
					tl = tableList
				}
				if strings.Contains(typeString, "field") {
					go func(db string, tableList []string) {
						database.FieldCheck(db, tableList)
					}(db, tl)
				}
				if strings.Contains(typeString, "max") {
					go func(db string, tableList []string) {
						database.MaxCheck(db, tableList)
					}(db, tl)
				}
			}
		}()

		taskFinish := 0
		bar := uiprogress.AddBar(database.TotalTask)
		bar.AppendCompleted()
		bar.AppendElapsed()
		bar.AppendFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("任务,异常(%4d/%4d,%4d)", b.Current(), database.TotalTask, database.TaskA.GetErrorTask())
		})
		for {
			select {
			case <-database.TaskChannel:
				bar.Incr()
				taskFinish++
				log.Debugf("finishTask:%d", taskFinish)
				break
			}

			if taskFinish == database.TotalTask {
				log.Debugf("finished")
				break
			}
		}
		uiprogress.Stop()

		fmt.Println("\n错误统计如下")
		PrintSort(database.TaskA.ErrorData)

		log.Info("任务结束")
	},
}

// 打印统计数据
func PrintSort(mp map[string]int) {
	data := map[string]interface{}{}
	nameLength := 10
	for s, i := range mp {
		if len(s) > nameLength {
			nameLength = len(s)
		}
		data[s] = i
	}
	s := helper.SortMap2Slice(data, func(s1 helper.SliceItem, s2 helper.SliceItem) bool {
		return s1.Value.(int) < s2.Value.(int)
	})
	tpl := fmt.Sprintf("%d", nameLength)
	for _, item := range s {
		fmt.Printf("库名:%-"+tpl+"s 错误数:%-5d\n", item.Key, item.Value)
	}
}

func GetTotalTask(typeString string, tableList []string, dbList []string, parallel int) int {
	bar := uiprogress.AddBar(len(dbList))
	bar.AppendCompleted()
	bar.AppendElapsed()
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("连接,任务(%4d/%4d,%4d)", b.Current(), len(dbList), database.TotalTask)
	})
	wg := sync.WaitGroup{}
	limitCheck := make(chan int, parallel)
	wg.Add(len(dbList))
	for _, db := range dbList {
		limitCheck <- 1
		go func(db string) {
			taskCount := 0
			var tableCount int
			if len(tableList) == 0 {
				if strings.Contains(typeString, "table") {
					taskCount += 1
				}
				tableCount = len(database.GetDbTables(db))
			} else {
				tableCount = len(tableList)
			}
			if strings.Contains(typeString, "field") {
				log.Infof("数据表数量:%s,%d", db, tableCount)
				taskCount += tableCount
			}
			if strings.Contains(typeString, "max") {
				taskCount += tableCount
			}
			database.TaskA.AddTotal(db, taskCount)
			Lk.Lock()
			database.TotalTask += taskCount
			Lk.Unlock()
			bar.Incr()
			wg.Done()
			<-limitCheck
		}(db)
	}
	wg.Wait()

	return database.TotalTask
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.PersistentFlags().String("dbs", "", "配置文件中数据库连接名。可选多个，英文逗号隔开。默认全部")
	diffCmd.PersistentFlags().String("types", "table,field,max", "可选多个，英文逗号隔开。可选:table,field,max。分别对比表，表字段，表数据")
	diffCmd.PersistentFlags().String("tables", "", "可选多个，英文逗号隔开。默认全部")
	diffCmd.PersistentFlags().Bool("debug", false, "是否debug，默认不开启。debug会输出详细信息，包括SQL日志")
	diffCmd.PersistentFlags().Int("parallel", 4, "协程并发次数。填写过大会造成客户端&数据库压力升高")
}
