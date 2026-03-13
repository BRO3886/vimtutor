package metrics

import (
	"sort"
	"time"
)

// Recorder tracks metrics during an active session.
type Recorder struct {
	session    Session
	events     []KeystrokeEvent
	results    []LessonResult
	currentXP  int
	stepStart  time.Time
	mistakeMap map[string]int
}

// NewRecorder creates a new recorder for the given session.
func NewRecorder(sessionID int64, lessonID, mode string) *Recorder {
	return &Recorder{
		session: Session{
			ID:        sessionID,
			LessonID:  lessonID,
			Mode:      mode,
			StartedAt: time.Now(),
		},
		mistakeMap: make(map[string]int),
		stepStart:  time.Now(),
	}
}

// RecordKey records a single keypress.
func (r *Recorder) RecordKey(key, challengeID string, correct bool) {
	r.events = append(r.events, KeystrokeEvent{
		SessionID:   r.session.ID,
		Timestamp:   time.Now(),
		Key:         key,
		WasCorrect:  correct,
		ChallengeID: challengeID,
	})
	if !correct && key != "esc" {
		r.mistakeMap[key]++
	}
}

// StartStep marks the beginning of a new step/challenge.
func (r *Recorder) StartStep() {
	r.stepStart = time.Now()
}

// RecordResult records the result of a step/challenge.
func (r *Recorder) RecordResult(lessonID, stepID string, attempts, keystrokes int, passed bool, mistakeKeys []string) {
	elapsed := time.Since(r.stepStart).Seconds()
	r.results = append(r.results, LessonResult{
		SessionID:      r.session.ID,
		LessonID:       lessonID,
		StepID:         stepID,
		Attempts:       attempts,
		KeystrokesUsed: keystrokes,
		TimeSpentSecs:  elapsed,
		Passed:         passed,
		MistakeKeys:    mistakeKeys,
	})
}

// AddXP adds XP to the current session.
func (r *Recorder) AddXP(amount int) {
	r.currentXP += amount
}

// Finalize ends the session and returns a summary.
func (r *Recorder) Finalize() (Session, []KeystrokeEvent, []LessonResult, SessionSummary) {
	now := time.Now()
	r.session.EndedAt = &now

	total := len(r.events)
	correct := 0
	for _, e := range r.events {
		if e.WasCorrect {
			correct++
		}
	}

	acc := 0.0
	if total > 0 {
		acc = float64(correct) / float64(total)
	}

	stepsCompleted := 0
	for _, res := range r.results {
		if res.Passed {
			stepsCompleted++
		}
	}

	summary := SessionSummary{
		Duration:        now.Sub(r.session.StartedAt),
		TotalKeystrokes: total,
		Accuracy:        acc,
		StepsCompleted:  stepsCompleted,
		XPEarned:        r.currentXP,
		MistakeMap:      r.mistakeMap,
	}

	return r.session, r.events, r.results, summary
}

// TopMistakes returns the N most common mistake keys.
func (r *Recorder) TopMistakes(n int) []MistakeEntry {
	entries := make([]MistakeEntry, 0, len(r.mistakeMap))
	for k, v := range r.mistakeMap {
		entries = append(entries, MistakeEntry{Key: k, Count: v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// KeystrokeCount returns total keystrokes recorded.
func (r *Recorder) KeystrokeCount() int {
	return len(r.events)
}

// SessionID returns the session ID.
func (r *Recorder) SessionID() int64 {
	return r.session.ID
}
