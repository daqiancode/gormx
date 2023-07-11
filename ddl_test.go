package gormx_test

import (
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
	Uid       int64 `gorm:"index;fk:User;"`
	ProductId int64 `gorm:"index;fk:Product"`
}

func TestMakeFK(t *testing.T) {
	conUrl := "root:123456@tcp(localhost:3306)/testing?charset=utf8&parseTime=True&loc=Local"
	gormx.CreateDB("mysql", conUrl)
	defer gormx.DropDB("mysql", conUrl)
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
