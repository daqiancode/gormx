# gormx



## Example:
```go
import (
	"fmt"
	"testing"

	"github.com/daqiancode/gormx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	Id   int64 `gorm:"primaryKey"`
	Name string
}
type Product struct {
	Id   int64 `gorm:"primaryKey"`
	Name string
}
type UserProduct struct {
	Id        int64 `gorm:"primaryKey"`
	Uid       int64 `gorm:"index;fk:User,ondelete=SET NULL,onupdate=CASCADE;"`
	ProductId int64 `gorm:"index;fk:Product.Id"`
}

// go clean -testcache
func TestMakeFK(t *testing.T) {
	conUrl := "root:123456@tcp(localhost:3306)/gormx_test?charset=utf8&parseTime=True&loc=Local"
	err := gormx.CreateDB("mysql", conUrl)
	if err != nil {
		fmt.Println(err)
	}
	// defer gormx.DropDB("mysql", conUrl)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: conUrl,
	}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	tables := []interface{}{&User{}, &Product{}, &UserProduct{}}
	ddl := gormx.NewDDL(db)
	err = db.AutoMigrate(tables...)
	if err != nil {
		t.Fatal(err)
	}
	ddl.AddTables(tables...)
	ddl.MakeFKs()
}
```