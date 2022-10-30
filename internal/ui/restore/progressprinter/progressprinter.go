package progressprinter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var progresswriter io.Writer = os.Stdout

type progressInfoEntry struct {
	name         string
	bytesWritten int64
	bytesTotal   int64
}

type RestoreProgressPrinter struct {
	sync.Mutex
	progressInfoMap map[string]progressInfoEntry
	filesFinished   int64
	filesTotal      int64
	allBytesWritten int64
	allBytesTotal   int64
	lastPrinted     string
	started         time.Time
}

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

func formatLeftTime(duration int64) string {
	durationSeconds := (duration / int64(time.Second))
	durationMinutes := (durationSeconds / 60)
	durationHours := (durationMinutes / 60)
	if durationMinutes >= 60 {
		return fmt.Sprintf("%d:%02d:%02d", durationHours, durationMinutes%60, durationSeconds%60)
	}
	return fmt.Sprintf("%d:%02d", durationMinutes%60, durationSeconds%60)
}

func calculatePercent(done, from int64) float64 {
	if from == 0 {
		return 0.0
	}
	if from == done {
		return 100.0
	}
	return float64(100) / float64(from) * float64(done)
}

func print(p *RestoreProgressPrinter) {
	timeLeft := formatLeftTime(int64(time.Since(p.started)))
	formattedAllBytesWritten := formatBytesInBestUnit(p.allBytesWritten)
	formattedAllBytesTotal := formatBytesInBestUnit(p.allBytesTotal)
	allPercent := calculatePercent(p.allBytesWritten, p.allBytesTotal)
	text := fmt.Sprintf("[%s]  %d / %d Files,  %s / %s,  %.2f %%   \r",
		timeLeft, p.filesFinished, p.filesTotal, formattedAllBytesWritten, formattedAllBytesTotal, allPercent)
	fmt.Fprint(progresswriter, text)
	p.lastPrinted = text
}

func New() *RestoreProgressPrinter {
	return &RestoreProgressPrinter{
		progressInfoMap: make(map[string]progressInfoEntry),
		started:         time.Now(),
	}
}

func (p *RestoreProgressPrinter) AddFile(size int64) {
	p.Lock()
	defer p.Unlock()
	p.filesTotal++
	p.allBytesTotal += size
}

func (p *RestoreProgressPrinter) LogProgress(name string, bytesWrittenPortion int64, bytesTotal int64) {
	p.Lock()
	defer p.Unlock()
	entry, exists := p.progressInfoMap[name]
	if !exists {
		entry.name = name
		entry.bytesTotal = bytesTotal
	}
	entry.bytesWritten = entry.bytesWritten + bytesWrittenPortion
	p.progressInfoMap[name] = entry
	p.allBytesWritten += bytesWrittenPortion
	if entry.bytesWritten == entry.bytesTotal {
		delete(p.progressInfoMap, name)
		p.filesFinished++
	}
	print(p)
}

func (p *RestoreProgressPrinter) PrintSummary() {
	time.Sleep(time.Second)
	p.Lock()
	defer p.Unlock()
	timeLeft := formatLeftTime(int64(time.Since(p.started)))
	formattedAllBytesTotal := formatBytesInBestUnit(p.allBytesTotal)
	if p.filesFinished == p.filesTotal && p.allBytesWritten == p.allBytesTotal {
		lineRemovalSpaces := strings.Repeat(" ", len(p.lastPrinted))
		fmt.Fprintf(progresswriter, "%s\rSummary: Restored %d Files (%s) in %s\n",
			lineRemovalSpaces, p.filesTotal, formattedAllBytesTotal, timeLeft)
	} else {
		formattedAllBytesWritten := formatBytesInBestUnit(p.allBytesWritten)
		fmt.Fprintf(progresswriter, "Summary: Restored %d / %d Files (%s / %s) in %s\n",
			p.filesFinished, p.filesTotal, formattedAllBytesWritten, formattedAllBytesTotal, timeLeft)
	}
}
