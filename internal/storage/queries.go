package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/vimtutor/internal/metrics"
)

// CreateSession inserts a new session and returns its ID.
func (d *DB) CreateSession(lessonID, mode string) (int64, error) {
	res, err := d.db.Exec(
		`INSERT INTO sessions (lesson_id, mode, started_at) VALUES (?, ?, ?)`,
		lessonID, mode, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// EndSession marks a session as ended.
func (d *DB) EndSession(id int64) error {
	_, err := d.db.Exec(
		`UPDATE sessions SET ended_at = ? WHERE id = ?`,
		time.Now().Format(time.RFC3339), id,
	)
	return err
}

// SaveKeystrokes batch-inserts keystroke events.
func (d *DB) SaveKeystrokes(events []metrics.KeystrokeEvent) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.Prepare(
		`INSERT INTO keystroke_events (session_id, ts, key, was_correct, challenge_id) VALUES (?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		correct := 0
		if e.WasCorrect {
			correct = 1
		}
		if _, err := stmt.Exec(e.SessionID, e.Timestamp.Format(time.RFC3339), e.Key, correct, e.ChallengeID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SaveLessonResults batch-inserts lesson results.
func (d *DB) SaveLessonResults(results []metrics.LessonResult) error {
	if len(results) == 0 {
		return nil
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.Prepare(`
		INSERT INTO lesson_results
		(session_id, lesson_id, step_id, attempts, keystrokes_used, time_spent_secs, passed, mistake_keys)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range results {
		passed := 0
		if r.Passed {
			passed = 1
		}
		mkJSON, _ := json.Marshal(r.MistakeKeys)
		if _, err := stmt.Exec(
			r.SessionID, r.LessonID, r.StepID,
			r.Attempts, r.KeystrokesUsed, r.TimeSpentSecs,
			passed, string(mkJSON),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// AddXP adds XP to the user's total and updates daily stats.
func (d *DB) AddXP(amount int, date string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(`UPDATE user_xp SET total_xp = total_xp + ? WHERE id = 1`, amount); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO daily_stats (date, xp_earned) VALUES (?, ?)
		ON CONFLICT(date) DO UPDATE SET xp_earned = xp_earned + excluded.xp_earned`,
		date, amount,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateDailyStats updates today's keystroke and time stats.
func (d *DB) UpdateDailyStats(date string, totalKeys, correctKeys int, timeSecs float64) error {
	_, err := d.db.Exec(`
		INSERT INTO daily_stats (date, total_keystrokes, correct_keystrokes, time_spent_secs)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			total_keystrokes   = total_keystrokes + excluded.total_keystrokes,
			correct_keystrokes = correct_keystrokes + excluded.correct_keystrokes,
			time_spent_secs    = time_spent_secs + excluded.time_spent_secs`,
		date, totalKeys, correctKeys, timeSecs,
	)
	return err
}

// UpdateProgress updates the lesson progress record.
func (d *DB) UpdateProgress(lessonID string, keystrokes int, timeSecs float64, passed bool) error {
	now := time.Now().Format(time.RFC3339)
	_, err := d.db.Exec(`
		INSERT INTO user_progress (lesson_id, best_keystrokes, best_time_secs, completed_count, last_attempted, unlocked)
		VALUES (?, ?, ?, ?, ?, 1)
		ON CONFLICT(lesson_id) DO UPDATE SET
			best_keystrokes = CASE WHEN excluded.best_keystrokes < best_keystrokes OR best_keystrokes IS NULL
				THEN excluded.best_keystrokes ELSE best_keystrokes END,
			best_time_secs  = CASE WHEN excluded.best_time_secs < best_time_secs OR best_time_secs IS NULL
				THEN excluded.best_time_secs ELSE best_time_secs END,
			completed_count = completed_count + CASE WHEN ? THEN 1 ELSE 0 END,
			last_attempted  = excluded.last_attempted,
			unlocked        = 1`,
		lessonID, keystrokes, timeSecs, now, passed,
	)
	return err
}

// UnlockLesson marks a lesson as unlocked.
func (d *DB) UnlockLesson(lessonID string) error {
	_, err := d.db.Exec(`
		INSERT INTO user_progress (lesson_id, unlocked) VALUES (?, 1)
		ON CONFLICT(lesson_id) DO UPDATE SET unlocked = 1`,
		lessonID,
	)
	return err
}

// GetTotalXP returns the user's total XP.
func (d *DB) GetTotalXP() (int, error) {
	var xp int
	err := d.db.QueryRow(`SELECT total_xp FROM user_xp WHERE id = 1`).Scan(&xp)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return xp, err
}

// GetAllProgress returns all lesson progress records.
func (d *DB) GetAllProgress() (map[string]metrics.UserProgress, error) {
	rows, err := d.db.Query(`
		SELECT lesson_id, best_keystrokes, best_time_secs, completed_count, last_attempted, unlocked
		FROM user_progress`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]metrics.UserProgress)
	for rows.Next() {
		var p metrics.UserProgress
		var bestKS sql.NullInt64
		var bestTime sql.NullFloat64
		var lastAttempted sql.NullString
		var unlocked int
		if err := rows.Scan(&p.LessonID, &bestKS, &bestTime, &p.CompletedCount, &lastAttempted, &unlocked); err != nil {
			return nil, err
		}
		if bestKS.Valid {
			p.BestKeystrokes = int(bestKS.Int64)
		}
		if bestTime.Valid {
			p.BestTimeSecs = bestTime.Float64
		}
		if lastAttempted.Valid {
			t, _ := time.Parse(time.RFC3339, lastAttempted.String)
			p.LastAttempted = &t
		}
		p.Unlocked = unlocked == 1
		result[p.LessonID] = p
	}
	return result, rows.Err()
}

// GetDailyStats returns the last N days of stats.
func (d *DB) GetDailyStats(days int) ([]metrics.DailyStat, error) {
	rows, err := d.db.Query(`
		SELECT date, total_keystrokes, correct_keystrokes, time_spent_secs, lessons_completed, xp_earned
		FROM daily_stats
		ORDER BY date DESC
		LIMIT ?`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []metrics.DailyStat
	for rows.Next() {
		var s metrics.DailyStat
		if err := rows.Scan(&s.Date, &s.TotalKeystrokes, &s.CorrectKeystrokes, &s.TimeSpentSecs, &s.LessonsCompleted, &s.XPEarned); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	// reverse to chronological order
	for i, j := 0, len(stats)-1; i < j; i, j = i+1, j-1 {
		stats[i], stats[j] = stats[j], stats[i]
	}
	return stats, rows.Err()
}

// GetStreakDays returns the current consecutive day streak.
func (d *DB) GetStreakDays() (int, error) {
	rows, err := d.db.Query(`
		SELECT date FROM daily_stats
		WHERE xp_earned > 0
		ORDER BY date DESC`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	streak := 0
	expected := time.Now().Format("2006-01-02")
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			return streak, err
		}
		if date == expected {
			streak++
			t, _ := time.Parse("2006-01-02", date)
			expected = t.AddDate(0, 0, -1).Format("2006-01-02")
		} else {
			break
		}
	}
	return streak, nil
}

// GetTopMistakes returns the most frequent mistake keys across all sessions.
func (d *DB) GetTopMistakes(n int) ([]metrics.MistakeEntry, error) {
	rows, err := d.db.Query(`
		SELECT mistake_keys FROM lesson_results WHERE mistake_keys IS NOT NULL AND mistake_keys != '[]'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var mkJSON string
		if err := rows.Scan(&mkJSON); err != nil {
			continue
		}
		var keys []string
		if err := json.Unmarshal([]byte(mkJSON), &keys); err != nil {
			continue
		}
		for _, k := range keys {
			counts[k]++
		}
	}

	entries := make([]metrics.MistakeEntry, 0, len(counts))
	for k, v := range counts {
		entries = append(entries, metrics.MistakeEntry{Key: k, Count: v})
	}

	// sort descending
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Count > entries[i].Count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n], rows.Err()
}

// GetOverallStats assembles all-time stats.
func (d *DB) GetOverallStats() (*metrics.OverallStats, error) {
	xp, err := d.GetTotalXP()
	if err != nil {
		return nil, err
	}

	level, levelName := metrics.LevelFromXP(xp)

	streak, err := d.GetStreakDays()
	if err != nil {
		return nil, err
	}

	var totalSessions int
	d.db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&totalSessions) //nolint:errcheck

	var totalTimeSecs float64
	d.db.QueryRow(`SELECT COALESCE(SUM(time_spent_secs), 0) FROM lesson_results`).Scan(&totalTimeSecs) //nolint:errcheck

	var lessonsCompleted int
	d.db.QueryRow(`SELECT COALESCE(SUM(lessons_completed), 0) FROM daily_stats`).Scan(&lessonsCompleted) //nolint:errcheck

	var totalKeys, correctKeys int
	d.db.QueryRow(`SELECT COALESCE(SUM(total_keystrokes),0), COALESCE(SUM(correct_keystrokes),0) FROM daily_stats`).Scan(&totalKeys, &correctKeys) //nolint:errcheck

	avgAcc := 0.0
	if totalKeys > 0 {
		avgAcc = float64(correctKeys) / float64(totalKeys)
	}

	topMistakes, err := d.GetTopMistakes(5)
	if err != nil {
		return nil, err
	}

	dailyXP, err := d.GetDailyStats(14)
	if err != nil {
		return nil, err
	}

	return &metrics.OverallStats{
		TotalXP:          xp,
		Level:            level,
		LevelName:        levelName,
		StreakDays:        streak,
		TotalSessions:    totalSessions,
		TotalTimeSecs:    totalTimeSecs,
		LessonsCompleted: lessonsCompleted,
		AvgAccuracy:      avgAcc,
		TopMistakes:      topMistakes,
		DailyXP:          dailyXP,
	}, nil
}

// IsLessonUnlocked checks if a lesson is unlocked.
func (d *DB) IsLessonUnlocked(lessonID string) bool {
	// Lesson "01" is always unlocked
	if strings.HasPrefix(lessonID, "01") {
		return true
	}
	var unlocked int
	d.db.QueryRow(`SELECT unlocked FROM user_progress WHERE lesson_id = ?`, lessonID).Scan(&unlocked) //nolint:errcheck
	return unlocked == 1
}

// GetSessionCount returns total session count.
func (d *DB) GetSessionCount() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&n) //nolint:errcheck
	return n
}

