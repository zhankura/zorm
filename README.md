# Zorm
*学习于gorm的一个golang orm*
## Examle

### 基础curd

```golang
type Test struct {
    ID uint `primary_key`
    Name string `zorm:"type:varchar(32);not null"`
    Out string `zorm:"type:varchar(128);not null"`
}
func main(){
    newDB, err := zorm.Open("mysql", "user:password@/database?charset=utf8&parseTime=true&loc=Asia%2FShanghai")
    if err != nil {
        fmt.Println(err)
    }
    var test Test
    newDB.Where("id = ?", 1).Find(&test)
    newDB.Where("id = ?", 1).First(&test)
    queryMap := map[string]interface{}{"id":1}
    newDB.Where(queryMap).Find(&test)
    tests := make([]Test, 0)
    newDB.Where(queryMap).Find(&tests)
    test = Test{
        Name:"test1",
        Out:"This is a test",
    }
    newDB.Insert(test)
    newDB.Where("id in (?)", []uint{1,2}).Update(test)
    newDB.Table("tests").Where("id in (?)", []uint{1,2}).Delete()
}
```

### 事务

```golang
type Test struct {
    ID uint `primary_key`
    Name string `zorm:"type:varchar(32);not null"`
    Out string `zorm:"type:varchar(128);not null"`
}
func main(){
    newDB, err := zorm.Open("mysql", "user:password@/database?charset=utf8&parseTime=true&loc=Asia%2FShanghai")
    tx := newDB.Begin()
    test = Test{
        Name:"test1",
        Out:"This is a test",
    }
    tx.Where("id = ?", 1).Find(&test)
    tx.Where("id = ?", 1).Delete()
    tx.Where("id in (?)", []uint{2, 3, 4}).Update(test)
    if err := tx.Commit();err != nil {
        tx.Rollback()
    }
}
```

### 假删除

```golang
type Test struct {
    ID uint `primary_key`
    Name string `zorm:"type:varchar(32);not null"`
    Out string `zorm:"type:varchar(128);not null"`
    CreatedAt time.Time
    DeletedAt time.Time
    DeletedAt *time.Time `sql:"index"`
}
func main(){
    newDB, err := zorm.Open("mysql", "user:password@/database?charset=utf8&parseTime=true&loc=Asia%2FShanghai")
    queryMap := map[string]interface{}{"id":1}
    test = Test{
        Name:"test1",
        Out:"This is a test",
    }
    newDB.Unscoped().Where(queryMap).Find(&test) //使用Unscoped()时,将按照DeletedAt是否为NULL,来判断
    newDB.Unscoped().Where(queryMap).Delete()
    newDB.Unscoped().Where(queryMap).Update(&test)
}
```