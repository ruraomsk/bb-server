package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"rura/bb-server/pgbase"
	"rura/teprol/logger"
	"sync"

	"time"

	_ "github.com/lib/pq"
)

var err error
var stop chan int
var ret chan int
var timer chan string
var mutex sync.Mutex

func respAllDataBases(w http.ResponseWriter, r *http.Request) {
	resjson, err := json.Marshal(&pgbase.ListDataBases)
	if err != nil {
		logger.Error.Printf("Error from %s %s", "/all", err.Error())
		return
	}
	pgbase.SendResponce(w, resjson)
}
func respGet(w http.ResponseWriter, r *http.Request) {
	uwa, err := pgbase.IsNewQuery(w, r)
	if err != nil {
		logger.Error.Printf("Doublicate  %s %s", r.RemoteAddr, err.Error())
		w.WriteHeader(http.StatusNoContent)
		return
	}
	for _, cdb := range pgbase.MapDataBases {
		if uwa.Query.DBName == cdb.BaseData.Name {
			uwa.SendOk()
			if cdb.BaseData.Connect {
				//ставим на передачу
				cdb.InChan <- uwa
			} else {
				uwa.SendEnd()
			}
			return
		}
	}
	uwa.SendError()
}
func gui() {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/all", respAllDataBases)
	http.HandleFunc("/get", respGet)
	logger.Info.Println("Listering on port 8080")
	fmt.Println("Listering on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err.Error())
	}

}
func sleeping() {
	for true {
		time.Sleep(10 * time.Second)
		timer <- "reconnect"
	}
}
func main() {
	err := logger.Init("/home/rura/log/bb-server")
	if err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
	logger.Info.Println("Start work...")
	fmt.Println("Start work...")
	err = pgbase.LoadDataBases("setup.json")
	if err != nil {
		fmt.Println("Error loading setup ", err.Error())
		return
	}

	stop = make(chan int, 100)
	ret = make(chan int, 1)
	count := 0
	for _, cdb := range pgbase.MapDataBases {
		if cdb.BaseData.Connect {
			cdb.StartWorkers(stop, ret)
			count++
		}
	}
	timer = make(chan string)

	go gui()
	go sleeping()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	STOP := false
mainCycle:
	for true {
		select {
		case <-c:
			{
				STOP = true
				for _, db := range pgbase.ListDataBases {
					logger.Info.Println("Stoping ", db.Name)
					stop <- 1
				}
				continue
				// time.Sleep(10 * time.Second)
				// break mainCycle
			}
		case <-ret:
			{
				count--
				if count == 0 {
					break mainCycle
				}
			}
		case <-timer:
			{
				if STOP {
					continue
				}
				mutex.Lock()
				count = 0
				for _, db := range pgbase.ListDataBases {
					if !db.Connect {
						logger.Info.Println("Reconect db ", db.Name)
						db.LoadDataBase()
					}
					if db.Connect {
						pgbase.MapDataBases[db.Name].StartWorkers(stop, ret)
						count++
					}
				}
				mutex.Unlock()
			}
		}
	}
	fmt.Println("Exit working...")
	logger.Info.Println("Exit working...")
}
