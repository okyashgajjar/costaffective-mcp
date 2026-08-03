package ledger

type TypedEvent interface {
	ToEvent() Event
}

type StashCreateEvent struct {
	Handle  string
	Tokens  int
	Summary string
}

func (e StashCreateEvent) ToEvent() Event {
	return Event{
		Kind:    "stash",
		Action:  "create",
		Handle:  e.Handle,
		Tokens:  e.Tokens,
		Summary: e.Summary,
	}
}

type IndexEvent struct {
	Files   int
	Trigger string
}

func (e IndexEvent) ToEvent() Event {
	return Event{
		Kind:    "index",
		Action:  "reindex",
		Files:   e.Files,
		Trigger: e.Trigger,
	}
}

type FactAddEvent struct {
	Summary string
}

func (e FactAddEvent) ToEvent() Event {
	return Event{
		Kind:    "fact",
		Action:  "add",
		Summary: e.Summary,
	}
}

type RecallReadEvent struct {
	Query  string
	Source string
}

func (e RecallReadEvent) ToEvent() Event {
	return Event{
		Kind:   "recall",
		Action: "read",
		Query:  e.Query,
		Source: e.Source,
	}
}

type WatchAutoReindexEvent struct {
	ChangedFiles []string
}

func (e WatchAutoReindexEvent) ToEvent() Event {
	return Event{
		Kind:         "watch",
		Action:       "auto_reindex",
		ChangedFiles: e.ChangedFiles,
	}
}

func AppendTyped(repoPath string, evt TypedEvent) error {
	return Append(repoPath, evt.ToEvent())
}
