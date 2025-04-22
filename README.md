# EmailWatcher

Simple Go package for watching an IMAP inbox and handling new emails.

## Usage

```go
cfg := emailwatcher.WatcherConfig{
	IMAPServer:    "imap.gmail.com:993",
	EmailAddress:  "your.email@gmail.com",
	EmailPassword: "your-app-password",
	CheckInterval: 5 * time.Second,
}

watcher := emailwatcher.NewWatcher(cfg)

go watcher.Start(func(email emailwatcher.Email) {
	fmt.Println(email.From)
	fmt.Println(email.Subject)
	fmt.Println(email.Body)
})
