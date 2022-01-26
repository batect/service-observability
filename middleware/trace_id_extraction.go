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

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

func extractTraceID(req *http.Request) string {
	span := trace.SpanFromContext(req.Context())

	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}

	return fmt.Sprintf("autogenerated-%s", uuid.New().String())
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	//nolint:forcetypeassert
	return ctx.Value(traceIDKey).(string)
}

func TraceIDExtractionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := ContextWithTraceID(req.Context(), extractTraceID(req))
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
