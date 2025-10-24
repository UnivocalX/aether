
package registry

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressTracker manages progress bar display
type ProgressTracker struct {
	bar   *progressbar.ProgressBar
	total int
}

// NewProgressTracker creates a new progress tracker for file uploads
func NewProgressTracker(totalFiles int) *ProgressTracker {
	bar := progressbar.NewOptions(totalFiles,
		progressbar.OptionSetDescription("Uploading files"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("files"),
	)

	return &ProgressTracker{
		bar:   bar,
		total: totalFiles,
	}
}

// Increment increases the progress bar by one
func (pt *ProgressTracker) Increment() {
	if pt.bar != nil {
		pt.bar.Add(1)
	}
}

// Finish completes the progress bar
func (pt *ProgressTracker) Finish() {
	if pt.bar != nil {
		pt.bar.Finish()
		fmt.Println() // Add newline after completion
	}
}

// Close closes the progress bar
func (pt *ProgressTracker) Close() {
	if pt.bar != nil {
		pt.bar.Close()
	}
}