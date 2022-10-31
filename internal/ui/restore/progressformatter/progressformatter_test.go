package progressformatter

import (
	"testing"
	"time"
)

func TestFormatBytesInBestUnit(test *testing.T) {
	k := float64(1024)
	m := float64(1024) * k
	g := float64(1024) * m
	t := float64(1024) * g
	expectedValues := []string{"42 B", "4.21 KiB", "42.34 MiB", "234.78 GiB", "13.42 TiB"}
	inputs := []int64{42, int64(k * 4.21), int64(m * 42.34), int64(g * 234.78), int64(t * 13.42)}
	actualValues := [5]string{}

	for i := 0; i < len(expectedValues); i++ {
		actualValues[i] = formatBytesInBestUnit(inputs[i])
	}

	for i := 0; i < len(expectedValues); i++ {
		actual := actualValues[i]
		expected := expectedValues[i]
		if expected != actual {
			test.Errorf("Bytes Formatter Error. Have %v, wanted %v", actual, expected)
		}
	}
}

func TestFormatLeftTime(test *testing.T) {
	second := int64(time.Second)
	minute := int64(time.Minute)
	hour := int64(time.Hour)
	expectedValues := []string{"0:07", "3:06", "2:01:56"}
	inputs := []int64{7 * second, 3*minute + 6*second, 2*hour + minute + 56*second}
	actualValues := [3]string{}

	for i := 0; i < len(expectedValues); i++ {
		actualValues[i] = formatLeftTime(inputs[i])
	}

	for i := 0; i < len(expectedValues); i++ {
		actual := actualValues[i]
		expected := expectedValues[i]
		if expected != actual {
			test.Errorf("Left time Formatter Error. Have %v, wanted %v", actual, expected)
		}
	}
}

func TestFormatPercent(test *testing.T) {
	expectedValues := []string{"42.23 %", "1.10 %", "0.12 %", "0.00 %", "0.00 %", "100.00 %"}
	dones := []int64{4_223, 11, 12, 0, 123, 42}
	froms := []int64{10_000, 1_000, 10_000, 10, 0, 42}
	actualValues := [6]string{}

	for i := 0; i < len(expectedValues); i++ {
		actualValues[i] = formatPercent(dones[i], froms[i])
	}

	for i := 0; i < len(expectedValues); i++ {
		actual := actualValues[i]
		expected := expectedValues[i]
		if expected != actual {
			test.Errorf("Calculate Percent Error. Have %v, wanted %v", actual, expected)
		}
	}
}

func TestNew(t *testing.T) {
	formatter := NewFormatter()

	actualFilesTotal := formatter.filesTotal
	actualAllBytesTotal := formatter.allBytesTotal
	if formatter.filesFinished != 0 {
		t.Errorf("Formatter initialization error on filesFinished. Have %v, wanted 0", formatter.filesFinished)
	}
	if actualFilesTotal != 0 {
		t.Errorf("Formatter initialization error on filesTotal. Have %v, wanted 0", actualFilesTotal)
	}
	if actualAllBytesTotal != 0 {
		t.Errorf("Formatter initialization error on allBytesTotal. Have %v, wanted 0", actualAllBytesTotal)
	}
}

func TestAddFile(t *testing.T) {
	expectedFilesTotal := int64(1)
	expectedAllBytesTotal := int64(100)
	formatter := NewFormatter()

	formatter.AddFile(expectedAllBytesTotal)

	actualFilesTotal := formatter.filesTotal
	actualAllBytesTotal := formatter.allBytesTotal
	if formatter.filesFinished != 0 {
		t.Errorf("Formatter initialization error on filesFinished. Have %v, wanted 0", formatter.filesFinished)
	}
	if actualFilesTotal != expectedFilesTotal {
		t.Errorf("Formatter initialization error on filesTotal. Have %v, wanted %v", actualFilesTotal, expectedFilesTotal)
	}
	if actualAllBytesTotal != expectedAllBytesTotal {
		t.Errorf("Formatter initialization error on allBytesTotal. Have %v, wanted %v", actualAllBytesTotal, expectedAllBytesTotal)
	}
}

