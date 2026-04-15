package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Dir           string
	Filename      string
	RotateBy      string
	RetentionDays int
}

type rotatingWriter struct {
	config      Config
	mu          sync.Mutex
	current     string
	file        *os.File
	lastCleanup time.Time
}

func New(cfg Config) (*log.Logger, error) {
	writer, err := newRotatingWriter(cfg)
	if err != nil {
		return nil, err
	}

	multi := io.MultiWriter(os.Stdout, writer)
	return log.New(multi, "", log.LstdFlags|log.Lmicroseconds), nil
}

func newRotatingWriter(cfg Config) (*rotatingWriter, error) {
	if cfg.Dir == "" {
		cfg.Dir = "logs"
	}
	if cfg.Filename == "" {
		cfg.Filename = "app"
	}
	if cfg.RotateBy == "" {
		cfg.RotateBy = "day"
	}
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = 7
	}
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	writer := &rotatingWriter{config: cfg}
	if err := writer.cleanupOldFiles(time.Now()); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	name := w.filePath(now)
	if err := w.rotate(name); err != nil {
		return 0, err
	}
	if err := w.cleanupOldFiles(now); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

func (w *rotatingWriter) filePath(now time.Time) string {
	layout := "20060102"
	if w.config.RotateBy == "hour" {
		layout = "2006010215"
	}
	filename := fmt.Sprintf("%s-%s.log", w.config.Filename, now.Format(layout))
	return filepath.Join(w.config.Dir, filename)
}

func (w *rotatingWriter) rotate(path string) error {
	if w.current == path && w.file != nil {
		return nil
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	w.current = path
	w.file = file
	return nil
}

func (w *rotatingWriter) cleanupOldFiles(now time.Time) error {
	cleanupPoint := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	if !w.lastCleanup.IsZero() && w.lastCleanup.Equal(cleanupPoint) {
		return nil
	}

	entries, err := os.ReadDir(w.config.Dir)
	if err != nil {
		return fmt.Errorf("read log directory: %w", err)
	}

	prefix := w.config.Filename + "-"
	cutoff := now.AddDate(0, 0, -w.config.RetentionDays)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ts, ok := w.parseTimestamp(strings.TrimPrefix(name, prefix), name, prefix)
		if !ok {
			continue
		}
		if ts.Before(cutoff) {
			if err := os.Remove(filepath.Join(w.config.Dir, name)); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove expired log file %s: %w", name, err)
			}
		}
	}

	w.lastCleanup = cleanupPoint
	return nil
}

func (w *rotatingWriter) parseTimestamp(suffix string, originalName string, prefix string) (time.Time, bool) {
	if !strings.HasPrefix(originalName, prefix) || !strings.HasSuffix(originalName, ".log") {
		return time.Time{}, false
	}

	timestamp := strings.TrimSuffix(suffix, ".log")
	layout := "20060102"
	if w.config.RotateBy == "hour" {
		layout = "2006010215"
	}
	if len(timestamp) != len(layout) {
		return time.Time{}, false
	}

	parsed, err := time.ParseInLocation(layout, timestamp, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}
