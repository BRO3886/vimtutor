# Architecture

vimtutor is a terminal UI application built in Go. It teaches vim interactively by running a real vim engine against a text buffer and checking whether the user achieved the expected result. Metrics are persisted to a local SQLite database and displayed in a progress dashboard.

---

## Package structure

```
vimtutor/
├── main.go                          entry point, delegates to cmd/
├── cmd/vimtutor/                    cobra CLI commands
│   ├── root.go                      root command, opens lesson menu
│   └── learn.go                     `learn`, `stats`, `reset` subcommands
└── internal/
    ├── engine/                      vim emulation (pure Go, no TUI deps)
    │   ├── buffer.go                text buffer + cursor + mode state
    │   ├── motion.go                all cursor movements + text objects
    │   ├── keyparser.go             key sequence state machine
    │   └── executor.go              routes parsed commands to buffer ops
    ├── curriculum/                  lesson content
    │   ├── lesson.go                Lesson / Step / Challenge data models
    │   ├── registry.go              global lesson registry (init()-based)
    │   └── lessons/                 one file per lesson, registered via init()
    ├── metrics/                     session tracking
    │   ├── models.go                data types + XP constants + level names
    │   └── recorder.go              in-memory accumulator for one session
    ├── storage/                     SQLite persistence
    │   ├── db.go                    open, schema migration, close
    │   └── queries.go               all read/write operations
    └── ui/                          Bubble Tea TUI
        ├── app.go                   root model, screen routing
        ├── theme/styles.go          lipgloss color palette + helpers
        ├── components/editor.go     buffer renderer + status bar
        └── screens/
            ├── menu.go              lesson selection screen
            ├── challenge.go         interactive lesson/challenge screen
            └── dashboard.go         progress stats screen
```

---

## The vim engine

The engine is the core of the project and has no dependency on the TUI. It can be tested in isolation.

### Buffer (`engine/buffer.go`)

The buffer holds the editor state:

```
lines  [][]rune   — text content, one slice per line
cursor Cursor      — {Line int, Col int}
mode   Mode        — Normal | Insert | Visual
registers map[rune]string  — named clipboard slots
```

All text mutations (insert rune, delete range, join lines, paste) are methods on `Buffer`. The buffer enforces cursor clamping so the cursor can never go out of bounds for the current mode — in Normal mode, the cursor can't sit past the last character; in Insert mode it can sit one past it.

### Key parser (`engine/keyparser.go`)

Vim key sequences are multi-keystroke. The parser is a state machine that buffers incoming keys and emits a `ParsedCmd` only when a complete command is assembled.

```
stateStart      → waiting for first key
stateCount      → consumed one or more digits (e.g. "3" in "3w")
stateRegister   → consumed " (register prefix)
stateOperator   → consumed d/c/y/>/< (waiting for motion or text object)
stateMotion     → consumed i/a (waiting for text object char)
stateAwaitChar  → consumed f/t/F/T/r (waiting for target character)
stateGPrefix    → consumed g (could be gg, g_, gj, gk)
stateZPrefix    → consumed z (could be zz, zt, zb)
```

`Feed(key string) (*ParsedCmd, needMore bool)` is the only public method. The caller feeds one key at a time. When `needMore=true`, the caller should show the pending keys in the status bar but not execute anything yet. When `needMore=false`, a `ParsedCmd` (or nil for unrecognized input) is returned and the parser resets.

A `ParsedCmd` carries:

```go
Count    int     // e.g. 3 in "3w"
Operator string  // "d", "c", "y", ">", "<"
Motion   Motion  // {Key: "w", Count: 3, Char: 'f'} etc.
Register rune    // named register, 0 = default
Linewise bool    // true for dd, cc, yy
Action   string  // "i", "a", "o", "p", "x", "esc", etc.
```

### Motions (`engine/motion.go`)

`ApplyMotion(m Motion) Range` moves the cursor and returns the swept range for use by operators. The range is `[L1,C1) → [L2,C2)` (exclusive end).

