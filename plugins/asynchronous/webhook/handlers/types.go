// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import (
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
	WebhookURI      string   `json:"webhookURI,omitempty"`
	WebhookPassword string   `json:"webhookPassword,omitempty"`
	SkipServerAuth  bool     `json:"skipServerAuth,omitempty"`
	QuestionMatch   []string `json:"questionMatch,omitempty"`
	FollowUpMatch   []string `json:"followUpMatch,omitempty"`
	ActionItemMatch []string `json:"actionItemMatch,omitempty"`
	TopicMatch      []string `json:"topicMatch,omitempty"`
	TrackerMatch    []string `json:"trackerMatch,omitempty"`
	EntityMatch     []string `json:"entityMatch,omitempty"`
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

	// housekeeping
	msgPublisher *interfacessdk.MessagePublisher
}
