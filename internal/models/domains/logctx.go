package domains

import (
	"context"

	"github.com/sirupsen/logrus"
)

const LoggingLevel = logrus.DebugLevel

func GetLogger(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(ContextKeyLogger{}).(*logrus.Entry); ok {
		return logger
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(LoggingLevel)

	return logrus.NewEntry(logger)
}
