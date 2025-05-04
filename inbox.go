package mailtracker

import (
	"io"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
)

func (w *Tracker) checkInbox(handler func(Email)) {
	c, err := client.DialTLS(w.cfg.IMAPServer, nil)
	if err != nil {
		return
	}
	defer c.Logout()

	if err := c.Login(w.cfg.EmailAddress, w.cfg.EmailPassword); err != nil {
		return
	}

	// Open INBOX in read-write mode so we can delete emails
	mbox, err := c.Select("INBOX", false)
	if err != nil || mbox.Messages == 0 {
		return
	}

	from := uint32(1)
	if mbox.Messages > 5 {
		from = mbox.Messages - 4
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	section := &imap.BodySectionName{}

	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{
			imap.FetchEnvelope, imap.FetchUid, imap.FetchFlags, section.FetchItem(),
		}, messages)
	}()

	var toDelete imap.SeqSet

	for msg := range messages {
		if msg == nil || msg.Envelope == nil {
			continue
		}

		if w.cachedEmails[msg.Uid] {
			continue
		}

		msgTime := msg.Envelope.Date
		if time.Since(msgTime) > w.cfg.CacheInterval {
			continue
		}

		w.cachedEmails[msg.Uid] = true

		r := msg.GetBody(section)
		if r == nil {
			continue
		}

		m, err := message.Read(r)
		if err != nil {
			continue
		}

		email := Email{
			From:    msg.Envelope.From[0].Address(),
			Subject: msg.Envelope.Subject,
			Body:    extractBody(m),
			UID:     msg.Uid,
			Time:    msgTime,
		}

		handler(email)

		if w.cfg.DeleteCached {
			toDelete.AddNum(msg.SeqNum)
		}
	}

	<-done

	// Mark all to-delete messages as \Deleted and expunge
	if w.cfg.DeleteCached && !toDelete.Empty() {
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.DeletedFlag}
		if err := c.Store(&toDelete, item, flags, nil); err == nil {
			_ = c.Expunge(nil)
		}
	}
}

func extractBody(m *message.Entity) string {
	if mr := m.MultipartReader(); mr != nil {
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			d, _ := io.ReadAll(p.Body)
			return string(d)
		}
	}
	d, _ := io.ReadAll(m.Body)
	return string(d)
}
