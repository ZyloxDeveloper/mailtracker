package mailtracker

import (
	"context"
	"sync"
	"time"
)

type Tracker struct {
	cfg          TrackerConfig
	cachedEmails map[uint32]bool

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	running bool
}

func NewTracker(cfg TrackerConfig) *Tracker {
	return &Tracker{
		cfg:          cfg,
		cachedEmails: make(map[uint32]bool),
	}
}

func (w *Tracker) Start(handler func(Email)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.running = true

	go func() {
		defer func() {
			w.mu.Lock()
			defer w.mu.Unlock()
			w.running = false
		}()

		for {
			select {
			case <-w.ctx.Done():
				return
			default:
				w.checkInbox(handler)
				time.Sleep(w.cfg.CheckInterval)
			}
		}
	}()
}

func (w *Tracker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
	}
}
