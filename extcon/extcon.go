package extcon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

//ExtContext расширенный контекст
type ExtContext struct {
	name       string
	isExpTime  bool
	expTime    time.Time
	isTimeout  bool
	timeout    time.Duration
	ctx        context.Context
	cancelFunc context.CancelFunc
	canceled   bool
	executed   bool
	status     string
}

var id uint64
var mutexID sync.Mutex

var mutex sync.Mutex
var work bool

//Contexts собраны все контексты
var contexts map[uint64]*ExtContext

func newID() uint64 {
	mutexID.Lock()
	defer mutexID.Unlock()
	id++
	return id
}

//NewContext создает новый расширенный контекст только с командой завершения
func NewContext(name string) (*ExtContext, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if !work {
		return nil, fmt.Errorf("stoped context system")
	}
	var ec ExtContext
	ec.name = name
	ctx, cancel := context.WithCancel(context.Background())
	ec.cancelFunc = cancel
	ec.ctx = ctx
	id := newID()
	contexts[id] = &ec
	return &ec, nil
}

//GetName return name context
func (ec *ExtContext) GetName() string {
	return ec.name
}

//GetStatus return string of status
func (ec *ExtContext) GetStatus() string {
	return ec.status
}

// Executed set status executed to true
func (ec *ExtContext) Executed() {
	mutex.Lock()
	defer mutex.Unlock()
	ec.executed = true
}

// IsExecuted return status executed
func (ec *ExtContext) IsExecuted() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return ec.executed
}

//Done rerurn chan for cancel
func (ec *ExtContext) Done() <-chan struct{} {
	return ec.ctx.Done()
}

//SetDeadLine устанавливает время дедлайна для контекста
func (ec *ExtContext) SetDeadLine(dt time.Time) {
	mutex.Lock()
	defer mutex.Unlock()
	if !work {
		return
	}
	ec.isExpTime = true
	ec.expTime = dt
}

//SetTimeOut устанавливает время таймаута для контекста
func (ec *ExtContext) SetTimeOut(dt time.Duration) {
	mutex.Lock()
	defer mutex.Unlock()
	if !work {
		return
	}
	ec.isTimeout = true
	ec.timeout = dt
}

//Cancel завершает контекст
func (ec *ExtContext) Cancel() {
	mutex.Lock()
	defer mutex.Unlock()
	ec.status = "cancel"
	ec.cancelF()
}

func (ec *ExtContext) cancelF() {
	ec.canceled = true
	ec.cancelFunc()
}

//BackgroundInit инициализируем
func BackgroundInit() {
	id = 0
	work = true
	contexts = make(map[uint64]*ExtContext)
}

func allstop(status string) {
	mutex.Lock()
	work = false
	for _, ec := range contexts {
		if ec.canceled {
			continue
		}
		ec.status = status
		ec.cancelF()
	}
	mutex.Unlock()
	all := make(chan int)
	go func() {
		time.Sleep(10 * time.Second)
		all <- 1
	}()
	timer := make(chan int)
	go func() {
		for true {
			time.Sleep(100 * time.Millisecond)
			timer <- 1
		}
	}()
	for true {
		select {
		case <-all:
			{
				return
			}
		case <-timer:
			{
				count := 0
				for _, ec := range contexts {
					if !ec.executed {
						count++
					}
				}
				if count == 0 {
					return
				}
			}
		}
	}

}

//BackgroundWork обычно вызывается для обслуживания разного
// ПОсле выхода нет контекстов
func BackgroundWork(step time.Duration, stop chan int) {
	if !work {
		BackgroundInit()
	}
	timer := make(chan int)
	go func() {
		for true {
			time.Sleep(step)
			timer <- 1
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for true {
		start := time.Now()
		select {
		case <-stop:
			{
				allstop("stop")
				return
			}
		case <-c:
			{
				allstop("kill")
				return
			}
		case <-timer:
			{
				duration := time.Now().Sub(start)
				mutex.Lock()
				for _, ec := range contexts {
					if ec.isExpTime {
						if ec.expTime.Before(time.Now()) && !ec.canceled {
							ec.status = "deadline"
							ec.cancelF()
						}
					}
					if ec.isTimeout {
						ec.timeout = ec.timeout - duration
						if ec.timeout <= 0 && !ec.canceled {
							ec.status = "timeout"
							ec.cancelF()
						}
					}
				}
				mutex.Unlock()
			}
		}
	}
}
