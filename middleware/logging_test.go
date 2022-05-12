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

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/batect/services-common/middleware"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Logging middleware", func() {
	var logger *logrus.Logger
	var hook *test.Hook

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		logger.Level = logrus.DebugLevel
	})

	Context("when the request starts", func() {
		BeforeEach(func() {
			m := middleware.LoggerMiddleware(logger, "my-project", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
			m.ServeHTTP(nil, createTestRequest())
		})

		It("logs a single message", func() {
			Expect(hook.Entries).To(HaveLen(1))
		})

		It("logs that message at info level", func() {
			Expect(hook.LastEntry().Level).To(Equal(logrus.DebugLevel))
		})

		It("logs a message indicating that the request is being processed", func() {
			Expect(hook.LastEntry().Message).To(Equal("Processing request."))
		})

		It("adds the expected trace ID to the message", func() {
			Expect(hook.LastEntry().Data).To(HaveKeyWithValue("trace", "projects/my-project/traces/abc-123-def"))
		})
	})

	Context("when the request logs a message", func() {
		BeforeEach(func() {
			m := middleware.LoggerMiddleware(logger, "my-project", http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				hook.Reset()

				logger := middleware.LoggerFromContext(r.Context())
				logger.Info("Inside request.")
			}))

			m.ServeHTTP(nil, createTestRequest())
		})

		It("logs the message to the provided logger", func() {
			Expect(hook.Entries).To(HaveLen(1))
		})

		It("adds the expected trace ID to the message", func() {
			Expect(hook.LastEntry().Data).To(HaveKeyWithValue("trace", "projects/my-project/traces/abc-123-def"))
		})
	})
})

func createTestRequest() *http.Request {
	req := httptest.NewRequest("PUT", "/blah", strings.NewReader("test"))

	return req.WithContext(middleware.ContextWithTraceID(req.Context(), "abc-123-def"))
}
