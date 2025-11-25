package domains

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantNew bool
	}{
		{
			name:    "Context without logger",
			ctx:     context.Background(),
			wantNew: true,
		},
		{
			name: "Context with logger",
			ctx: context.WithValue(context.Background(), ContextKeyLogger{},
				logrus.NewEntry(logrus.New())),
			wantNew: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := GetLogger(tt.ctx)

			assert.NotNil(t, logger)
			assert.IsType(t, &logrus.Entry{}, logger)
		})
	}
}
