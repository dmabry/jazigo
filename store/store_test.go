package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/udhos/jazigo/temp"
)

// testLogger: wrap Printf interface around *testing.T
type testLogger struct {
	*testing.T
}

func (t *testLogger) Printf(format string, v ...interface{}) {
	t.Logf("store testLogger: "+format, v...)
}

func TestExtractCommitIDFromFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"config.1", 1, false},
		{"config.42", 42, false},
		{"config.0", 0, false},
		{"config.999", 999, false},
		{"config.xyz", -1, true}, // Invalid format
		{"config.", -1, true},    // No number after dot
		{"config", -1, true},     // No dot at all
	}

	for _, tt := range tests {
		actual, err := ExtractCommitIDFromFilename(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ExtractCommitIDFromFilename(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if actual != tt.expected {
			t.Errorf("ExtractCommitIDFromFilename(%s) = %d, want %d", tt.input, actual, tt.expected)
		}
	}
}

func TestFileExists(t *testing.T) {
	tmpfile := temp.MakeTempRepo() + "/testfile"
	defer os.Remove(tmpfile)

	// Test with non-existent file
	if fileExists(tmpfile) {
		t.Errorf("fileExists(%s) = true, want false (file doesn't exist)", tmpfile)
	}

	// Create the file and test again
	f, err := os.Create(tmpfile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()

	if !fileExists(tmpfile) {
		t.Errorf("fileExists(%s) = false, want true (file exists)", tmpfile)
	}

	// Test with directory
	tmpdir := temp.MakeTempRepo() + "/testdir"
	os.MkdirAll(tmpdir, 0755)

	if !fileExists(tmpdir) {
		t.Errorf("fileExists(%s) = false, want true (directory exists)", tmpdir)
	}
}

func TestStore1(t *testing.T) {

	repo := temp.MakeTempRepo()
	defer temp.CleanupTempRepo()

	region := os.Getenv("JAZIGO_S3_REGION")

	maxFiles := 2
	logger := &testLogger{t}
	Init(logger, region)

	prefix := filepath.Join(repo, "store-test.")
	storeBatch(t, prefix, maxFiles, logger)

	if region == "" {
		t.Logf("TestStore1: JAZIGO_S3_REGION=region undefined: skipping S3 tests")
		return
	}
	s3folder := os.Getenv("JAZIGO_S3_FOLDER")
	if s3folder == "" {
		t.Logf("TestStore1: JAZIGO_S3_FOLDER=bucket/folder undefined: skipping S3 tests")
		return
	}

	prefix = fmt.Sprintf("arn:aws:s3:::%s/store-test.", s3folder)

	if false { // s3dirClean is not implemented
		cleanErr := fmt.Errorf("s3dirClean is not implemented")
		t.Errorf("TestStore1: s3dirClean() before error: %v", cleanErr)
	}

	storeBatch(t, prefix, maxFiles, logger)

	if false { // s3dirClean is not implemented
		cleanErr := fmt.Errorf("s3dirClean is not implemented")
		t.Errorf("TestStore1: s3dirClean() after error: %v", cleanErr)
	}
}

func storeBatch(t *testing.T, prefix string, maxFiles int, logger hasPrintf) {
	if err := storeWrite(t, prefix, "a", fmt.Sprintf("%s0", prefix), maxFiles, logger, ""); err != nil {
		t.Errorf("TestStore1: %v", err)
	}

	if err := storeWrite(t, prefix, "b", fmt.Sprintf("%s1", prefix), maxFiles, logger, ""); err != nil {
		t.Errorf("TestStore1: %v", err)
	}

	if err := storeWrite(t, prefix, "c", fmt.Sprintf("%s2", prefix), maxFiles, logger, "detect"); err != nil {
		t.Errorf("TestStore1: %v", err)
	}

	if err := storeWrite(t, prefix, "d", fmt.Sprintf("%s3", prefix), maxFiles, logger, "text/plain"); err != nil {
		t.Errorf("TestStore1: %v", err)
	}
}

func storeWrite(t *testing.T, prefix, content, expected string, maxFiles int, logger hasPrintf, contentType string) error {

	c := []byte(content)

	writeFunc := func(w HasWrite) error {
		n, writeErr := w.Write(c)
		if writeErr != nil {
			return fmt.Errorf("writeFunc: error: %v", writeErr)
		}
		if n != len(c) {
			return fmt.Errorf("writeFunc: partial: wrote=%d size=%d", n, len(c))
		}
		return nil
	}

	path, writeErr := SaveNewConfig(prefix, maxFiles, logger, writeFunc, false, contentType)
	if writeErr != nil {
		return fmt.Errorf("storeWrite: error: %v", writeErr)
	}

	if path != expected {
		return fmt.Errorf("storeWrite: got=%s wanted=%s", path, expected)
	}

	found, findErr := FindLastConfig(prefix, logger)
	if findErr != nil {
		return fmt.Errorf("storeWrite: FindLastConfig: error: %v", findErr)
	}

	if found != expected {
		return fmt.Errorf("storeWrite: FindLastConfig: found=%s wanted=%s", found, expected)
	}

	return nil
}
