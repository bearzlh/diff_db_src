package main

import (
	"fmt"
	"testing"

	"diff_db/database"
	"diff_db/helper"
)

func TestLevel(t *testing.T) {
	a := map[string]interface{}{"a": 1, "b": 0, "c": 3, "d": 4}
	s := helper.SortMap2Slice(a, func(s1 helper.SliceItem, s2 helper.SliceItem) bool {
		return s1.Value.(int) < s2.Value.(int)
	})

	fmt.Println(s)
}

func TestQuery(t *testing.T) {
	database.InitMysql()
	database.CmdChannel = make(chan int, 4)
	database.DataDiff("cps", "orders")
}
