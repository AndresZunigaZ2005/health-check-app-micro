package notifier

import "log"

type Notifier struct{}

func New() *Notifier { return &Notifier{} }

func (n *Notifier) Notify(recipients []string, subject, body string) {
	if len(recipients) == 0 {
		return
	}
	// por ahora solo log
	log.Printf("[NOTIFY] to=%v subject=%s body=%s", recipients, subject, body)
}
