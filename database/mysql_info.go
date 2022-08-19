package database

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"

	"diff_db/config"
	"diff_db/log"
)

// CmdChannel 任务并行执行数量
var CmdChannel chan int

// TaskChannel 任务状态追踪
var TaskChannel chan int

// TotalTask 总任务数
var TotalTask int

type TaskAnalysis struct {
	Rw         sync.RWMutex
	ErrorData  map[string]int
	TaskTotal  map[string]int
	TaskFinish map[string]int
}

func (T *TaskAnalysis) AddTotal(db string, total int) {
	T.Rw.Lock()
	if _, e := T.TaskTotal[db]; e {
		T.TaskTotal[db] += total
	} else {
		T.TaskTotal[db] = total
	}
	T.Rw.Unlock()
}

func (T *TaskAnalysis) GetTotalTask(db string) int {
	T.Rw.Lock()
	total := T.TaskTotal[db]
	T.Rw.Unlock()
	return total
}

func (T *TaskAnalysis) IncreaseError(db string) {
	T.Rw.Lock()
	if _, e := T.ErrorData[db]; e {
		T.ErrorData[db]++
	} else {
		T.ErrorData[db] = 1
	}
	T.Rw.Unlock()
}

func (T *TaskAnalysis) GetError(db string) int {
	T.Rw.Lock()
	if r, e := T.ErrorData[db]; e {
		T.Rw.Unlock()
		return r
	}
	T.Rw.Unlock()
	return 0
}

func (T *TaskAnalysis) GetErrorTask() int {
	T.Rw.Lock()
	result := 0
	for _, i := range T.ErrorData {
		result += i
	}
	T.Rw.Unlock()
	return result
}

func (T *TaskAnalysis) IncreaseFinish(db string) {
	T.Rw.Lock()
	if _, e := T.TaskFinish[db]; e {
		T.TaskFinish[db]++
	} else {
		T.TaskFinish[db] = 1
	}
	T.Rw.Unlock()
}

func (T *TaskAnalysis) GetFinish(db string) int {
	T.Rw.Lock()
	if r, e := T.TaskFinish[db]; e {
		T.Rw.Unlock()
		return r
	}
	T.Rw.Unlock()
	return 0
}

var TaskA TaskAnalysis

func init() {
	TaskA = TaskAnalysis{
		Rw:         sync.RWMutex{},
		ErrorData:  map[string]int{},
		TaskTotal:  map[string]int{},
		TaskFinish: map[string]int{},
	}
}

// TableCheck 数据表对比
func TableCheck(db string) {
	CmdChannel <- 1
	r1, _ := EngineMap[db].QueryString("show tables")
	r2, _ := EngineCompareMap[db].QueryString("show tables")
	// 将列表转化为map，查询key是否存在
	m1 := map[string]string{}
	m2 := map[string]string{}
	for _, v := range r1 {
		for _, v2 := range v {
			m1[v2] = ""
		}
	}
	for _, v := range r2 {
		for _, v2 := range v {
			m2[v2] = ""
		}
	}
	for k := range m1 {
		if _, ok := m2[k]; !ok {
			log.Errorf("数据表缺失:%s,%s", db, k)
			TaskA.IncreaseError(db)
		}
	}
	TaskChannel <- 1
	<-CmdChannel
	TaskA.IncreaseFinish(db)
}

// GetDbTables 获取数据库中的数据表
func GetDbTables(db string) []string {
	tableInfo, _ := EngineMap[db].QueryString("show tables")
	var tables []string
	for _, item := range tableInfo {
		for _, table := range item {
			tables = append(tables, table)
		}
	}

	return tables
}

// FieldCheck 单表结构对比
func FieldCheck(db string, tables []string) {
	for _, table := range tables {
		CmdChannel <- 1
		FieldDiff(db, table)
		TaskChannel <- 1
		<-CmdChannel
		TaskA.IncreaseFinish(db)
	}
}

