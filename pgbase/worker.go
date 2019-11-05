package pgbase

import (
	"database/sql"
	"fmt"
	"rura/teprol/logger"
)

//Monitor Управляет workers
func (cdb *CtrlDataBase) Monitor(stop, ret chan int) {

	select {
	case <-stop{
		logger.Info.Printf("Database stoped %s ", cdb.BaseData.Name)
		ret <- 1
	
	}
	case <-cdb.StopAll{
		
	}
}
}

//CloseAll закрывает все соединения и отключает
func (cdb *CtrlDataBase) CloseAll() {
	for i := 0; i < cdb.BaseData.OpenCount; i++ {
		cdb.StopAll <- 1
	}
	cdb.StopAll <- 0

	cdb.Mutex.Lock()
	defer cdb.Mutex.Unlock()
	for _, dbc := range cdb.MapConnect {
		dbc.Close()
	}
	cdb.BaseData.Connect = false
	if cdb.IsWork {
		cdb.DoWorkArea.SendEnd()
	}
	cdb.IsWork = false
	cdb.BaseData.Connect = false
}

//StartWorkers запускает читателей для этой базы данных
func (cdb *CtrlDataBase) StartWorkers(stop, ret chan int) error {
	cdb.MapConnect = make(map[int]*sql.DB)
	cdb.InChan = make(chan *WorkArea)
	cdb.OutChan = make(chan UserResponce)
	cdb.StopAll = make(chan int)
	cdb.IsWork = false
	if !cdb.BaseData.Connect {
		err := fmt.Errorf("Database not connected %s", cdb.BaseData.Name)
		return err
	}
	for i := 0; i < cdb.BaseData.OpenCount; i++ {
		con, err := sql.Open("postgres", cdb.StrConnect)
		if err != nil {
			cdb.CloseAll()
			return err
		}
		cdb.MapConnect[i] = con
	}
	go cdb.Monitor(stop, ret)
	return nil
}
