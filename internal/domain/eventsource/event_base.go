package eventsource

import "time"

type EventBase struct {
	ID        string
	Version   uint
	CreatedAt time.Time
}

func (a EventBase) GetID() string {
	return a.ID
}

func (a EventBase) GetVersion() uint {
	return a.Version
}

func (a EventBase) GetCreatedAt() time.Time {
	return a.CreatedAt
}
