package jsonlog

import (
	"bytes"
	"testing"
)

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	minLevel := LevelInfo

	logger := NewLogger(&buf, minLevel)

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.out != &buf {
		t.Errorf("expected logger.out to be %v, got %v", &buf, logger.out)
	}

	if logger.minLevel != minLevel {
		t.Errorf("expected minLevel %d, got %d", minLevel, logger.minLevel)
	}
}