Word motions (`w`, `b`, `e`, `W`, `B`, `E`) are implemented as forward/backward scans that classify each rune as `charSpace | charWord | charPunct`. WORD variants treat all non-whitespace as one class.

Find motions (`f`, `t`, `F`, `T`) scan the current line. `;` and `,` repeat the last find in the same or opposite direction.

Text objects (`iw`, `aw`, `i"`, `a(`, `i{`, etc.) are computed separately via `TextObjectRange(obj string) Range` and do not move the cursor — they just return a range for the operator to act on.

### Executor (`engine/executor.go`)

`Execute(buf, cmd) bool` routes a `ParsedCmd` to the appropriate buffer operation and returns `true` if the buffer entered Insert mode (so the caller can switch its key-handling path).

One vim quirk is explicitly modeled here: `cw` is equivalent to `ce` — it deletes to end of word rather than to start of next word, so it doesn't eat the trailing space. The executor converts `c`+`w` → `c`+`e` before computing the range.

`InsertKey(buf, key) bool` handles keys while in Insert mode (printable characters, backspace, ctrl+w, ctrl+u, enter, esc).

---

## Curriculum

### Lesson model (`curriculum/lesson.go`)

```
Lesson
  ID            LessonID (string alias)
  Title
  Difficulty    Beginner | Intermediate | Advanced
  Prerequisites []LessonID
  Steps         []Step

Step
  ID, Title, Explanation   — teaching text shown above the editor
  InitText                 — buffer content when this step begins
  CursorLine, CursorCol    — where the cursor starts
  Challenge                — nil for explanatory steps, non-nil for interactive ones

Challenge
  Goal          — human-readable task description
  ExpectedText  — what buf.Text() must equal to pass (empty = don't check)
  ExpectedLine  — required cursor line after (-1 = don't check)
  ExpectedCol   — required cursor col after (-1 = don't check)
  Hint          — shown after 2 failed attempts
```

`Challenge.Check(text, line, col)` returns `(passed bool, reason string)`.

### Registry (`curriculum/registry.go`)

Each `lessons/NN_*.go` file calls `curriculum.Register(lesson)` from its `init()` function. The `lessons` package is imported for side effects only (`_ "…/curriculum/lessons"`) in both `ui/app.go` and `ui/screens/menu.go`. This keeps lessons decoupled from the registry — adding a new lesson is just adding a new file.

### The 10 lessons

| # | Title | Tags |
|---|-------|------|
| 01 | Basic Motions | motion, basics |
| 02 | Word Motions | motion, word |
| 03 | Operators: d, c, y | operator, edit |
| 04 | Text Objects | text-object, operator |
| 05 | Insert Mode Mastery | insert, basics |
| 06 | Search & Find | search, motion |
| 07 | Marks & Jumps | motion, marks |
| 08 | Registers | registers, yank |
| 09 | Macros | macro, advanced |
| 10 | Advanced Combinations | advanced, combo |

Each lesson unlocks the next when completed.

---

## Metrics

### Recorder (`metrics/recorder.go`)

`Recorder` accumulates data in memory during an active session. It is created at lesson start and holds:

- `[]KeystrokeEvent` — every key pressed, with timestamp and correctness flag
- `[]LessonResult` — one record per step: attempts, keystrokes used, time, mistakes
- `currentXP` — XP earned during this session
- `mistakeMap map[string]int` — key → mistake count

`Finalize()` closes the session timer and returns `(Session, []KeystrokeEvent, []LessonResult, SessionSummary)` for the caller to persist.

### XP system

| Action | XP |
|--------|----|
| Pass a challenge (first attempt) | 20 |
| Pass a challenge (after hint) | 5 |
| Complete a step (first time) | 10 |
| Complete a full lesson | 50 |
| Daily streak bonus | 25/day |
| Beat personal best keystrokes | 15 |

Levels advance every 100 XP. Level names: Noob → Normal Mode → Visual Thinker → Power User → Operator → Text Object Master → Search Wizard → Macro Recorder → Register Keeper → Vimscript Initiate → Vim Grandmaster.

---

## Storage

