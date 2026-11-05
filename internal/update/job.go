package update

import (
	"errors"
	"sync"
	"time"
)

// Update phases as the admin page sees them.
const (
	PhaseIdle        = "idle"
	PhaseDownloading = "downloading"
	PhaseVerifying   = "verifying"
	PhaseRestarting  = "restarting"
	PhaseFailed      = "failed"
)

// Failure codes; the admin page turns them into plain words.
const (
	CodeDownload = "download" // could not fetch the release
	CodeStalled  = "stalled"  // the download stopped sending data
	CodeWrite    = "write"    // no way to write into the exe folder
	CodeStart    = "start"    // the new version would not start
	CodeLocked   = "locked"   // the exe stayed locked, the old version came back
	CodeUnknown  = "unknown"
)

// Failure is an update error with a code the admin page understands.
type Failure struct {
	Code string
	Err  error
}

func (f *Failure) Error() string { return f.Err.Error() }
func (f *Failure) Unwrap() error { return f.Err }

func failureCode(err error) string {
	var f *Failure
	if errors.As(err, &f) {
		return f.Code
	}
	return CodeUnknown
}

// Progress is the current state of an update.
type Progress struct {
	Phase  string `json:"phase"`
	Latest string `json:"latest,omitempty"`
	Done   int64  `json:"done,omitempty"`
	Total  int64  `json:"total,omitempty"`
	Code   string `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// Job runs one update at a time and remembers how it went. A failed install from
// the previous run shows up as its starting state.
type Job struct {
	mu sync.Mutex
	p  Progress
}

func NewJob(dataDir string) *Job {
	j := &Job{p: Progress{Phase: PhaseIdle}}
	if code := takeSwapResult(dataDir); code != "" {
		j.p = Progress{Phase: PhaseFailed, Code: code}
	}
	return j
}

func (j *Job) Status() Progress {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.p
}

// Dismiss forgets a finished failure.
func (j *Job) Dismiss() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.p.Phase == PhaseFailed {
		j.p = Progress{Phase: PhaseIdle}
	}
}

// Start begins an update in the background unless one is already running. quit is
// called once the new version has taken over.
func (j *Job) Start(check func() Info, dataDir string, quit func()) {
	j.mu.Lock()
	if j.p.Phase != PhaseIdle && j.p.Phase != PhaseFailed {
		j.mu.Unlock()
		return
	}
	j.p = Progress{Phase: PhaseDownloading}
	j.mu.Unlock()

	go func() {
		info := check()
		if !info.Available {
			code, detail := CodeUnknown, "no update available"
			if info.Error != "" {
				code, detail = CodeDownload, info.Error
			}
			j.fail(code, detail)
			return
		}
		j.report(PhaseDownloading, info.Latest, 0, 0)
		err := Apply(info, dataDir, func(phase string, done, total int64) {
			j.report(phase, info.Latest, done, total)
		})
		if err != nil {
			j.fail(failureCode(err), err.Error())
			return
		}
		j.report(PhaseRestarting, info.Latest, 0, 0)
		time.Sleep(time.Second) // let the admin page see the last step
		quit()
	}()
}

func (j *Job) report(phase, latest string, done, total int64) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.p = Progress{Phase: phase, Latest: latest, Done: done, Total: total}
}

func (j *Job) fail(code, detail string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.p = Progress{Phase: PhaseFailed, Latest: j.p.Latest, Code: code, Detail: detail}
}
