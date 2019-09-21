package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"zorm"
)

type Crop struct {
	Name    string `zorm:"type:varchar(32);not null"`
	IconUrl string `zorm:"type:varchar(511);not null"`
	Note    string `zorm:"type:TEXT"`
	ShopId  uint   `zorm:"not null"`
}

func main() {
	testDB, err := zorm.Open("mysql", "root:wshwoaini@/auto_fertilizer?charset=utf8&parseTime=true&loc=Asia%2FShanghai")
	if err != nil {
		fmt.Println(err)
	}
	crops := make([]Crop, 0)
	testDB = testDB.Where("id in (?)", []uint{11, 12, 13}).Find(&crops)
	println(testDB.RowsAffected)
	fmt.Println(crops)
}
