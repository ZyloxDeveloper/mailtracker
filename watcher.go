package mailtracker

import (
	"context"
	"time"
)

type Tracker struct {
	cfg          TrackerConfig
	cachedEmails map[uint32]bool
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewTracker(cfg TrackerConfig) *Tracker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Tracker{
		cfg:          cfg,
		cachedEmails: make(map[uint32]bool),
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (w *Tracker) Start(handler func(Email)) {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.checkInbox(handler)
			time.Sleep(w.cfg.CheckInterval)
		}
	}
}

func (w *Tracker) Stop() {
	w.cancel()
}
