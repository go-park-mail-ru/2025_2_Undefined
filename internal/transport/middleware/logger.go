package middleware

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/sirupsen/logrus"
)

func AccessLogMiddleware(logger *logrus.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%016x", rand.Int())[:10]

		middlewareLogger := logger.WithFields(logrus.Fields{
			"request_id":  requestID,
			"method":      r.Method,
			"remote_addr": r.RemoteAddr,
			"path":        r.URL.Path,
		})

		contextLogger := logrus.NewEntry(logger).WithField("request_id", requestID)
		ctx := context.WithValue(r.Context(), domains.ContextKeyLogger{}, contextLogger)

		startTime := time.Now()
		middlewareLogger.Info("request started")

		defer func() {
			duration := time.Since(startTime)
			middlewareLogger.WithField("duration", duration).Info("request completed")
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
