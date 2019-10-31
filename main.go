package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"rura/bb-server/pgbase"
	"rura/teprol/logger"

	"time"
)

var databases []*pgbase.DataBase
var err error
var stop chan int
var ret chan int
var timer chan string

func sendResponce(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(res)
}
func respAllDataBases(w http.ResponseWriter, r *http.Request) {
	resjson, err := json.Marshal(&databases)
	if err != nil {
		logger.Error.Printf("Error from %s %s", "/all", err.Error())
		return
	}
	sendResponce(w, resjson)
}
func respData(w http.ResponseWriter, r *http.Request) {
	query := []byte(r.URL.Query().Get("query"))
	fmt.Println(">", r.URL.Query().Get("query"), "<")
	var uq pgbase.UserQuery
	// b, _ := json.Marshal(uq)
	// fmt.Println(">", string(b), "<")

	err := json.Unmarshal(query, &uq)
	if err != nil {
		logger.Error.Printf("Error from %s %s", "/get", err.Error())
		sendResponce(w, []byte("Bad"))
		return
	}
	sendResponce(w, []byte("Ok"))
}
func gui() {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/all", respAllDataBases)
	http.HandleFunc("/get", respData)
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
	databases, err = pgbase.LoadDataBases("setup.json")
	if err != nil {
		fmt.Println("Error loading setup ", err.Error())
		return
	}
	stop = make(chan int, 100)
	ret = make(chan int, 1)
	count := 0
	for _, db := range databases {
		if db.Connect {
			err = db.StartWorkers(stop, ret)
			if err == nil {
				count++
			}
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
				for _, db := range databases {
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
				for _, db := range databases {
					if !db.Connect {
						logger.Info.Println("Reconect db ", db.Name)

						err = db.LoadDataBase()
						if err == nil {
							count++
						}
					}
				}
			}
		}

	}
	fmt.Println("Exit working...")
	logger.Info.Println("Exit working...")
}
