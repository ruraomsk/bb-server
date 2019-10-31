package pgbase

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"rura/teprol/logger"
	"sort"
)

//ListDataBases Перечень загруженных баз данных
var ListDataBases []*DataBase

//MapDataBases Управляющие структуры
var MapDataBases map[string]*CtrlDataBase

//LoadDataBases загрузка описания баз данных
func LoadDataBases(path string) error {
	var databases []*DataBase
	Uses = make(map[string]*WorkArea)
	ListDataBases = make([]*DataBase, 0)
	MapDataBases = make(map[string]*CtrlDataBase)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Printf("Error reading file %s! %s", path, err.Error())
		return err
	}
	err = json.Unmarshal(buf, &databases)
	if err != nil {
		return err
	}
	for _, db := range databases {
		s := db.LoadDataBase()
		ListDataBases = append(ListDataBases, db)
		cdb := new(CtrlDataBase)
		cdb.BaseData = db
		cdb.StrConnect = s
		MapDataBases[db.Name] = cdb
	}
	return nil
}

//LoadDataBase Загружает описание одной базы данных
func (db *DataBase) LoadDataBase() (StrConnect string) {
	StrConnect = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db.Host, db.Port, db.User, db.Password, db.DBname)
	con, err := sql.Open("postgres", StrConnect)
	if err != nil {
		logger.Error.Printf("Open dataBase %s error %s", StrConnect, err.Error())
		db.Connect = false
		return
	}
	if err = con.Ping(); err != nil {
		logger.Error.Printf("Not ping %s error %s", StrConnect, err.Error())
		db.Connect = false
		return
	}
	db.Connect = true
	rows, err := con.Query("SELECT table_name FROM information_schema.tables WHERE table_schema NOT IN ('information_schema','pg_catalog');")
	if err != nil {
		logger.Error.Printf("Query for name table error %s", err.Error())
		return
	}
	defer rows.Close()
	db.Tables = make([]Table, 0)
	for rows.Next() {
		table := new(Table)
		table.Variables = make([]Variable, 0)
		err = rows.Scan(&table.Name)
		if err != nil {
			logger.Error.Printf("Scan for name table error %s", err.Error())
			return
		}
		str := "SELECT subq.attname, d.description FROM(SELECT a.attname, c.oid, a.attnum FROM pg_class c, pg_attribute a WHERE c.oid = a.attrelid AND c.relname = '" + table.Name + "'AND a.attnum > 0) subq LEFT OUTER JOIN pg_description d ON(d.objsubid = subq.attnum AND d.objoid = subq.oid);"
		rnames, err := con.Query(str)
		if err != nil {
			logger.Error.Printf("Query for name variables for table %s error %s", table.Name, err.Error())
			return
		}
		defer rnames.Close()
		isTm := false
		for rnames.Next() {
			var name, desc string
			// var desc string
			v := new(Variable)
			err := rnames.Scan(&name, &desc)
			if err != nil {
				if name != "tm" {
					logger.Error.Printf("Scan for name variables error %s", err.Error())
					desc = ""
				} else {
					desc = "Метка времени"
				}
			}
			v.Name = name
			v.Description = desc
			if v.Name == "tm" {
				isTm = true
			}
			table.Variables = append(table.Variables, *v)
		}
		rnames.Close()
		if isTm {
			db.Tables = append(db.Tables, *table)
		}
	}
	rows.Close()
	for ii := range db.Tables {
		sort.Slice(db.Tables[ii].Variables, func(i, j int) bool {
			return db.Tables[ii].Variables[i].Name < db.Tables[ii].Variables[j].Name
		})
	}
	return
}
