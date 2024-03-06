package gormx

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type FKAction string

const (
	FKEmpty    FKAction = ""
	FKCascade  FKAction = "CASCADE"
	FKNoAction FKAction = "NO ACTION"
	FKSetNull  FKAction = "SET NULL"
	FKRestrict FKAction = "RESTRICT"
)

type DDL struct {
	db              *gorm.DB
	schemas         sync.Map // struct Type : Schema
	DefaultOnDelete FKAction
	DefaultOnUpdate FKAction
}

func NewDDL(db *gorm.DB) *DDL {
	return &DDL{
		db:              db,
		DefaultOnDelete: FKCascade,
		DefaultOnUpdate: FKCascade,
	}
}

func (s *DDL) AddTables(tables ...interface{}) {
	for _, v := range tables {
		schema.Parse(v, &s.schemas, s.db.NamingStrategy)
	}
}

func (s *DDL) Range(f func(structType reflect.Type, tableSchema *schema.Schema) bool) {
	s.schemas.Range(func(key, value interface{}) bool {
		return f(key.(reflect.Type), value.(*schema.Schema))
	})
}

//	func (s *DDL) AddFK(table, target interface{}, fk string) {
//		srcSch := s.GetSchema(table)
//		dstSch := s.GetSchema(target)
//		s.AddForeignKey(srcSch.Table, fk, dstSch.Table, dstSch.PrimaryFieldDBNames[0], FKRestrict, FKCascade)
//	}
// func (s *DDL) MakeFKName(table, fkey, target, targetCol string) string {
// 	return fmt.Sprintf("fk_%s.%s_%s.%s", table, fkey, target, targetCol)
// }

func (s *DDL) GetTableName(tableStruct interface{}) string {
	stmt := &gorm.Statement{DB: s.db}
	stmt.Parse(tableStruct)
	return stmt.Schema.Table
}

func (s *DDL) GetTablePK(tableStruct interface{}) string {
	stmt := &gorm.Statement{DB: s.db}
	stmt.Parse(tableStruct)
	return stmt.Schema.PrimaryFieldDBNames[0]
}

func (s *DDL) AddFKs(table interface{}) {
	// stmt := &gorm.Statement{DB: s.db}
	sch, err := schema.Parse(table, &s.schemas, s.db.NamingStrategy)
	log.Println(sch, err)
	for _, f := range sch.Fields {
		log.Println(f.TagSettings)
	}
}
func (s *DDL) MakeFKs() {
	s.Range(func(structType reflect.Type, src *schema.Schema) bool {
		for _, f := range src.Fields {
			v, ok := f.TagSettings["fk"]
			if !ok {
				v, ok = f.TagSettings["FK"]
			}
			if ok {
				fkInfo := s.ParseFKInfo(v)
				dropFKSql := fkInfo.DropFKSql(src.Table, f.DBName)
				fkSql := fkInfo.FKSql(src.Table, f.DBName)
				log.Println(dropFKSql)
				s.db.Exec(dropFKSql)
				log.Println(fkSql)
				err := s.db.Exec(fkSql).Error
				if err != nil {
					log.Println(err)
				}
			}
		}
		return true
	})
}

func (s *DDL) ForeignKeyCheck(enable bool) error {
	if enable {
		return s.db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
	}
	return s.db.Exec("SET FOREIGN_KEY_CHECKS=0").Error
}

func (s *DDL) MatchTableName(structType reflect.Type, tableName string) bool {
	return strings.EqualFold(tableName, structType.Name())
}

func (s *DDL) GetSchemaByStructName(structName string) *schema.Schema {
	var r *schema.Schema
	s.Range(func(structType reflect.Type, tableSchema *schema.Schema) bool {
		if s.MatchTableName(structType, structName) {
			r = tableSchema
			return false
		}
		return true
	})
	return r
}
func (s *DDL) GetSchema(obj interface{}) *schema.Schema {
	r, _ := s.schemas.Load(reflect.Indirect(reflect.ValueOf(obj)).Type())
	return r.(*schema.Schema)
}

type FKInfo struct {
	Table    string
	Field    string
	OnDelete FKAction
	OnUpdate FKAction
}

func (s *FKInfo) FKName(srcTable, srcField string) string {
	return fmt.Sprintf("fk_%s_%s", srcTable, srcField)
}

func (s *FKInfo) DropFKSql(srcTable, srcField string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", srcTable, s.FKName(srcTable, srcField))
}
func (s *FKInfo) FKSql(srcTable, srcField string) string {
	fkName := s.FKName(srcTable, srcField)
	fkSql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)", srcTable, fkName, srcField, s.Table, s.Field)
	if s.OnDelete != FKEmpty {
		fkSql += " ON DELETE " + string(s.OnDelete)
	}
	if s.OnUpdate != FKEmpty {
		fkSql += " ON UPDATE " + string(s.OnUpdate)
	}
	return fkSql
}

// tag:
// eg1. fk:User.Id,ondelete=CASCADE,onupdate=CASCADE  -> ALTER TABLE User ADD CONSTRAINT fkname FOREIGN KEY (id) REFERENCES on delete CASCADE,on update CASCADE
// eg2. fk:User,ondelete=SET NULL,onupdate=CASCADE  -> ALTER TABLE User ADD CONSTRAINT fkname FOREIGN KEY (id) REFERENCES on delete CASCADE,on update CASCADE
func (s *DDL) ParseFKInfo(tag string) FKInfo {
	parts := strings.Split(tag, ",")
	r := FKInfo{}
	structNameAndField := strings.Split(parts[0], ".")
	r.Table = s.db.NamingStrategy.TableName(structNameAndField[0])
	if len(structNameAndField) > 1 {
		r.Field = s.db.NamingStrategy.ColumnName(r.Table, structNameAndField[1])
	} else {
		r.Field = s.GetSchemaByStructName(structNameAndField[0]).PrimaryFieldDBNames[0]
	}

	for i := 1; i < len(parts); i++ {
		kv := strings.Split(parts[i], "=")
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		value := strings.ToUpper(strings.TrimSpace(kv[1]))
		switch key {
		case "ondelete":
			r.OnDelete = FKAction(value)
		case "onupdate":
			r.OnUpdate = FKAction(value)
		}
	}
	return r

}
