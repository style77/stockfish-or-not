package timer

import (
	"log"
	"sync"
	"time"
)

type Timer struct {
	Duration  int
	IsRunning bool
	IsPaused  bool
	IsOver    bool
	IsStarted bool
	mux       sync.Mutex

	Stop         chan struct{}
	Pause        chan struct{}
	Resume       chan struct{}
	TickCallback func(int)
}

func NewTimer(duration int, tickCallback func(int)) *Timer {
	return &Timer{
		Duration:     duration,
		IsRunning:    false,
		IsPaused:     false,
		IsOver:       false,
		IsStarted:    false,
		Stop:         make(chan struct{}),
		Pause:        make(chan struct{}),
		Resume:       make(chan struct{}),
		TickCallback: tickCallback,
	}
}

func (t *Timer) StartTimer() {
	t.mux.Lock()
	t.IsRunning = true
	t.IsPaused = false
	t.IsOver = false
	t.IsStarted = true
	t.mux.Unlock()

	go func() {
		for {
			t.mux.Lock()
			if t.IsOver {
				t.mux.Unlock()
				break
			}

			if t.IsPaused {
				t.mux.Unlock()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			t.Duration--

			if t.Duration <= 0 {
				t.IsOver = true
				t.IsRunning = false
				t.mux.Unlock()
				t.TickCallback(0)
				log.Println("Timer is over")

				break
			}

			t.TickCallback(t.Duration)
			t.mux.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()
}

func (t *Timer) StopTimer() {
	close(t.Stop)
	t.mux.Lock()
	t.IsRunning = false
	t.IsOver = true
	t.mux.Unlock()
}

func (t *Timer) PauseTimer() {
	t.mux.Lock()
	t.IsPaused = true
	t.mux.Unlock()
}

func (t *Timer) ResumeTimer() {
	t.mux.Lock()
	t.IsPaused = false
	t.mux.Unlock()
}

func (t *Timer) Close() {
	close(t.Stop)
}
