package pkg

import (
	"fmt"
	"time"
)

type Clock struct {
	Remaining time.Duration
	Increment time.Duration
	Paused    bool
}

func (cl *Clock) String() string {
	return fmt.Sprintf("%d:%02d", int(cl.Remaining.Minutes()), int(cl.Remaining.Seconds())%60)
}

func NewClock(duration, increment time.Duration) *Clock {
	cl := &Clock{
		Remaining: duration,
		Increment: increment,
		Paused:    true,
	}
	go cl.run()
	return cl
}

func (cl *Clock) run() {
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			if !cl.Paused {
				cl.Remaining -= time.Second
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
