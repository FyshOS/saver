package saver

import "sync"

// pauser is a simple gate that animation goroutines can wait on so they may be
// paused (for example while the system is suspended) and resumed cleanly
// without busy-waiting. A paused goroutine blocks rather than continuing to
// queue canvas updates, which avoids a backlog building up while the display
// is not being drawn. Create one with newPauser.
type pauser struct {
	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
}

// newPauser returns a pauser in the running (not paused) state.
func newPauser() *pauser {
	p := &pauser{}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// wait blocks while the pauser is paused and returns immediately otherwise.
// Call it once at the top of each animation frame.
func (p *pauser) wait() {
	p.mu.Lock()
	for p.paused {
		p.cond.Wait()
	}
	p.mu.Unlock()
}

// setPaused pauses or resumes goroutines waiting on this pauser. Resuming wakes
// every waiter.
func (p *pauser) setPaused(paused bool) {
	p.mu.Lock()
	p.paused = paused
	p.mu.Unlock()
	p.cond.Broadcast()
}
