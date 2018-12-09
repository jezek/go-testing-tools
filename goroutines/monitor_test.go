package goroutines

import (
	"sync"
	"testing"
)

func TestMonitor(t *testing.T) {
	step := 0
	lm0 := NewMonitor()
	if lgrs, want := lm0.LeakingGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d. len(lm0.LeakingGoroutines) = %d, want %d", step, len(lgrs), want)
	}
	if lgrs, want := lm0.LostGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm0.LostGoroutines) = %d, want %d", step, len(lgrs), want)
	}

	done := make(chan struct{})
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		<-done
		wg.Done()
	}()
	step++

	if lgrs, want := lm0.LeakingGoroutines(), 1; len(lgrs) != want {
		t.Errorf("step %d, len(lm0.LeakingGoroutines) = %d, want %d", step, len(lgrs), want)
	}
	if lgrs, want := lm0.LostGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm0.LostGoroutines) = %d, want %d", step, len(lgrs), want)
	}

	lm1 := NewMonitor()
	if lgrs, want := lm1.LeakingGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm1.LeakingGoroutines) = %d, want %d", step, len(lgrs), want)
	}
	if lgrs, want := lm1.LostGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm1.LostGoroutines) = %d, want %d", step, len(lgrs), want)
	}

	wg.Add(1)
	go func() {
		<-done
		wg.Done()
	}()
	step++

	if lgrs, want := lm0.LeakingGoroutines(), 2; len(lgrs) != want {
		t.Errorf("step %d, len(lm0.LeakingGoroutines) = %d, want %d", step, len(lgrs), want)
	}
	if lgrs, want := lm0.LostGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm0.LostGoroutines) = %d, want %d", step, len(lgrs), want)
	}

	if lgrs, want := lm1.LeakingGoroutines(), 1; len(lgrs) != want {
		t.Errorf("step %d, len(lm1.LeakingGoroutines) = %d, want %d", step, len(lgrs), want)
	}
	if lgrs, want := lm1.LostGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm1.LostGoroutines) = %d, want %d", step, len(lgrs), want)
	}

	close(done)
	wg.Wait()
	step++

	if lgrs, want := lm0.LeakingGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm0.LeakingGoroutines) = %d, want %d", step, len(lgrs), want)
	}
	if lgrs, want := lm0.LostGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm0.LostGoroutines) = %d, want %d", step, len(lgrs), want)
	}

	if lgrs, want := lm1.LeakingGoroutines(), 0; len(lgrs) != want {
		t.Errorf("step %d, len(lm1.LeakingGoroutines) = %d, want %d", step, len(lgrs), want)
	}
	if lgrs, want := lm1.LostGoroutines(), 1; len(lgrs) != want {
		t.Errorf("step %d, len(lm1.LostGoroutines) = %d, want %d", step, len(lgrs), want)
	}

	lm0.TestingErrorLeaking(t, "lm0")
	//TODO multiple leak monitors with report ignore tests
}
