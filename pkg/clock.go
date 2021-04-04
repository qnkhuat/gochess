package pkg

import (
	"fmt"
	"log"
	"time"
)

type Clock struct {
	Duration  time.Duration
	Remaining time.Duration
	Increment time.Duration
	Paused    bool
}

func (cl *Clock) String() string {
	return fmt.Sprintf("%d:%02d", int(cl.Remaining.Minutes()), int(cl.Remaining.Seconds())%60)
}

func NewClock(duration, increment time.Duration) *Clock {
	cl := &Clock{
		Duration:  duration,
		Remaining: duration,
		Increment: increment,
		Paused:    true,
	}
	go cl.Run()
	return cl
}

func (cl *Clock) Run() {
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			if !cl.Paused {
				cl.Remaining -= time.Second
				log.Printf("Logging: %s", cl)
			}
		}
	}
}

func (cl *Clock) Tick() {
	cl.Paused = false
	cl.Remaining += cl.Increment
}

func (cl *Clock) Pause() {
	cl.Paused = true
}

func (cl *Clock) Reset() {
	cl.Remaining = cl.Duration
}
