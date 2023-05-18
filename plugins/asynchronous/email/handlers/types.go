// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import (
	"text/template"

	interfacessdk "github.com/dvonthenen/enterprise-conversation-application/pkg/middleware-plugin-sdk/interfaces"
	utils "github.com/dvonthenen/enterprise-conversation-application/pkg/utils"
	sdkinterfaces "github.com/dvonthenen/symbl-go-sdk/pkg/api/async/v1/interfaces"
)

/*
	Single Transaction
*/
type ConversationResult struct {
	ConversationID   string                          `json:"conversationId,omitempty"`
	MessageResult    *sdkinterfaces.MessageResult    `json:"messageResult,omitempty"`
	QuestionResult   *sdkinterfaces.QuestionResult   `json:"questionResult,omitempty"`
	FollowUpResult   *sdkinterfaces.FollowUpResult   `json:"followUpResult,omitempty"`
	ActionItemResult *sdkinterfaces.ActionItemResult `json:"actionItemResult,omitempty"`
	TopicResult      *sdkinterfaces.TopicResult      `json:"topicResult,omitempty"`
	TrackerResult    *sdkinterfaces.TrackerResult    `json:"trackerResult,omitempty"`
	EntityResult     *sdkinterfaces.EntityResult     `json:"entityResult,omitempty"`
}

/*
	Config
*/
type Config struct {
	Template          string   `json:"template,omitempty"`
	SkipServerAuth    bool     `json:"skipServerAuth,omitempty"`
	EmailTo           string   `json:"emailTo,omitempty"`
	EmailFrom         string   `json:"emailFrom,omitempty"`
	EmailSubject      string   `json:"emailSubject,omitempty"`
	EmailSmtpAddr     string   `json:"emailSmtpAddr,omitempty"`
	EmailSmtpPort     string   `json:"emailPort,omitempty"`
	EmailSmtpUsername string   `json:"emailSmtpUsername,omitempty"`
	EmailSmtpPassword string   `json:"emailSmtpPassword,omitempty"`
	QuestionMatch     []string `json:"questionMatch,omitempty"`
	FollowUpMatch     []string `json:"followUpMatch,omitempty"`
	ActionItemMatch   []string `json:"actionItemMatch,omitempty"`
	TopicMatch        []string `json:"topicMatch,omitempty"`
	TrackerMatch      []string `json:"trackerMatch,omitempty"`
	EntityMatch       []string `json:"entityMatch,omitempty"`
}

/*
	Handler for messages
*/
type HandlerOptions struct {
	ConfigFile string
}

type Handler struct {
	// handler options
	options HandlerOptions
	config  Config

	// properties
	cache         map[string]*utils.MessageCache
	conversations map[string]*ConversationResult
	triggers      map[string][]string
	template      *template.Template

	// housekeeping
	msgPublisher *interfacessdk.MessagePublisher
}
