package testutil

import "github.com/samantonio28/subscriber-inf/internal/logger"

// NopLogger — заглушка интерфейса logger.Logger: ничего не пишет,
// With* методы возвращают nil. Используется в тестах, где логгер
// не является предметом проверки.
type NopLogger struct{}

var _ logger.Logger = (*NopLogger)(nil)

func (n *NopLogger) WithFields(fields map[string]any) logger.Entry { return nil }
func (n *NopLogger) WithError(err error) logger.Entry              { return nil }
func (n *NopLogger) Debug(args ...any)                             {}
func (n *NopLogger) Info(args ...any)                              {}
func (n *NopLogger) Warn(args ...any)                              {}
func (n *NopLogger) Error(args ...any)                             {}
