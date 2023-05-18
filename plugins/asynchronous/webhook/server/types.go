// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package server

import (
	middlewaresdk "github.com/dvonthenen/enterprise-reference-implementation/pkg/middleware-plugin-sdk"
)

// ServerOptions for the main HTTP endpoint
type ServerOptions struct {
	CrtFile     string
	KeyFile     string
	BindAddress string
	BindPort    int
	RabbitURI   string
	ConfigFile  string
}

type Server struct {
	// server options
	options ServerOptions

	// middleware
	middlewareAnalyzer *middlewaresdk.AsynchronousAnalyzer
}
