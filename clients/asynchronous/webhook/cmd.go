// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	prettyjson "github.com/hokaccha/go-prettyjson"
	klog "k8s.io/klog/v2"
)

const (
	DefaultPort int = 17000

	PropertyWebhookPassword string = "Symbl-Webhook-Plugin-Secret"
)

type LogLevel int64

const (
	LogLevelDefault   LogLevel = iota
	LogLevelErrorOnly          = 1
	LogLevelStandard           = 2
	LogLevelElevated           = 3
	LogLevelFull               = 4
	LogLevelDebug              = 5
	LogLevelTrace              = 6
	LogLevelVerbose            = 7
)

// var (
// 	// ErrInvalidInput required input was not found
// 	ErrInvalidInput = errors.New("required input was not found")
// )

// SampleServerOptions for the main HTTP endpoint
type SampleServerOptions struct {
	CrtFile         string
	KeyFile         string
	BindPort        int
	WebhookPassword string
}

type SampleServer struct {
	// housekeeping
	options *SampleServerOptions

	// server
	server *http.Server
}

func New(options SampleServerOptions) (*SampleServer, error) {
	if options.BindPort == 0 {
		options.BindPort = DefaultPort
	}

	server := &SampleServer{
		options: &options,
	}
	return server, nil
}

func (ss *SampleServer) postWebhook(c *gin.Context) {
	klog.V(6).Infof("postWebhook ENTER\n")

	// for key, value := range c.Request.Header {
	// 	klog.V(5).Infof("HTTP Header: %s = %v\n", key, value)
	// }

	webhookPassword := c.Request.Header[PropertyWebhookPassword][0]
	klog.V(7).Infof("webhookPassword: %s\n", webhookPassword)

	if !strings.EqualFold(ss.options.WebhookPassword, webhookPassword) {
		errStr := "Webhook Password does not match"
		klog.V(1).Infof("%s\n", errStr)
		c.String(http.StatusBadRequest, errStr)
		return
	}

	conversationId := c.Param("conversation_id")
	klog.V(4).Infof("conversationId: %s\n", conversationId)

	byData, err := c.GetRawData()
	if err != nil {
		errStr := fmt.Sprintf("c.GetRawData failed. Err: %v", err)
		klog.V(1).Infof("%s\n", errStr)
		c.String(http.StatusBadRequest, errStr)
		return
	}

	prettyJson, err := prettyjson.Format(byData)
	if err != nil {
		fmt.Printf("prettyjson.Marshal failed. Err: %v\n", err)
		os.Exit(1)
	}
	klog.V(2).Infof("\n\nBody Raw:\n\n%s\n\n", prettyJson)

	klog.V(4).Infof("postWebhook Succeeded\n")
	klog.V(6).Infof("postWebhook LEAVE\n")

	// c.Writer.WriteHeader(http.StatusOK)
	// c.IndentedJSON(http.StatusOK, resp)
	c.String(http.StatusOK, "Conversation insights received")
}

func (ss *SampleServer) Start() error {
	klog.V(6).Infof("SampleServer.Start ENTER\n")

	// redirect
	router := gin.Default()
	router.POST("/v1/webhook/:conversation_id", ss.postWebhook)

	// server
	ss.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ss.options.BindPort),
		Handler: router,
	}

	// start the main entry endpoint to direct traffic
	go func() {
		// this is a blocking call
		err := ss.server.ListenAndServeTLS(ss.options.CrtFile, ss.options.KeyFile)
		if err != nil {
			klog.V(6).Infof("ListenAndServeTLS server stopped. Err: %v\n", err)
		}
	}()

	// TODO: start metrics and tracing

	klog.V(4).Infof("SampleServer.Start Succeeded\n")
	klog.V(6).Infof("SampleServer.Start LEAVE\n")

	return nil
}

func (ss *SampleServer) Stop() error {
	klog.V(6).Infof("SampleServer.Stop ENTER\n")

	// TODO: stop metrics and tracing

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ss.server.Shutdown(ctx); err != nil {
		klog.V(1).Infof("Server Shutdown Failed. Err: %v\n", err)
	}

	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		klog.V(1).Infof("timeout of 5 seconds.")
	}

	klog.V(4).Infof("SampleServer.Stop Succeeded\n")
	klog.V(6).Infof("SampleServer.Stop LEAVE\n")

	return nil
}

func main() {
	klog.InitFlags(nil)
	flag.Set("v", strconv.FormatInt(int64(LogLevelTrace), 10))
	flag.Parse()

	server, err := New(SampleServerOptions{
		CrtFile:         "localhost.crt",
		KeyFile:         "localhost.key",
		WebhookPassword: "MyPassword", // TODO: this should come from env
	})
	if err != nil {
		klog.V(1).Infof("New failed. Err: %v\n", err)
		os.Exit(1)
	}

	// start
	fmt.Printf("Starting server...\n")
	err = server.Start()
	if err != nil {
		fmt.Printf("server.Start() failed. Err: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Press ENTER to exit!\n\n")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	// stop
	err = server.Stop()
	if err != nil {
		fmt.Printf("server.Stop() failed. Err: %v\n", err)
	}

	fmt.Printf("Server stopped...\n\n")
}
