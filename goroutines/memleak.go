package goroutines

import (
	"bytes"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Goroutine struct {
	ID    int
	Name  string
	Stack []byte
}

type Monitor struct {
	goroutines map[int]Goroutine
	report     []*Monitor
}

func NewMonitor(monitors ...*Monitor) *Monitor {
	return &Monitor{
		Monitor{}.collectGoroutines(),
		monitors,
	}
}

// inspired by https://golang.org/src/runtime/debug/stack.go?s=587:606#L21
// stack returns a formatted stack trace of all goroutines.
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
func (_ Monitor) stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func (l Monitor) collectGoroutines() map[int]Goroutine {
	res := make(map[int]Goroutine)
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

		res[id] = Goroutine{id, name, st}
	}
	return res
}

func (l Monitor) LeakingGoroutines() map[int]Goroutine {
	goroutines := l.collectGoroutines()
	res := map[int]Goroutine{}
	for id, gr := range goroutines {
		if _, ok := l.goroutines[id]; ok {
			continue
		}
		res[id] = gr
	}
	return res
}

func (l Monitor) LostGoroutines() map[int]Goroutine {
	goroutines := l.collectGoroutines()
	res := map[int]Goroutine{}
	for id, gr := range l.goroutines {
		if _, ok := goroutines[id]; ok {
			continue
		}
		res[id] = gr
	}
	return res
}

func (l Monitor) TestingErrorLeaking(t *testing.T, desc string) {
	if len(l.LeakingGoroutines()) == 0 {
		return
	}
	leakTimeout := 10 * time.Millisecond
	time.Sleep(leakTimeout)
	//t.Logf("possible goroutine leakage, waiting %v", leakTimeout)
	grs := l.LeakingGoroutines()
	for _, gr := range grs {
		t.Errorf("%s: %s is leaking", desc, gr.Name)
	}
	for _, rl := range l.report {
		grsa := make([]Goroutine, 0, len(grs))
		for _, gr := range grs {
			grsa = append(grsa, gr)
		}
		rl.ignoreLeak(grsa...)
	}
}
func (l *Monitor) ignoreLeak(grs ...Goroutine) {
	for _, gr := range grs {
		l.goroutines[gr.ID] = gr
	}
}
