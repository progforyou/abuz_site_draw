package axscheduler

import (
	"time"
)

type AxScheduler struct {
	exit     chan bool
	method   func()
	duration time.Duration
}

func NewScheduler(duration time.Duration, execute func()) *AxScheduler {
	return &AxScheduler{
		method:   execute,
		duration: duration,
	}
}

func (a *AxScheduler) Start() {
	if a.exit != nil {
		return
	}
	a.exit = make(chan bool, 1)
	go a.execute()
}

func (a *AxScheduler) Stop() {
	if a.exit == nil {
		return
	}
	a.exit <- true
}

func (a *AxScheduler) execute() {
	a.method()
	c1 := time.Tick(a.duration)
	for {
		select {
		case <-c1:
			a.method()
		case <-a.exit: //Exit thread
			close(a.exit)
			return
		}
	}
}
