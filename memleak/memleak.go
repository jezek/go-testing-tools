package memleak

import (
	"bytes"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Leaks monitor

type goroutine struct {
	id    int
	name  string
	stack []byte
}

type leaks struct {
	name       string
	goroutines map[int]goroutine
	report     []*leaks
}

func leaksMonitor(name string, monitors ...*leaks) *leaks {
	return &leaks{
		name,
		leaks{}.collectGoroutines(),
		monitors,
	}
}

// ispired by https://golang.org/src/runtime/debug/stack.go?s=587:606#L21
// stack returns a formatted stack trace of all goroutines.
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
func (_ leaks) stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func (l leaks) collectGoroutines() map[int]goroutine {
	res := make(map[int]goroutine)
	stacks := bytes.Split(l.stack(), []byte{'\n', '\n'})

	regexpId := regexp.MustCompile(`^\s*goroutine\s*(\d+)`)
	for _, st := range stacks {
		lines := bytes.Split(st, []byte{'\n'})
		if len(lines) < 2 {
			panic("routine stach has less tnan two lines: " + string(st))
		}

		idMatches := regexpId.FindSubmatch(lines[0])
		if len(idMatches) < 2 {
			panic("no id found in goroutine stack's first line: " + string(lines[0]))
		}
		id, err := strconv.Atoi(string(idMatches[1]))
		if err != nil {
			panic("converting goroutine id to number error: " + err.Error())
		}
		if _, ok := res[id]; ok {
			panic("2 goroutines with same id: " + strconv.Itoa(id))
		}
		name := strings.TrimSpace(string(lines[1]))

		//filter out our stack routine
		if strings.Contains(name, "xgb.leaks.stack") {
			continue
		}

		res[id] = goroutine{id, name, st}
	}
	return res
}

func (l leaks) leakingGoroutines() []goroutine {
	goroutines := l.collectGoroutines()
	res := []goroutine{}
	for id, gr := range goroutines {
		if _, ok := l.goroutines[id]; ok {
			continue
		}
		res = append(res, gr)
	}
	return res
}
func (l leaks) checkTesting(t *testing.T) {
	if len(l.leakingGoroutines()) == 0 {
		return
	}
	leakTimeout := 10 * time.Millisecond
	time.Sleep(leakTimeout)
	//t.Logf("possible goroutine leakage, waiting %v", leakTimeout)
	grs := l.leakingGoroutines()
	for _, gr := range grs {
		t.Errorf("%s: %s is leaking", l.name, gr.name)
		//t.Errorf("%s: %s is leaking\n%v", l.name, gr.name, string(gr.stack))
	}
	for _, rl := range l.report {
		rl.ignoreLeak(grs...)
	}
}
func (l *leaks) ignoreLeak(grs ...goroutine) {
	for _, gr := range grs {
		l.goroutines[gr.id] = gr
	}
}
