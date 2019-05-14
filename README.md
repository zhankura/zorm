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