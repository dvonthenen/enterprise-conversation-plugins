// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package server

import (
	"os"

	middlewaresdk "github.com/dvonthenen/enterprise-reference-implementation/pkg/middleware-plugin-sdk"
	interfacessdk "github.com/dvonthenen/enterprise-reference-implementation/pkg/middleware-plugin-sdk/interfaces"
	klog "k8s.io/klog/v2"

	handlers "github.com/dvonthenen/enterprise-conversation-plugins/plugins/asynchronous/email/handlers"
)

func New(options ServerOptions) (*Server, error) {
	if options.BindPort == 0 {
		options.BindPort = DefaultPort
	}

	var configFile string
	if v := os.Getenv("EMAIL_CONFIG_FILE"); v != "" {
		klog.V(4).Info("EMAIL_CONFIG_FILE found")
		configFile = v
	} else {
		klog.Errorf("EMAIL_CONFIG_FILE not found\n")
		return nil, ErrInvalidInput
	}
	options.ConfigFile = configFile

	// server
	server := &Server{
		options: options,
	}
	return server, nil
}

func (s *Server) Init() error {
	klog.V(6).Infof("Server.Init ENTER\n")

	// middleware analyzer
	err := s.RebuildAsynchronousAnalyzer()
	if err != nil {
		klog.V(1).Infof("RebuildAsynchronousAnalyzer failed. Err: %v\n", err)
		klog.V(6).Infof("Server.Init LEAVE\n")
		return err
	}

	klog.V(4).Infof("Server.Init Succeeded\n")
	klog.V(6).Infof("Server.Init LEAVE\n")

	return nil
}

func (s *Server) Start() error {
	klog.V(6).Infof("Server.Start ENTER\n")

	// rebuild middleware if needed
	if s.middlewareAnalyzer == nil {
		klog.V(4).Infof("Calling RebuildAsynchronousAnalyzer...\n")
		err := s.RebuildAsynchronousAnalyzer()
		if err != nil {
			klog.V(1).Infof("RebuildAsynchronousAnalyzer failed. Err: %v\n", err)
			klog.V(6).Infof("Server.Start LEAVE\n")
			return err
		}
	}

	// start middleware
	err := s.middlewareAnalyzer.Init()
	if err != nil {
		klog.V(1).Infof("middlewareAnalyzer.Init() failed. Err: %v\n", err)
		klog.V(6).Infof("Server.Start LEAVE\n")
		return err
	}

	// TODO: start metrics and tracing

	klog.V(4).Infof("Server.Start Succeeded\n")
	klog.V(6).Infof("Server.Start LEAVE\n")

	return nil
}

func (s *Server) RebuildAsynchronousAnalyzer() error {
	klog.V(6).Infof("Server.RebuildAsynchronousAnalyzer ENTER\n")

	// teardown
	if s.middlewareAnalyzer != nil {
		err := (*s.middlewareAnalyzer).Teardown()
		if err != nil {
			klog.V(1).Infof("middlewareAnalyzer.Teardown failed. Err: %v\n", err)
		}
		s.middlewareAnalyzer = nil
	}

	// create handler
	messageHandler := handlers.NewHandler(handlers.HandlerOptions{
		ConfigFile: s.options.ConfigFile,
	})

	err := messageHandler.ParseConfig()
	if err != nil {
		klog.V(1).Infof("ParseConfig failed. Err: %v\n", err)
		klog.V(6).Infof("Server.RebuildAsynchronousAnalyzer LEAVE\n")
		return err
	}

	// create middleware
	var callback interfacessdk.AsynchronousCallback
	callback = messageHandler

	middlewareAnalyzer, err := middlewaresdk.NewAsynchronousAnalyzer(middlewaresdk.AsynchronousAnalyzerOption{
		RabbitURI: s.options.RabbitURI,
		Callback:  &callback,
	})
	if err != nil {
		klog.V(1).Infof("NewAsynchronousAnalyzer failed. Err: %v\n", err)
		klog.V(6).Infof("Server.RebuildAsynchronousAnalyzer LEAVE\n")
		return err
	}

	// housekeeping
	s.middlewareAnalyzer = middlewareAnalyzer

	klog.V(4).Infof("Server.RebuildAsynchronousAnalyzer Succeeded\n")
	klog.V(6).Infof("Server.RebuildAsynchronousAnalyzer LEAVE\n")

	return nil
}

func (s *Server) Stop() error {
	klog.V(6).Infof("Server.Stop ENTER\n")

	// TODO: stop metrics and tracing

	// clean up middleware
	if s.middlewareAnalyzer != nil {
		err := s.middlewareAnalyzer.Teardown()
		if err != nil {
			klog.V(1).Infof("middlewareAnalyzer.Teardown failed. Err: %v\n", err)
		}
	}
	s.middlewareAnalyzer = nil

	klog.V(4).Infof("Server.Stop Succeeded\n")
	klog.V(6).Infof("Server.Stop LEAVE\n")

	return nil
}
