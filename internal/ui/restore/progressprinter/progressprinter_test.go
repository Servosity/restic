package progressprinter

import (
	"bytes"
	"os"
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

func TestCalculatePercent(test *testing.T) {
	expectedValues := []float64{42.230000000000004, 1.1, 0.12, 0.0, 0.0, 100.0}
	dones := []int64{4_223, 11, 12, 0, 123, 42}
	froms := []int64{10_000, 1_000, 10_000, 10, 0, 42}
	actualValues := [6]float64{}

	for i := 0; i < len(expectedValues); i++ {
		actualValues[i] = calculatePercent(dones[i], froms[i])
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
	printer := New(os.Stdout)

	actualFilesTotal := printer.filesTotal
	actualAllBytesTotal := printer.allBytesTotal
	if printer.filesFinished != 0 {
		t.Errorf("Printer initialization error on filesFinished. Have %v, wanted 0", printer.filesFinished)
	}
	if actualFilesTotal != 0 {
		t.Errorf("Printer initialization error on filesTotal. Have %v, wanted 0", actualFilesTotal)
	}
	if actualAllBytesTotal != 0 {
		t.Errorf("Printer initialization error on allBytesTotal. Have %v, wanted 0", actualAllBytesTotal)
	}
}

func TestAddFile(t *testing.T) {
	expectedFilesTotal := int64(1)
	expectedAllBytesTotal := int64(100)
	printer := New(os.Stdout)

	printer.AddFile(expectedAllBytesTotal)

	actualFilesTotal := printer.filesTotal
	actualAllBytesTotal := printer.allBytesTotal
	if printer.filesFinished != 0 {
		t.Errorf("Printer initialization error on filesFinished. Have %v, wanted 0", printer.filesFinished)
	}
	if actualFilesTotal != expectedFilesTotal {
		t.Errorf("Printer initialization error on filesTotal. Have %v, wanted %v", actualFilesTotal, expectedFilesTotal)
	}
	if actualAllBytesTotal != expectedAllBytesTotal {
		t.Errorf("Printer initialization error on allBytesTotal. Have %v, wanted %v", actualAllBytesTotal, expectedAllBytesTotal)
	}
}

func TestFirstProgressOnAFile(t *testing.T) {
	expectedBytesWritten := int64(5)
	expectedBytesTotal := int64(17)
	printer := New(os.Stdout)
	printer.AddFile(100)

	printer.LogProgress("test", expectedBytesWritten, expectedBytesTotal)

	actualBytesWritten := printer.progressInfoMap["test"].bytesWritten
	actualBytesTotal := printer.progressInfoMap["test"].bytesTotal
	actualSize := len(printer.progressInfoMap)
	if actualBytesTotal != expectedBytesTotal {
		t.Errorf("Error on First Progress. Wrong bytesTotal value. Have %v, wanted %v", actualBytesTotal, expectedBytesTotal)
	}
	if actualBytesWritten != expectedBytesWritten {
		t.Errorf("Error on First Progress. Wrong bytesWritten value. Have %v, wanted %v", actualBytesWritten, expectedBytesWritten)
	}
	if actualSize != 1 {
		t.Errorf("Error on First Progress. Wrong map size. Have %v, wanted 1", actualSize)
	}
}

func TestSubsequentProgressOnAFile(t *testing.T) {
	fileSize := int64(100)
	bytesWrittenOnOtherFile := int64(20)
	firstBytesWritten := int64(5)
	secondBytesWritten := int64(10)
	expectedBytesWritten := firstBytesWritten + secondBytesWritten
	expectedAllBytesWritten := bytesWrittenOnOtherFile + firstBytesWritten + secondBytesWritten
	printer := New(os.Stdout)
	printer.AddFile(100)
	printer.AddFile(50)
	printer.allBytesWritten = bytesWrittenOnOtherFile
	printer.LogProgress("test", firstBytesWritten, fileSize)

	printer.LogProgress("test", secondBytesWritten, fileSize)

	actualBytesWritten := printer.progressInfoMap["test"].bytesWritten
	actualAllBytesWritten := printer.allBytesWritten
	if actualBytesWritten != expectedBytesWritten {
		t.Errorf("Error on Subsequent Progress. Wrong bytesWritten value. Have %v, wanted %v", actualBytesWritten, expectedBytesWritten)
	}
	if actualAllBytesWritten != expectedAllBytesWritten {
		t.Errorf("Error on Subsequent Progress. Wrong allBytesWritten value. Have %v, wanted %v", actualAllBytesWritten, expectedBytesWritten)
	}
}

func TestLastProgressOnAFile(t *testing.T) {
	fileSize := int64(100)
	printer := New(os.Stdout)
	printer.AddFile(fileSize)
	printer.LogProgress("test", int64(30), fileSize)
	printer.LogProgress("test", int64(35), fileSize)

	printer.LogProgress("test", int64(35), fileSize)

	actualSize := len(printer.progressInfoMap)
	actualFinished := printer.filesFinished
	actualAllBytesWritten := printer.allBytesWritten
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
	fileSize := int64(100)
	printer := New(os.Stdout)
	printer.AddFile(fileSize)
	printer.AddFile(50)
	printer.LogProgress("test1", int64(50), int64(50))
	printer.LogProgress("test2", int64(50), fileSize)
	expectedAllBytesWritten := printer.allBytesTotal
	expectedFinished := printer.filesTotal

	printer.LogProgress("test2", int64(50), fileSize)

	actualFinished := printer.filesFinished
	actualAllBytesWritten := printer.allBytesWritten
	if actualFinished != expectedFinished {
		t.Errorf("Error on Last Progress of all. Wrong filesFinished. Have %v, wanted %v", actualFinished, expectedFinished)
	}
	if actualAllBytesWritten != expectedAllBytesWritten {
		t.Errorf("Error on Last Progress of all. Wrong AllBytesWritten. Have %v, wanted %v", actualAllBytesWritten, expectedAllBytesWritten)
	}
}

func TestPrintingWithSpecificValues(t *testing.T) {
	var writer bytes.Buffer
	printer := New(&writer)
	printer.filesTotal = 10
	printer.filesFinished = 6
	printer.allBytesTotal = 1024
	printer.allBytesWritten = 512
	expectedPrint := "[0:01]  6 / 10 Files,  512 B / 1.00 KiB,  50.00 %   \r"
	time.Sleep(time.Second)

	print(printer)

	printed := writer.String()
	if printed != expectedPrint {
		t.Errorf("Printer Error.\nHave   %q\nwanted %q", printed, expectedPrint)
	}
}

func TestPrintingForThe4CasesOfProgress(t *testing.T) {
	writers := [4]bytes.Buffer{}
	names := []string{"test", "test", "test", "test2"}
	dones := []int64{10, 70, 10, 10}
	froms := []int64{90, 90, 90, 10}
	expectedValues := []string{
		"[0:00]  0 / 2 Files,  10 B / 100 B,  10.00 %   \r",
		"[0:00]  0 / 2 Files,  80 B / 100 B,  80.00 %   \r",
		"[0:00]  1 / 2 Files,  90 B / 100 B,  90.00 %   \r",
		"[0:00]  2 / 2 Files,  100 B / 100 B,  100.00 %   \r",
	}
	actualValues := [4]string{}
	printer := New(os.Stdout)
	printer.AddFile(90)
	printer.AddFile(10)

	for i := 0; i < len(expectedValues); i++ {
		printer.progresswriter = &writers[i]
		printer.LogProgress(names[i], dones[i], froms[i])
		actualValues[i] = writers[i].String()
	}

	for i := 0; i < len(expectedValues); i++ {
		actual := actualValues[i]
		expected := expectedValues[i]
		if expected != actual {
			t.Errorf("Printer Error.\nHave   %q\nwanted %q", actual, expected)
		}
	}
}

func TestPrintSummaryOnSuccess(t *testing.T) {
	expectedPrint := "  \rSummary: Restored 2 Files (100 B) in 0:01\n"
	var writer bytes.Buffer
	printer := New(&writer)
	printer.AddFile(50)
	printer.AddFile(50)
	printer.filesFinished = 2
	printer.allBytesWritten = 100
	printer.lastPrinted = "--"

	printer.PrintSummary()

	actualPrinted := writer.String()
	if expectedPrint != actualPrinted {
		t.Errorf("PrintSummary Error.\nHave   %q\nwanted %q", actualPrinted, expectedPrint)
	}
}

func TestPrintSummaryOnErrors(t *testing.T) {
	expectedPrint := "Summary: Restored 1 / 2 Files (70 B / 100 B) in 0:01\n"
	var writer bytes.Buffer
	printer := New(&writer)
	printer.AddFile(50)
	printer.AddFile(50)
	printer.filesFinished = 1
	printer.allBytesWritten = 70

	printer.PrintSummary()

	actualPrinted := writer.String()
	if expectedPrint != actualPrinted {
		t.Errorf("PrintSummary Error.\nHave   %q\nwanted %q", actualPrinted, expectedPrint)
	}
}
