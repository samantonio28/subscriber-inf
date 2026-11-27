package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	WithFields(fields map[string]any) Entry
	WithError(err error) Entry
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
}

type LogrusLogger struct {
	log      *logrus.Logger
	extended bool
}

type Fields *logrus.Fields
type Entry *logrus.Entry

// ExtendedMode читает env-флаг расширенного логирования (ЛР №6):
// LOG_MODE=extended → Debug-уровень и расширенные поля access-лога.
// По умолчанию — стандартное логирование (Info).
func ExtendedMode() bool {
	return os.Getenv("LOG_MODE") == "extended"
}

// NewLogrusLoggerMode создаёт логгер в одном из двух режимов (ЛР №6):
//   - extended=false (по умолчанию): уровень Info, компактная запись запроса;
//   - extended=true: уровень Debug (в лог попадают Debug-сообщения usecase-слоя)
//     и расширенный access-лог (query-параметры, размеры тел).
//
// Переключение через LOG_MODE=extended позволяет измерять накладные расходы
// логирования в нагрузочных прогонах.
func NewLogrusLoggerMode(accessLogPath string, extended bool) (*LogrusLogger, error) {
	var Log LogrusLogger
	Log.log = logrus.New()
	Log.log.SetFormatter(&logrus.JSONFormatter{})

	if extended {
		Log.log.SetLevel(logrus.DebugLevel)
	} else {
		Log.log.SetLevel(logrus.InfoLevel)
	}

	if err := os.MkdirAll(filepath.Dir(accessLogPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	accessFile, err := os.OpenFile(accessLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	Log.log.SetOutput(accessFile)
	Log.extended = extended

	return &Log, nil
}

// NewLogrusLogger — прежний конструктор (стандартное логирование).
func NewLogrusLogger(accessLogPath string) (*LogrusLogger, error) {
	return NewLogrusLoggerMode(accessLogPath, false)
}

func (l *LogrusLogger) WithFields(fields map[string]any) Entry {
	return l.log.WithFields(fields)
}

func (l *LogrusLogger) WithError(err error) Entry {
	return l.log.WithError(err)
}

func (l *LogrusLogger) Debug(args ...any) {
	l.log.Debug(args...)
}

func (l *LogrusLogger) Info(args ...any) {
	l.log.Info(args...)
}

func (l *LogrusLogger) Warn(args ...any) {
	l.log.Warn(args...)
}

func (l *LogrusLogger) Error(args ...any) {
	l.log.Error(args...)
}
