package progressformatter

import (
	"fmt"
	"sync"
	"time"
)

type progressInfoEntry struct {
	bytesWritten int64
	bytesTotal   int64
}

type RestoreProgressFormatter struct {
	sync.Mutex
	progressInfoMap map[string]progressInfoEntry
	filesFinished   int64
	filesTotal      int64
	allBytesWritten int64
	allBytesTotal   int64
	started         time.Time
}

// TODO: replace with internal/ui/format -> FormatBytes when pull-request 3983 is merged
func formatBytesInBestUnit(inBytes int64) string {
	value := float64(inBytes)
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	unitIndex := 0
	for value >= 1024 && unitIndex < 4 {
		value = value / float64(1024)
		unitIndex++
	}
	if unitIndex == 0 {
		return fmt.Sprintf("%d %v", uint64(value), units[unitIndex])
	}
	return fmt.Sprintf("%.2f %v", value, units[unitIndex])
}

// TODO: replace with internal/ui/format -> FormatDuration when pull-request 3983 is merged
func formatLeftTime(duration int64) string {
	durationSeconds := (duration / int64(time.Second))
	durationMinutes := (durationSeconds / 60)
	durationHours := (durationMinutes / 60)
	if durationMinutes >= 60 {
		return fmt.Sprintf("%d:%02d:%02d", durationHours, durationMinutes%60, durationSeconds%60)
	}
	return fmt.Sprintf("%d:%02d", durationMinutes%60, durationSeconds%60)
}

// TODO: replace with internal/ui/format -> FormatPercent when pull-request 3983 is merged
func formatPercent(done, from int64) string {
	if from == 0 {
		return "0.00 %"
	}
	if from == done {
		return "100.00 %"
	}
	percent :=  float64(100) / float64(from) * float64(done)
	return fmt.Sprintf("%.2f %%", percent)
}

func format(p *RestoreProgressFormatter) string {
	timeLeft := formatLeftTime(int64(time.Since(p.started)))
	formattedAllBytesWritten := formatBytesInBestUnit(p.allBytesWritten)
	formattedAllBytesTotal := formatBytesInBestUnit(p.allBytesTotal)
	allPercent := formatPercent(p.allBytesWritten, p.allBytesTotal)
	return fmt.Sprintf("  [%s]  %d / %d Files,  %s / %s,  %s  ",
		timeLeft, p.filesFinished, p.filesTotal, formattedAllBytesWritten, formattedAllBytesTotal, allPercent)
}

func NewFormatter() *RestoreProgressFormatter {
	return &RestoreProgressFormatter{
		progressInfoMap: make(map[string]progressInfoEntry),
		started:         time.Now(),
	}
}

func (p *RestoreProgressFormatter) AddFile(size int64) {
	p.Lock()
	defer p.Unlock()
	
	p.filesTotal++
	p.allBytesTotal += size
}

func (p *RestoreProgressFormatter) FormatProgress(name string, bytesWrittenPortion int64, bytesTotal int64) string {
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
	
	return format(p)
}

func (p *RestoreProgressFormatter) FormatSummary() string {
	p.Lock()
	defer p.Unlock()
	
	timeLeft := formatLeftTime(int64(time.Since(p.started)))
	formattedAllBytesTotal := formatBytesInBestUnit(p.allBytesTotal)
	if p.filesFinished == p.filesTotal && p.allBytesWritten == p.allBytesTotal {
		return fmt.Sprintf("Summary: Restored %d Files (%s) in %s", p.filesTotal, formattedAllBytesTotal, timeLeft)
	} else {
		formattedAllBytesWritten := formatBytesInBestUnit(p.allBytesWritten)
		return fmt.Sprintf("Summary: Restored %d / %d Files (%s / %s) in %s",
			p.filesFinished, p.filesTotal, formattedAllBytesWritten, formattedAllBytesTotal, timeLeft)
	}
}
