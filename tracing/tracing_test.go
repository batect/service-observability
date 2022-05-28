// Copyright 2019-2022 Charles Korn.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing_test

import (
	"net/http/httptest"

	"github.com/batect/services-common/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Naming HTTP request spans", func() {
	req := httptest.NewRequest("PUT", "/blah", nil)

	Describe("given an operation name is provided", func() {
		var name string

		BeforeEach(func() {
			name = tracing.NameHTTPRequestSpan("Server", req)
		})

		It("includes the operation name, HTTP method and URL in the name", func() {
			Expect(name).To(Equal("Server: PUT /blah"))
		})
	})

	Describe("given an operation name is not provided", func() {
		var name string

		BeforeEach(func() {
			name = tracing.NameHTTPRequestSpan("", req)
		})

		It("does not include the operation name in the span name", func() {
			Expect(name).To(Equal("PUT /blah"))
		})
	})
})
