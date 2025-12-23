package writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCSVWriter_NewCSVWriter_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")

	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}
	if cw == nil {
		t.Fatal("NewCSVWriter returned nil")
	}
	defer cw.Close()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("CSV file was not created")
	}
}

func TestCSVWriter_NewCSVWriter_WritesHeaders(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")
	obsVersion := "30.0.0"
	streamDomain := "live.twitch.tv"

	cw, err := NewCSVWriter(filename, obsVersion, streamDomain)
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}
	cw.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, obsVersion) {
		t.Error("CSV header should contain OBS version")
	}
	if !strings.Contains(contentStr, streamDomain) {
		t.Error("CSV header should contain stream domain")
	}
	if !strings.Contains(contentStr, "timestamp") {
		t.Error("CSV header should contain column names")
	}
}

func TestCSVWriter_WriteMetrics_SingleRow(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}

	data := MetricsData{
		Timestamp:           time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsRTT:              50 * time.Millisecond,
		GoogleRTT:           25 * time.Millisecond,
		StreamActive:        true,
		OutputBytes:         1024.0,
		OutputSkippedFrames: 5.0,
		OutputFrames:        100.0,
		ObsCpuUsage:         15.5,
		ObsMemoryUsage:      512.0,
		SystemCpuUsage:      45.2,
		SystemMemoryUsage:   60.0,
	}

	err = cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}
	cw.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines (header info, column headers, data), got %d", len(lines))
	}

	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "2025-12-23") {
		t.Error("Data row should contain timestamp")
	}
	if !strings.Contains(lastLine, "50.00") {
		t.Error("Data row should contain obs RTT")
	}
	if !strings.Contains(lastLine, "true") {
		t.Error("Data row should contain stream active status")
	}
}

func TestCSVWriter_WriteMetrics_MultipleRows(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		data := MetricsData{
			Timestamp:    time.Date(2025, 12, 23, 10, i, 0, 0, time.UTC),
			ObsRTT:       time.Duration(i*10) * time.Millisecond,
			StreamActive: true,
		}
		err = cw.WriteMetrics(data)
		if err != nil {
			t.Fatalf("WriteMetrics failed on iteration %d: %v", i, err)
		}
	}
	cw.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 5 {
		t.Fatalf("Expected at least 5 lines (header info, column headers, 3 data rows), got %d", len(lines))
	}
}

func TestCSVWriter_WriteMetrics_WithErrors(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}

	obsErr := fmt.Errorf("obs ping failed")
	streamErr := fmt.Errorf("stream error")

	data := MetricsData{
		Timestamp:    time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsPingError: obsErr,
		StreamError:  streamErr,
		StreamActive: false,
	}

	err = cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}
	cw.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "obs_ping") {
		t.Error("Error column should contain obs ping error")
	}
	if !strings.Contains(contentStr, "stream") {
		t.Error("Error column should contain stream error")
	}
}

func TestCSVWriter_NewCSVWriter_InvalidPath(t *testing.T) {
	filename := "/invalid/path/that/does/not/exist/test.csv"

	_, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")

	if err == nil {
		t.Error("Expected error when creating file in invalid path")
	}
}

func TestCSVWriter_Close_FlushesData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}

	data := MetricsData{
		Timestamp:    time.Now(),
		StreamActive: true,
	}
	_ = cw.WriteMetrics(data)

	err = cw.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file after close: %v", err)
	}

	if len(content) == 0 {
		t.Error("File should contain data after close")
	}
}

func TestCSVWriter_WriteMetrics_ZeroValues(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}

	data := MetricsData{
		Timestamp:           time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsRTT:              0,
		GoogleRTT:           0,
		StreamActive:        false,
		OutputBytes:         0,
		OutputSkippedFrames: 0,
		OutputFrames:        0,
		ObsCpuUsage:         0,
		ObsMemoryUsage:      0,
		SystemCpuUsage:      0,
		SystemMemoryUsage:   0,
	}

	err = cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}
	cw.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}
}

func TestCSVWriter_WriteMetrics_SpecialCharactersInError(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}

	data := MetricsData{
		Timestamp:    time.Now(),
		StreamError:  fmt.Errorf("error with \"quotes\" and, commas"),
		StreamActive: false,
	}

	err = cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}
	cw.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	if len(content) == 0 {
		t.Error("CSV file should contain data")
	}
}

func TestCSVWriter_WriteMetrics_HighRTTValues(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	cw, err := NewCSVWriter(filename, "30.0.0", "live.twitch.tv")
	if err != nil {
		t.Fatalf("NewCSVWriter failed: %v", err)
	}

	data := MetricsData{
		Timestamp:    time.Now(),
		ObsRTT:       1500 * time.Millisecond,
		GoogleRTT:    2000 * time.Millisecond,
		StreamActive: true,
	}

	err = cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}
	cw.Close()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "1500.00") {
		t.Error("CSV should contain high RTT value in milliseconds")
	}
}