SQLite via `modernc.org/sqlite` (pure Go, no CGo). The database lives at `~/.vimtutor/vimtutor.db` and is created on first run.

### Schema

```sql
sessions          — one row per lesson attempt (lesson_id, mode, started_at, ended_at)
keystroke_events  — one row per keypress (session_id, ts, key, was_correct, challenge_id)
lesson_results    — one row per step completion (attempts, keystrokes_used, time_spent_secs, mistake_keys JSON)
daily_stats       — one row per calendar day (total/correct keystrokes, time, xp_earned)
user_progress     — one row per lesson (best_keystrokes, best_time_secs, completed_count, unlocked)
user_xp           — single-row table holding total XP
```

`db.SaveSession()` is called by `ui/app.go` after the lesson ends. It runs the full write sequence: close the session, batch-insert keystrokes and results, update daily stats, add XP. If `session_id` is 0 (CreateSession failed), the per-keystroke writes are skipped but XP/stats still accumulate.

`db.UnlockLesson(nextID)` is called immediately after saving, before returning to the menu.

---

## TUI

Built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) with [Lip Gloss](https://github.com/charmbracelet/lipgloss) for styling. Color palette is Tokyo Night dark.

### Screen model (`ui/app.go`)

`AppModel` is the root `tea.Model`. It owns a `screenKind` enum and one sub-model per screen. Only the active screen's `Update` and `View` are called each frame.

```
screenMenu       → MenuScreen
screenChallenge  → ChallengeScreen
screenDashboard  → DashboardScreen
```

Screen transitions happen via typed message structs (`StartLessonMsg`, `ShowDashboardMsg`, `BackToMenuMsg`). The root model inspects the incoming `tea.Msg` and switches screens accordingly. This avoids shared state between screens — each screen is self-contained.

After lesson completion, `app.go` calls `challenge.FinalizedData()`, persists everything to the DB, unlocks the next lesson, then sets `a.menu = nil` so the menu is rebuilt fresh from DB on the next render.

### ChallengeScreen (`ui/screens/challenge.go`)

The most complex component. It owns:

- A `*engine.Buffer` — the live editor state
- A `*engine.Parser` — accumulates keystrokes into commands
- A `*metrics.Recorder` — tracks everything

On each `tea.KeyMsg`:
1. If in Insert mode → `engine.InsertKey(buf, key)`
2. Otherwise → `parser.Feed(key)`
3. If `needMore` → show pending keys in status bar, no action yet
4. If complete command → `engine.Execute(buf, cmd)`
5. Call `checkChallenge()` — runs `challenge.Check(buf.Text(), cursor)` to test for pass

When all steps are done, `advance()` calls `recorder.Finalize()` and stores the result on the screen for `app.go` to retrieve.

### Editor renderer (`ui/components/editor.go`)

`RenderEditor(buf, opts)` produces a styled string of the buffer. It handles:

- A scroll window centered on the cursor when the buffer is taller than the editor height
- Line numbers (current line in yellow, others in dim)
- Current line background highlight
- Cursor rendered as a reverse-video block in Normal mode, underlined green in Insert mode
- Tab expansion (4 spaces)

`RenderStatusBar(buf, pendingKeys, width, extra)` renders the bottom bar with a mode pill (colored background), cursor position, pending key sequence, and right-side step counter.

### Package dependency graph

```
cmd/vimtutor
    └── internal/ui          (Run, RunLesson)
        ├── internal/ui/theme        (styles, no deps)
        ├── internal/ui/screens
        │   ├── internal/ui/components
        │   │   ├── internal/engine
        │   │   └── internal/ui/theme
        │   ├── internal/curriculum
        │   ├── internal/metrics
        │   └── internal/storage
        ├── internal/curriculum
        │   └── internal/curriculum/lessons  (side-effect import)
        ├── internal/metrics
        └── internal/storage
                └── internal/metrics
```

`internal/engine` and `internal/ui/theme` are leaf packages with no internal dependencies. This prevents import cycles — the `ui` package can import `ui/screens`, and `ui/screens` can import `ui/theme`, but neither reaches back up to `ui`.