func TestFirstProgressOnAFile(t *testing.T) {
	expectedBytesWritten := int64(5)
	expectedBytesTotal := int64(100)
	expectedText := "  [0:00]  0 / 1 Files,  5 B / 100 B,  5.00 %  "
	formatter := NewFormatter()
	formatter.AddFile(100)

	actualText := formatter.FormatProgress("test", expectedBytesWritten, expectedBytesTotal)

	actualBytesWritten := formatter.progressInfoMap["test"].bytesWritten
	actualBytesTotal := formatter.progressInfoMap["test"].bytesTotal
	actualSize := len(formatter.progressInfoMap)
	if actualBytesTotal != expectedBytesTotal {
		t.Errorf("Error on First Progress. Wrong bytesTotal value. Have %v, wanted %v", actualBytesTotal, expectedBytesTotal)
	}
	if actualBytesWritten != expectedBytesWritten {
		t.Errorf("Error on First Progress. Wrong bytesWritten value. Have %v, wanted %v", actualBytesWritten, expectedBytesWritten)
	}
	if actualSize != 1 {
		t.Errorf("Error on First Progress. Wrong map size. Have %v, wanted 1", actualSize)
	}
	if actualText != expectedText {
		t.Errorf("Error on First Progress. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
	}
}

func TestSubsequentProgressOnAFile(t *testing.T) {
	fileSize := int64(100)
	bytesWrittenOnOtherFile := int64(20)
	firstBytesWritten := int64(5)
	secondBytesWritten := int64(10)
	expectedBytesWritten := firstBytesWritten + secondBytesWritten
	expectedAllBytesWritten := bytesWrittenOnOtherFile + firstBytesWritten + secondBytesWritten
	expectedText := "  [0:00]  0 / 2 Files,  35 B / 150 B,  23.33 %  "

	formatter := NewFormatter()
	formatter.AddFile(100)
	formatter.AddFile(50)
	formatter.allBytesWritten = bytesWrittenOnOtherFile
	formatter.FormatProgress("test", firstBytesWritten, fileSize)

	actualText := formatter.FormatProgress("test", secondBytesWritten, fileSize)

	actualBytesWritten := formatter.progressInfoMap["test"].bytesWritten
	actualAllBytesWritten := formatter.allBytesWritten
	if actualBytesWritten != expectedBytesWritten {
		t.Errorf("Error on Subsequent Progress. Wrong bytesWritten value. Have %v, wanted %v", actualBytesWritten, expectedBytesWritten)
	}
	if actualAllBytesWritten != expectedAllBytesWritten {
		t.Errorf("Error on Subsequent Progress. Wrong allBytesWritten value. Have %v, wanted %v", actualAllBytesWritten, expectedBytesWritten)
	}
	if actualText != expectedText {
		t.Errorf("Error on Subsequent Progress. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
	}
}

func TestLastProgressOnAFile(t *testing.T) {
	fileSize := int64(100)
	expectedText := "  [0:00]  1 / 1 Files,  100 B / 100 B,  100.00 %  "
	formatter := NewFormatter()
	formatter.AddFile(fileSize)
	formatter.FormatProgress("test", int64(30), fileSize)
	formatter.FormatProgress("test", int64(35), fileSize)

	actualText := formatter.FormatProgress("test", int64(35), fileSize)

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
	if actualText != expectedText {
		t.Errorf("Error on Last Progress. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
	}
}

func TestLastProgressOnLastFile(t *testing.T) {
	fileSize := int64(100)
	expectedText := "  [0:00]  2 / 2 Files,  150 B / 150 B,  100.00 %  "
	formatter := NewFormatter()
	formatter.AddFile(fileSize)
	formatter.AddFile(50)
	formatter.FormatProgress("test1", int64(50), int64(50))
	formatter.FormatProgress("test2", int64(50), fileSize)
	expectedAllBytesWritten := formatter.allBytesTotal
	expectedFinished := formatter.filesTotal

	actualText := formatter.FormatProgress("test2", int64(50), fileSize)

	actualFinished := formatter.filesFinished
	actualAllBytesWritten := formatter.allBytesWritten
	if actualFinished != expectedFinished {
		t.Errorf("Error on Last Progress of all. Wrong filesFinished. Have %v, wanted %v", actualFinished, expectedFinished)
	}
	if actualAllBytesWritten != expectedAllBytesWritten {
		t.Errorf("Error on Last Progress of all. Wrong AllBytesWritten. Have %v, wanted %v", actualAllBytesWritten, expectedAllBytesWritten)
	}
	if actualText != expectedText {
		t.Errorf("Error on Last Progress of all. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
	}
}

func TestPrintSummaryOnSuccess(t *testing.T) {
	expectedText := "Summary: Restored 2 Files (100 B) in 0:00"
	formatter := NewFormatter()
	formatter.AddFile(50)
	formatter.AddFile(50)
	formatter.filesFinished = 2
	formatter.allBytesWritten = 100

	actualText := formatter.FormatSummary()

	if actualText != expectedText {
		t.Errorf("Error on FormatSummary. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
	}
}

func TestPrintSummaryOnErrors(t *testing.T) {
	expectedText := "Summary: Restored 1 / 2 Files (70 B / 100 B) in 0:00"
	formatter := NewFormatter()
	formatter.AddFile(50)
	formatter.AddFile(50)
	formatter.filesFinished = 1
	formatter.allBytesWritten = 70

	actualText := formatter.FormatSummary()

	if actualText != expectedText {
		t.Errorf("Error on FormatSummary. Wrong text.\nHave   %q\nwanted %q", actualText, expectedText)
	}
}
