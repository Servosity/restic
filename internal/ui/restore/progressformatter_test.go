package restore

import (
	"testing"
	"time"

	"github.com/restic/restic/internal/test"
)

type stubPrinter struct{}

func (p *stubPrinter) Update(filesFinished, filesTotal, allBytesWritten, allBytesTotal uint64, started time.Time) {
}
func (p *stubPrinter) Finish(filesFinished, filesTotal, allBytesWritten, allBytesTotal uint64, started time.Time) {
}

func TestNew(t *testing.T) {
	formatter := NewProgress(nil)
	test.Equals(t, uint64(0), formatter.filesFinished)
	test.Equals(t, uint64(0), formatter.filesTotal)
	test.Equals(t, uint64(0), formatter.allBytesWritten)
	test.Equals(t, uint64(0), formatter.allBytesTotal)
}

func TestAddFile(t *testing.T) {
	expectedFilesTotal := uint64(1)
	expectedAllBytesTotal := uint64(100)

	formatter := NewProgress(nil)
	formatter.AddFile(expectedAllBytesTotal)

	test.Equals(t, uint64(0), formatter.filesFinished)
	test.Equals(t, expectedFilesTotal, formatter.filesTotal)
	test.Equals(t, uint64(0), formatter.allBytesWritten)
	test.Equals(t, expectedAllBytesTotal, formatter.allBytesTotal)
}

func TestFirstProgressOnAFile(t *testing.T) {
	expectedBytesWritten := uint64(5)
	expectedBytesTotal := uint64(100)

	formatter := NewProgress(&stubPrinter{})
	formatter.AddFile(100)
	formatter.AddProgress("test", expectedBytesWritten, expectedBytesTotal)

	test.Equals(t, 1, len(formatter.progressInfoMap))
	test.Equals(t, expectedBytesWritten, formatter.progressInfoMap["test"].bytesWritten)
	test.Equals(t, expectedBytesTotal, formatter.progressInfoMap["test"].bytesTotal)
}

func TestSubsequentProgressOnAFile(t *testing.T) {
	fileSize := uint64(100)
	bytesWrittenOnOtherFile := uint64(20)
	firstBytesWritten := uint64(5)
	secondBytesWritten := uint64(10)
	expectedBytesWritten := firstBytesWritten + secondBytesWritten
	expectedAllBytesWritten := bytesWrittenOnOtherFile + firstBytesWritten + secondBytesWritten

	formatter := NewProgress(&stubPrinter{})
	formatter.AddFile(100)
	formatter.AddFile(50)
	formatter.allBytesWritten = bytesWrittenOnOtherFile
	formatter.AddProgress("test", firstBytesWritten, fileSize)
	formatter.AddProgress("test", secondBytesWritten, fileSize)

	actualBytesWritten := formatter.progressInfoMap["test"].bytesWritten
	actualAllBytesWritten := formatter.allBytesWritten
	if actualBytesWritten != expectedBytesWritten {
		t.Errorf("Error on Subsequent Progress. Wrong bytesWritten value. Have %v, wanted %v", actualBytesWritten, expectedBytesWritten)
	}
	if actualAllBytesWritten != expectedAllBytesWritten {
		t.Errorf("Error on Subsequent Progress. Wrong allBytesWritten value. Have %v, wanted %v", actualAllBytesWritten, expectedBytesWritten)
	}
}

func TestLastProgressOnAFile(t *testing.T) {
	fileSize := uint64(100)
	formatter := NewProgress(&stubPrinter{})
	formatter.AddFile(fileSize)
	formatter.AddProgress("test", 30, fileSize)
	formatter.AddProgress("test", 35, fileSize)
	formatter.AddProgress("test", 35, fileSize)

	actualSize := len(formatter.progressInfoMap)
	actualFinished := formatter.filesFinished
	actualAllBytesWritten := formatter.allBytesWritten
	if actualSize != 0 {
		t.Errorf("Error on Last Progress. Wrong map size. Have %v, wanted 0", actualSize)
	}
	if actualFinished != 1 {
		t.Errorf("Error on Last Progress. Wrong filesFinished. Have %v, wanted 1", actualFinished)
	}
	if actualAllBytesWritten != fileSize {
		t.Errorf("Error on Last Progress. Wrong AllBytesWritten. Have %v, wanted %v", actualAllBytesWritten, fileSize)
	}
}

func TestLastProgressOnLastFile(t *testing.T) {
	fileSize := uint64(100)
	formatter := NewProgress(&stubPrinter{})
	formatter.AddFile(fileSize)
	formatter.AddFile(50)
	formatter.AddProgress("test1", 50, 50)
	formatter.AddProgress("test2", 50, fileSize)
	expectedAllBytesWritten := formatter.allBytesTotal
	expectedFinished := formatter.filesTotal

	formatter.AddProgress("test2", 50, fileSize)

	actualFinished := formatter.filesFinished
	actualAllBytesWritten := formatter.allBytesWritten
	if actualFinished != expectedFinished {
		t.Errorf("Error on Last Progress of all. Wrong filesFinished. Have %v, wanted %v", actualFinished, expectedFinished)
	}
	if actualAllBytesWritten != expectedAllBytesWritten {
		t.Errorf("Error on Last Progress of all. Wrong AllBytesWritten. Have %v, wanted %v", actualAllBytesWritten, expectedAllBytesWritten)
	}
}

// func TestPrintSummaryOnSuccess(t *testing.T) {
// 	expectedText := "Summary: Restored 2 Files (100 B) in 0:00"
// 	formatter := NewProgress(&stubPrinter{})
// 	formatter.AddFile(50)
// 	formatter.AddFile(50)
// 	formatter.filesFinished = 2
// 	formatter.allBytesWritten = 100

// 	actualText := formatter.Finish()

// 	if actualText != expectedText {
// 		t.Errorf("Error on FormatSummary. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
// 	}
// }

// func TestPrintSummaryOnErrors(t *testing.T) {
// 	expectedText := "Summary: Restored 1 / 2 Files (70 B / 100 B) in 0:00"
// 	formatter := NewProgress(&stubPrinter{})
// 	formatter.AddFile(50)
// 	formatter.AddFile(50)
// 	formatter.filesFinished = 1
// 	formatter.allBytesWritten = 70

// 	actualText := formatter.Finish()

// 	if actualText != expectedText {
// 		t.Errorf("Error on FormatSummary. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
// 	}
// }
