package pgbase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	//StatusOk - запрос принят начата обработка (данных нет в ответе)
	StatusOk = "ok"
	//StatusData - очереная порция данных
	StatusData = "data"
	//StatusEnd - данные закончились (данных нет в ответе)
	StatusEnd = "end"
	//StatusError - какая то ошибка в запросе (данных нет в ответе)
	StatusError = "error"
)

//Uses все открытые рабочие области
var Uses map[string]*WorkArea

//MutexUses мутекс для рабочих областей
var MutexUses sync.Mutex

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

//WorkArea одна рабочая область
type WorkArea struct {
	Ready      bool
	ID         int64
	RemoteAddr string
	Query      UserQuery
	W          http.ResponseWriter
}

//UserResponce стандартный ответ
type UserResponce struct {
	ID      int64  `json:"id"`
	Status  string `json:"status"`
	Headers []Head `json:"headers,omitempty"`
	Datas   []Data `json:"datas,omitempty"`
}

//Head имена и типы для ответа
type Head struct {
	Name   string `json:"name"`
	IsBool bool   `json:"isbool"`
}

//Data собственно данные для ответа
type Data struct {
	Time   time.Time `json:"tm"`
	Values []float32 `json:"vals"`
}

//SendResponce правильный возрат ответа на запрос
func SendResponce(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(res)
}

//IsNewQuery проверяет есть для данного запроса запущеный сбор данных
func IsNewQuery(w http.ResponseWriter, r *http.Request) (*WorkArea, error) {
	wa := new(WorkArea)
	_, ok := Uses[r.RemoteAddr]
	if !ok {
		wa.RemoteAddr = r.RemoteAddr
		wa.Ready = true
	} else {
		wa = Uses[r.RemoteAddr]
	}
	if !wa.Ready {
		//Этому пользователю еще отгружают данные
		return nil, fmt.Errorf("Wait while sending data")
	}
	s := r.URL.Query().Get("query")
	var uq UserQuery
	err := json.Unmarshal([]byte(s), &uq)
	if err != nil {
		return nil, fmt.Errorf("Error query %s", err.Error())
	}
	wa.Query = uq
	wa.W = w
	MutexUses.Lock()
	Uses[r.RemoteAddr] = wa
	MutexUses.Unlock()
	return wa, nil
}
func (wa *WorkArea) send(status string) {
	var ur UserResponce
	ur.ID = wa.ID
	ur.Status = status
	r, _ := json.Marshal(&ur)
	SendResponce(wa.W, r)
}

//SendOk посылает ок пользователю
func (wa *WorkArea) SendOk() {
	wa.send(StatusOk)
}

//SendEnd send End status
func (wa *WorkArea) SendEnd() {
	wa.send(StatusEnd)
}

//SendError send Err status
func (wa *WorkArea) SendError() {
	wa.send(StatusError)
}