// FieldDiff 字段对比
func FieldDiff(db string, table string) {
	r1, _ := EngineMap[db].QueryString(fmt.Sprintf("desc %s", table))
	r2, _ := EngineCompareMap[db].QueryString(fmt.Sprintf("desc %s", table))
	m1 := map[string]map[string]string{}
	m2 := map[string]map[string]string{}
	// 字段映射
	for _, m := range r1 {
		m1[m["Field"]] = m
	}
	for _, m := range r2 {
		m2[m["Field"]] = m
	}
	// 对比
	for s, v := range m1 {
		// 判断字段缺失
		if v1, ok := m2[s]; ok {
			// 判断属性
			for k2, v2 := range v {
				if v2 != v1[k2] {
					log.Errorf("字段属性不一致:%s %s.%s 属性:%s '%s' vs '%s'", db, table, s, k2, v2, v1[k2])
					TaskA.IncreaseError(db)
				}
			}
		} else {
			log.Errorf("字段缺失:%s,%s,%s", db, table, s)
			TaskA.IncreaseError(db)
		}
	}
}

// MaxCheck 数量对比
func MaxCheck(db string, tables []string) {
	for _, table := range tables {
		CmdChannel <- 1
		MaxDiff(db, table)
		TaskChannel <- 1
		<-CmdChannel
		TaskA.IncreaseFinish(db)
	}
}

// MaxDiff max对比
func MaxDiff(db string, table string) {
	var err1, err2 error
	priName := GetPriKey(db, table)
	var query string
	if priName != "" {
		query = fmt.Sprintf("select max(%s) as m from %s", priName, table)
	} else {
		query = fmt.Sprintf("select count(1) as m from %s", table)
	}
	var r1, r2 []map[string]string
	r1, err1 = EngineMap[db].QueryString(query)
	if err1 != nil {
		log.Errorf("1查询错误:%s,%s %v", db, table, err1)
	}
	r2, err2 = EngineCompareMap[db].QueryString(query)
	if err2 != nil {
		log.Errorf("2查询错误:%s,%s %v", db, table, err2)
	}
	if err1 == nil && err2 == nil {
		n1, _ := strconv.Atoi(r1[0]["m"])
		n2, _ := strconv.Atoi(r2[0]["m"])
		maxThreshold := config.Config.GetInt("max_threshold")
		if math.Abs(float64(n1-n2)) > float64(maxThreshold) {
			log.Errorf("对比数据量不一致:%s %s |%d - %d| > %d", db, table, n1, n2, maxThreshold)
			TaskA.IncreaseError(db)
		}
	}
}

// DataCheck 数据对比
func DataCheck(db string, tables []string) {
	for _, table := range tables {
		CmdChannel <- 1
		DataDiff(db, table)
		TaskChannel <- 1
		<-CmdChannel
		TaskA.IncreaseFinish(db)
	}
}

// DataDiff 数据对比
func DataDiff(db string, table string) {
	priName := GetPriKey(db, table)
	if priName != "" {
		dataThreshold := config.Config.GetInt("data_threshold")
		query1 := fmt.Sprintf("select * from %s order by %s desc limit %d", table, priName, dataThreshold)
		query2 := fmt.Sprintf("select * from %s order by %s desc limit %d", table, priName, dataThreshold)
		r1, err1 := EngineMap[db].QueryString(query1)
		r2, err2 := EngineCompareMap[db].QueryString(query2)
		if err1 == nil && err2 == nil {
			data1, _ := json.Marshal(r1)
			sData1 := string(data1)
			data2, _ := json.Marshal(r2)
			sData2 := string(data2)
			if sData1 != sData2 {
				log.Errorf("对比数据不一致:%s %s", db, table)
				log.Errorf("对比数据不一致:data1:%s", sData1)
				log.Errorf("对比数据不一致:data2:%s", sData2)
				TaskA.IncreaseError(db)
			}
		} else {
			log.Error("查询有误，请联系管理员")
		}
	} else {
		log.Error("主键不一致，不进行数据对比")
	}
}

// GetPriKey 获取主键
func GetPriKey(db string, table string) string {
	r1, err := EngineMap[db].QueryString(fmt.Sprintf("show index from %s where Key_name='PRIMARY';", table))
	if err != nil || len(r1) == 0 {
		return ""
	}

	return r1[0]["Column_name"]
}