// RecordLessonCompletion marks a lesson day as completed.
func (d *DB) RecordLessonCompletion(date string) error {
	_, err := d.db.Exec(`
		INSERT INTO daily_stats (date, lessons_completed) VALUES (?, 1)
		ON CONFLICT(date) DO UPDATE SET lessons_completed = lessons_completed + 1`,
		date,
	)
	return err
}

// SaveSession persists a session and its associated data.
func (d *DB) SaveSession(sess metrics.Session, events []metrics.KeystrokeEvent, results []metrics.LessonResult, summary metrics.SessionSummary) error {
	if err := d.EndSession(sess.ID); err != nil {
		return fmt.Errorf("end session: %w", err)
	}

	if err := d.SaveKeystrokes(events); err != nil {
		return fmt.Errorf("save keystrokes: %w", err)
	}

	if err := d.SaveLessonResults(results); err != nil {
		return fmt.Errorf("save results: %w", err)
	}

	date := time.Now().Format("2006-01-02")

	correctKeys := int(float64(summary.TotalKeystrokes) * summary.Accuracy)
	if err := d.UpdateDailyStats(date, summary.TotalKeystrokes, correctKeys, summary.Duration.Seconds()); err != nil {
		return fmt.Errorf("update daily stats: %w", err)
	}

	if err := d.AddXP(summary.XPEarned, date); err != nil {
		return fmt.Errorf("add xp: %w", err)
	}

	return nil
}
