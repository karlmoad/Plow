package objects

import "time"

type LogItemEntry struct {
	TrackingId string
	FileName   string
	Reference  string
	Hash       string
	Status     bool
	ApplyDate  time.Time
	Message    string
	Partial    bool
}

type LogEntry struct {
	TrackingId        string
	Message           string
	Start             time.Time
	End               time.Time
	AppliedBy         string
	TotalChanges      int
	SuccessfulChanges int
	FailedChanges     int
	Completed         bool
	FastForward       bool
}

type TrackingLog struct {
	Empty bool
	index map[string]int
	items []LogEntry
}

func NewTrackingLog() *TrackingLog {
	return &TrackingLog{index: make(map[string]int), items: make([]LogEntry, 0)}
}

func (t *TrackingLog) Add(entry LogEntry) {
	idx := len(t.items)
	t.items = append(t.items, entry)
	t.index[entry.TrackingId] = idx
}

func (t *TrackingLog) Find(id string) bool {
	if _, ok := t.index[id]; ok {
		return ok
	}
	return false
}

func (t *TrackingLog) FindAndGet(id string) (*LogEntry, bool) {
	if v, ok := t.index[id]; ok {
		return &t.items[v], ok
	}
	return nil, false
}

func (t *TrackingLog) GetLastProcessed() *LogEntry {
	if len(t.items) > 0 {
		return &t.items[0]
	}
	return nil
}
