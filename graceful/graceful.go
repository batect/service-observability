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

package graceful

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func RunServerWithGracefulShutdown(srv *http.Server) error {
	connectionDrainingFinished := shutdownOnInterrupt(srv)

	logrus.WithField("address", srv.Addr).Info("Server starting.")

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("could not start HTTP server: %w", err)
	}

	<-connectionDrainingFinished

	logrus.Info("Server gracefully stopped.")

	return nil
}

func shutdownOnInterrupt(srv *http.Server) chan struct{} {
	connectionDrainingFinished := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		logrus.Info("Interrupt received, draining connections...")

		if err := srv.Shutdown(context.Background()); err != nil {
			logrus.WithError(err).Error("Shutting down HTTP server failed.")
		}

		close(connectionDrainingFinished)
	}()

	return connectionDrainingFinished
}
