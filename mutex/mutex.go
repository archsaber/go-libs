package mutex

import (
	"errors"
	"log"
	"time"
)

// simple mutex with lock try timeout & multiple lock capacity

// Mutex Simple mutex with timeout
type Mutex struct {
	c chan bool
}

// NewMutex give a new Mutex initailized with timeout and having upto lockCap concurrent locks
func NewMutex(lockCap int) (*Mutex, error) {
	if lockCap < 1 {
		return nil, errors.New("Invalicd lock capcity, only greter than 0 values allowed")
	}

	return &Mutex{make(chan bool, lockCap)}, nil
}

// Lock impl
func (m *Mutex) Lock() {
	m.c <- true
}

// Unlock impl
func (m *Mutex) Unlock() {
	if m.Cap() < 1 {
		log.Println("ERROR in archsaber/go-libs/mutex : unlock called with out taking the lock")
		return
	}
	<-m.c
}

// Cap returns current capacity
func (m *Mutex) Cap() int {
	return len(m.c)
}

// TryLock try to get a lock within the given timeout and return true, else return false
func (m *Mutex) TryLock(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	var result bool

	select {
	case m.c <- true:
		timer.Stop()
		result = true
	case <-time.After(timeout):
		result = false
	}

	return result
}
