package pgbase

import (
	"database/sql"
	"sync"
)

//CtrlDataBase Управление соединением
type CtrlDataBase struct {
	BaseData   *DataBase
	Mutex      sync.Mutex
	InChan     chan *WorkArea    //Принимаем запросы на исполнение
	StopAll    chan int          //остановится всем workers
	OutChan    chan UserResponce //Отправляем ответы
	DoWorkArea *WorkArea         // Текущий исполняемый запрос
	MapConnect map[int]*sql.DB
	IsWork     bool
	StrConnect string
}

//DataBase структура для подключения базы данных
type DataBase struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Host        string  `json:"host"`
	Port        int     `json:"port"`
	User        string  `json:"user"`
	Password    string  `json:"password"`
	DBname      string  `json:"dbname"`
	Connect     bool    `json:"connected"`
	Step        int     `json:"step"`
	OpenCount   int     `json:"count"`
	Tables      []Table `json:"tables"`
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
