package memleak

import (
	"sync"
	"testing"
)

func TestMonitor(t *testing.T) {
	lm := NewMonitor("lm")
	if lgrs := lm.LeakingGoroutines(); len(lgrs) != 0 {
		t.Errorf("leakingGoroutines returned %d leaking goroutines, want 0", len(lgrs))
	}

	done := make(chan struct{})
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		<-done
		wg.Done()
	}()

	if lgrs := lm.LeakingGoroutines(); len(lgrs) != 1 {
		t.Errorf("leakingGoroutines returned %d leaking goroutines, want 1", len(lgrs))
	}

	wg.Add(1)
	go func() {
		<-done
		wg.Done()
	}()

	if lgrs := lm.LeakingGoroutines(); len(lgrs) != 2 {
		t.Errorf("leakingGoroutines returned %d leaking goroutines, want 2", len(lgrs))
	}

	close(done)
	wg.Wait()

	if lgrs := lm.LeakingGoroutines(); len(lgrs) != 0 {
		t.Errorf("leakingGoroutines returned %d leaking goroutines, want 0", len(lgrs))
	}

	lm.CheckTesting(t)
	//TODO multiple leak monitors with report ignore tests
}
