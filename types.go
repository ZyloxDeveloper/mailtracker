package mailtracker

import "time"

type Email struct {
	From    string
	Subject string
	Body    string
	UID     uint32
	Time    time.Time
}

type TrackerConfig struct {
	IMAPServer    string
	EmailAddress  string
	EmailPassword string
	CheckInterval time.Duration
	CacheInterval time.Duration
	DeleteCached  bool
}
