package domains

import (
	"context"

	"github.com/sirupsen/logrus"
)

func GetLogger(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(ContextKeyLogger{}).(*logrus.Entry); ok {
		return logger
	}

	return logrus.NewEntry(logrus.New())
}
