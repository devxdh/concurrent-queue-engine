package worker

type Task struct {
	ID          int64
	URL         string
	Status      string
	Attempts    int
	MaxAttempts int
	LastError   *string
}
