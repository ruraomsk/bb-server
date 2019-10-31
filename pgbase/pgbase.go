package pgbase

import (
	"sync"
	"time"
)

//DataBase структура для подключения базы данных
type DataBase struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Host        string     `json:"host"`
	Port        int        `json:"port"`
	User        string     `json:"user"`
	Password    string     `json:"password"`
	DBname      string     `json:"dbname"`
	Connect     bool       `json:"connected"`
	Step        int        `json:"step"`
	OpenCount   int        `json:"count"`
	StrConnect  string     `json:"-"`
	Mutex       sync.Mutex `json:"-"`
	IsWork      bool       `json:"-"`
	Tables      []Table    `json:"tables"`
}

//Table описание таблицы
type Table struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Variables   []Variable `json:"vars"`
}

//Variable описание одной переменной
type Variable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

//UserQuery запрос прользователя
type UserQuery struct {
	DBName  string    `json:"db"`
	TmStart time.Time `json:"start"`
	TmEnd   time.Time `json:"end"`
	Whos    []Who
}

//Who уточнение считывваемые переменные
type Who struct {
	Tname string `json:"table"`
	Vname string `json:"name"`
}

// type UserResponce struct{
// 	Headers []Head
// 	Data []
// }
