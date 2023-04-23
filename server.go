// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

package muxpress

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type server struct {
	logger  logrus.FieldLogger
	context func() context.Context
	stopCh  chan struct{}
}

func newServer(context func() context.Context, logger logrus.FieldLogger) *server {
	srv := &server{
		context: context,
		logger:  logger,
		stopCh:  make(chan struct{}),
	}

	return srv
}

func (s *server) serve(listener net.Listener, handler http.Handler) {
	srv := new(http.Server)
	srv.Handler = handler

	errCh := make(chan error)

	go func() {
		s.logger.Debug("server started")
		errCh <- srv.Serve(listener)
	}()

	var err error

	ctx := s.context()

	select {
	case <-s.stopCh:
		break
	case <-ctx.Done():
		break
	case err = <-errCh:
		break
	}

	if err != nil {
		s.logger.WithError(err).Error("server aborted")

		return
	}

	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		s.logger.WithError(err).Errorf("server shutdown failed")
	}

	s.logger.Debug("server stopped")
}

func (s *server) listenAndServe(addr string, handler http.Handler) (*net.TCPAddr, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s.stopCh = make(chan struct{})

	go s.serve(listener, handler)

	a, _ := listener.Addr().(*net.TCPAddr)

	return a, nil
}

func (s *server) shutdown() {
	s.stopCh <- struct{}{}
}

const shutdownTimeout = 500 * time.Millisecond
