package proc

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	seq   uint32
	lock  sync.Mutex
	lns   []*listener
	mask  uint32
	sigch chan os.Signal
)

func registerSignalListener() {
	// https://golang.org/pkg/os/signal/#Notify
	sigch = make(chan os.Signal, 1)

	// https://colobu.com/2015/10/09/Linux-Signals/
	signal.Notify(
		sigch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		for {
			sig := <-sigch
			debugf(fmt.Sprintf("PID: %d. Received %v.", pid, sig))
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				// gracefully shuts down the process.
				Shutdown(syscall.SIGTERM)
				signal.Stop(sigch)
				os.Exit(0)
			default:
				if !Notify(sig) {
					debugf("PID %d. Got unregistered signal: %v.", pid, sig)
				}
			}
		}
	}()
}

// the max across all systems
// see go/src/os/signal/signal.go
const numSig = 65

func signum(sig os.Signal) int {
	switch sig := sig.(type) {
	case syscall.Signal:
		i := int(sig)
		if i < 0 || i >= numSig {
			return -1
		}
		return i
	default:
		return -1
	}
}

type listener struct {
	id   uint32
	fn   func()
	sig  int
	once bool
}

func add(sig os.Signal, fn func(), once bool) uint32 {
	if n := signum(sig); n > -1 {
		lock.Lock()
		defer lock.Unlock()

		// see go/src/os/signal/signal.go
		if (mask>>uint(n&31))&1 == 0 {
			mask |= 1 << uint(n&31)
			signal.Notify(sigch, sig)
		}

		id := atomic.AddUint32(&seq, 1)
		lns = append(lns, &listener{
			id:   id,
			fn:   wrap(fn, once),
			sig:  n,
			once: once,
		})
		return id
	}
	return 0
}

func wrap(fn func(), once bool) func() {
	if !once {
		return fn
	}
	var so sync.Once
	return func() { so.Do(fn) }
}

func On(sig os.Signal, fn func()) uint32 {
	return add(sig, fn, false)
}

func Once(sig os.Signal, fn func()) uint32 {
	return add(sig, fn, true)
}

func Cancel(ids ...uint32) {
	n := len(ids)
	for _, id := range ids {
		if id == 0 {
			n--
		}
	}
	if n == 0 {
		return
	}
	lock.Lock()
	lns = slices.DeleteFunc(lns, func(l *listener) bool {
		return slices.Contains(ids, l.id)
	})
	lock.Unlock()
}

func Notify(sig os.Signal) bool {
	n := signum(sig)
	if n == -1 {
		return false
	}

	lock.Lock()
	l := len(lns)
	fs := make([]func(), 0, l)

	for i := l - 1; i >= 0; i-- {
		if l := lns[i]; l.sig == n {
			fs = append(fs, l.fn)
			if l.once {
				lns = slices.Delete(lns, i, i+1)
			}
		}
	}
	lock.Unlock()

	if len(fs) == 0 {
		return false
	}

	var wg sync.WaitGroup
	var run = safeRunner(&wg)
	for _, fn := range fs {
		if fn != nil {
			run(fn)
		}
	}
	wg.Wait()

	return true
}

func safeRunner(wg *sync.WaitGroup) func(func()) {
	return func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer recovery()
			fn()
		}()
	}
}

func recovery() {
	if p := recover(); p != nil {
		debugf("%+v\n%s", p, debug.Stack())
	}
}
