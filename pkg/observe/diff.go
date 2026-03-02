package observe

// DiffResult describes differences between two runs.
type DiffResult struct {
	RunA     string
	RunB     string
	StepsA   int
	StepsB   int
	Diverged int    // step index where runs diverged (-1 if identical)
	Message  string // human-readable summary
}

// Diff compares two event sequences.
func Diff(eventsA, eventsB []Event) DiffResult {
	r := DiffResult{StepsA: len(eventsA), StepsB: len(eventsB)}
	if len(eventsA) > 0 {
		r.RunA = eventsA[0].RunID
	}
	if len(eventsB) > 0 {
		r.RunB = eventsB[0].RunID
	}

	n := len(eventsA)
	if len(eventsB) < n {
		n = len(eventsB)
	}
	for i := 0; i < n; i++ {
		if !eventsEqual(eventsA[i], eventsB[i]) {
			r.Diverged = i
			r.Message = "runs diverged at step " + string(rune('0'+i))
			return r
		}
	}
	if len(eventsA) != len(eventsB) {
		r.Diverged = n
		r.Message = "runs have different step counts"
		return r
	}
	r.Diverged = -1
	r.Message = "runs are identical"
	return r
}

func eventsEqual(a, b Event) bool {
	if a.Type != b.Type || a.StepID != b.StepID {
		return false
	}
	if !a.Timestamp.Equal(b.Timestamp) {
		return false
	}
	if a.Model.Provider != b.Model.Provider || a.Model.Name != b.Model.Name {
		return false
	}
	return true
}
