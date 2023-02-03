package restore

import (
	"fmt"
	"sync"
	"time"

	"github.com/restic/restic/internal/ui"
	"github.com/restic/restic/internal/ui/termstatus"
)

type Progress struct {
	sync.Mutex
	progressInfoMap map[string]progressInfoEntry
	filesFinished   uint64
	filesTotal      uint64
	allBytesWritten uint64
	allBytesTotal   uint64
	started         time.Time

	printer ProgressPrinter
}

type progressInfoEntry struct {
	bytesWritten uint64
	bytesTotal   uint64
}

type ProgressPrinter interface {
	Update(filesFinished, filesTotal, allBytesWritten, allBytesTotal uint64, started time.Time)
	Finish(filesFinished, filesTotal, allBytesWritten, allBytesTotal uint64, started time.Time)
}

func NewProgress(printer ProgressPrinter) *Progress {
	return &Progress{
		progressInfoMap: make(map[string]progressInfoEntry),
		started:         time.Now(),
		printer:         printer,
	}
}

func (p *Progress) AddFile(size uint64) {
	p.Lock()
	defer p.Unlock()

	p.filesTotal++
	p.allBytesTotal += size
}

func (p *Progress) AddProgress(name string, bytesWrittenPortion uint64, bytesTotal uint64) {
	p.Lock()
	defer p.Unlock()

	entry, exists := p.progressInfoMap[name]
	if !exists {
		entry.bytesTotal = bytesTotal
	}
	entry.bytesWritten += bytesWrittenPortion
	p.progressInfoMap[name] = entry

	p.allBytesWritten += bytesWrittenPortion
	if entry.bytesWritten == entry.bytesTotal {
		delete(p.progressInfoMap, name)
		p.filesFinished++
	}

	p.printer.Update(p.filesFinished, p.filesTotal, p.allBytesWritten, p.allBytesTotal, p.started)
}

func (p *Progress) Finish() {
	p.Lock()
	defer p.Unlock()

	p.printer.Finish(p.filesFinished, p.filesTotal, p.allBytesWritten, p.allBytesTotal, p.started)
}

type textPrinter struct {
	terminal *termstatus.Terminal
}

func NewProgressPrinter(terminal *termstatus.Terminal) ProgressPrinter {
	return &textPrinter{
		terminal: terminal,
	}
}

func (t *textPrinter) Update(filesFinished, filesTotal, allBytesWritten, allBytesTotal uint64, started time.Time) {
	timeLeft := ui.FormatDuration(time.Since(started))
	formattedAllBytesWritten := ui.FormatBytes(allBytesWritten)
	formattedAllBytesTotal := ui.FormatBytes(allBytesTotal)
	allPercent := ui.FormatPercent(allBytesWritten, allBytesTotal)
	progress := fmt.Sprintf("[%s]  %d / %d Files,  %s / %s,  %s",
		timeLeft, filesFinished, filesTotal, formattedAllBytesWritten, formattedAllBytesTotal, allPercent)

	t.terminal.SetStatus([]string{progress})
}

func (t *textPrinter) Finish(filesFinished, filesTotal, allBytesWritten, allBytesTotal uint64, started time.Time) {
	t.terminal.SetStatus([]string{})

	timeLeft := ui.FormatDuration(time.Since(started))
	formattedAllBytesTotal := ui.FormatBytes(allBytesTotal)

	var summary string
	if filesFinished == filesTotal && allBytesWritten == allBytesTotal {
		summary = fmt.Sprintf("Summary: Restored %d Files (%s) in %s", filesTotal, formattedAllBytesTotal, timeLeft)
	} else {
		formattedAllBytesWritten := ui.FormatBytes(allBytesWritten)
		summary = fmt.Sprintf("Summary: Restored %d / %d Files (%s / %s) in %s",
			filesFinished, filesTotal, formattedAllBytesWritten, formattedAllBytesTotal, timeLeft)
	}

	t.terminal.Print(summary)
}
