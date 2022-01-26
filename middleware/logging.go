// Copyright 2019-2022 Charles Korn.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func loggerForRequest(logger logrus.FieldLogger, projectID string, req *http.Request) logrus.FieldLogger {
	traceID := TraceIDFromContext(req.Context())

	return logger.WithFields(logrus.Fields{
		"trace": fmt.Sprintf("projects/%s/traces/%s", projectID, traceID),
	})
}

func ContextWithLogger(ctx context.Context, logger logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) logrus.FieldLogger {
	//nolint:forcetypeassert
	return ctx.Value(loggerKey).(logrus.FieldLogger)
}

func LoggerMiddleware(baseLogger logrus.FieldLogger, projectID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger := loggerForRequest(baseLogger, projectID, req)
		logger.Debug("Processing request.")

		ctx := ContextWithLogger(req.Context(), logger)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
