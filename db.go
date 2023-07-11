package gormx

import (
	"database/sql"
	"log"
	"net/url"
	"regexp"
)

func processConnectionUrl(connectionUrl string) (string, string) {
	p := regexp.MustCompile(`/(\w+)`)
	dbName := ""
	conUrl := p.ReplaceAllStringFunc(connectionUrl, func(s string) string {
		dbName = s[1:]
		return "/"
	})
	return conUrl, dbName
}

func CreateDB(driverName, connectionUrl string) error {
	conUrl, dbName := processConnectionUrl(connectionUrl)
	return CreateDBWithConUrl(driverName, conUrl, dbName)
}
func DropDB(driverName, connectionUrl string) error {
	conUrl, dbName := processConnectionUrl(connectionUrl)
	return DropDBWithConUrl(driverName, conUrl, dbName)
}
func CreateDBWithConUrl(driverName, connectionUrl, dbName string) error {
	log.Printf("Create database %s\n", dbName)
	url.Parse(connectionUrl)
	db, err := sql.Open(driverName, connectionUrl)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE DATABASE " + dbName)
	return err
}

func DropDBWithConUrl(driverName, connectionUrl, dbName string) error {
	log.Printf("Drop database %s\n", dbName)
	url.Parse(connectionUrl)
	db, err := sql.Open(driverName, connectionUrl)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("DROP DATABASE " + dbName)
	return err
}
