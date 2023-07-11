package gormx_test

import (
	"testing"

	"github.com/daqiancode/gormx"
	_ "github.com/go-sql-driver/mysql"
)

func TestCreateDB(t *testing.T) {
	dbUrl := "root:123456@tcp(localhost:3306)/test2?charset=utf8&parseTime=True&loc=Local"

	err := gormx.CreateDB("mysql", dbUrl)
	if err != nil {
		t.Error(err)
	}
	err = gormx.DropDB("mysql", dbUrl)
	if err != nil {
		t.Error(err)
	}
}
