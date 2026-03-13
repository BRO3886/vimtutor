package metrics

import "time"

// Session represents one practice session.
type Session struct {
	ID        int64
	LessonID  string
	Mode      string // "lesson", "drill", "freeplay"
	StartedAt time.Time
	EndedAt   *time.Time
}

// KeystrokeEvent records a single keypress during a session.
type KeystrokeEvent struct {
	SessionID   int64
	Timestamp   time.Time
	Key         string
	WasCorrect  bool
	ChallengeID string
}

// LessonResult records the result of one step/challenge attempt.
type LessonResult struct {
	SessionID      int64
	LessonID       string
	StepID         string
	Attempts       int
	KeystrokesUsed int
	TimeSpentSecs  float64
	Passed         bool
	MistakeKeys    []string
}

// DailyStat aggregates per-day metrics.
type DailyStat struct {
	Date               string // YYYY-MM-DD
	TotalKeystrokes    int
	CorrectKeystrokes  int
	TimeSpentSecs      float64
	LessonsCompleted   int
	XPEarned           int
}

// UserProgress tracks per-lesson progress.
type UserProgress struct {
	LessonID       string
	BestKeystrokes int
	BestTimeSecs   float64
	CompletedCount int
	LastAttempted  *time.Time
	Unlocked       bool
}

// SessionSummary is computed, not stored.
type SessionSummary struct {
	Duration        time.Duration
	TotalKeystrokes int
	Accuracy        float64 // 0.0 - 1.0
	StepsCompleted  int
	XPEarned        int
	MistakeMap      map[string]int // key -> count
}

// OverallStats aggregates all-time metrics.
type OverallStats struct {
	TotalXP          int
	Level            int
	LevelName        string
	StreakDays       int
	TotalSessions    int
	TotalTimeSecs    float64
	LessonsCompleted int
	AvgAccuracy      float64
	TopMistakes      []MistakeEntry
	DailyXP          []DailyStat // last 14 days
}

// MistakeEntry is a key with its mistake count.
type MistakeEntry struct {
	Key   string
	Count int
}

// XP constants
const (
	XPStepFirstTime    = 10
	XPChallengeFirst   = 20
	XPChallengeHinted  = 5
	XPLessonComplete   = 50
	XPDailyStreak      = 25
	XPBeatPersonalBest = 15
	XPDrillComplete    = 30
)

var levelNames = []string{
	"Noob",
	"Normal Mode",
	"Visual Thinker",
	"Power User",
	"Operator",
	"Text Object Master",
	"Search Wizard",
	"Macro Recorder",
	"Register Keeper",
	"Vimscript Initiate",
	"Vim Grandmaster",
}

// LevelFromXP returns level (0-indexed) and name for a given XP total.
func LevelFromXP(xp int) (int, string) {
	level := xp / 100
	if level >= len(levelNames) {
		level = len(levelNames) - 1
	}
	return level, levelNames[level]
}

// XPToNextLevel returns XP needed to reach next level.
func XPToNextLevel(xp int) (current, needed int) {
	current = xp % 100
	needed = 100
	return
}
