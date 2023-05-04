// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	klog "k8s.io/klog/v2"

	interfacessdk "github.com/dvonthenen/enterprise-reference-implementation/pkg/middleware-plugin-sdk/interfaces"
	shared "github.com/dvonthenen/enterprise-reference-implementation/pkg/shared"
	utils "github.com/dvonthenen/enterprise-reference-implementation/pkg/utils"
)

func NewHandler(options HandlerOptions) *Handler {
	handler := Handler{
		options:       options,
		cache:         make(map[string]*utils.MessageCache),
		conversations: make(map[string]*ConversationResult),
		triggers:      make(map[string][]string),
	}
	return &handler
}

func (h *Handler) SetClientPublisher(mp *interfacessdk.MessagePublisher) {
	klog.V(4).Infof("SetClientPublisher called...\n")
	h.msgPublisher = mp
}

func (h *Handler) ParseConfig() error {
	klog.V(6).Infof("ParseConfig ENTER\n")

	byData, err := os.ReadFile(h.options.ConfigFile)
	if err != nil {
		klog.V(1).Infof("os.ReadFile failed. Err: %v\n", err)
		klog.V(6).Infof("ParseConfig LEAVE\n")
		return err
	}
	klog.V(5).Infof("\n\nbyData:\n%s\n\n", string(byData))

	err = json.Unmarshal(byData, &h.config)
	if err != nil {
		klog.V(1).Infof("json.Unmarshal failed. Err: %v\n", err)
		klog.V(6).Infof("ParseConfig LEAVE\n")
		return err
	}

	// password for env
	var webhookPassword string
	if v := os.Getenv("WEBHOOK_PASSWORD"); v != "" {
		klog.V(4).Info("WEBHOOK_PASSWORD found")
		webhookPassword = v
	} else {
		klog.Errorf("WEBHOOK_PASSWORD not found\n")
		klog.V(6).Infof("ParseConfig LEAVE\n")
		return ErrInvalidInput
	}
	h.config.WebhookPassword = webhookPassword

	klog.V(4).Infof("ParseConfig Succeeded\n")
	klog.V(6).Infof("ParseConfig LEAVE\n")
	return nil
}

func (h *Handler) InitializedConversation(im *shared.InitializationResult) error {
	conversationId := im.InitializationMessage.ConversationID
	klog.V(2).Infof("InitializedConversation - conversationID: %s\n", conversationId)

	// housekeeping
	h.cache[conversationId] = utils.NewMessageCache()
	h.conversations[conversationId] = &ConversationResult{
		ConversationID: conversationId,
	}
	h.triggers[conversationId] = make([]string, 0)

	return nil
}

func (h *Handler) MessageResult(mr *shared.MessageResult) error {
	result := h.conversations[mr.ConversationID]
	if result == nil {
		return ErrConversationNotFound
	}
	result.MessageResult = mr.MessageResult

	cache := h.cache[mr.ConversationID]
	if cache != nil {
		for _, msg := range mr.MessageResult.Messages {
			cache.Push(msg.ID, msg.Text, msg.From.ID, msg.From.Name, "")
		}
	} else {
		klog.V(1).Infof("MessageCache for ConversationID(%s) not found.", mr.ConversationID)
	}

	return nil
}

func (h *Handler) QuestionResult(qr *shared.QuestionResult) error {
	result := h.conversations[qr.ConversationID]
	if result == nil {
		return ErrConversationNotFound
	}
	result.QuestionResult = qr.QuestionResult

	for _, question := range qr.QuestionResult.Questions {
		for _, regex := range h.config.FollowUpMatch {
			match, err := regexp.MatchString(regex, question.Text)
			if err != nil {
				klog.V(6).Infof("MatchString failed. Err: %v\n", err)
				continue
			}
			if !match {
				klog.V(6).Infof("%s != %s\n", regex, question.Text)
				continue
			}
			klog.V(2).Infof("Match %s = %s\n", regex, question.Text)
			h.triggers[qr.ConversationID] = append(h.triggers[qr.ConversationID], fmt.Sprintf("Question - %s", question.Text))
		}
	}

	return nil
}

func (h *Handler) FollowUpResult(fur *shared.FollowUpResult) error {
	result := h.conversations[fur.ConversationID]
	if result == nil {
		return ErrConversationNotFound
	}
	result.FollowUpResult = fur.FollowUpResult

	for _, followUp := range fur.FollowUpResult.FollowUps {
		for _, regex := range h.config.FollowUpMatch {
			match, err := regexp.MatchString(regex, followUp.Text)
			if err != nil {
				klog.V(6).Infof("MatchString failed. Err: %v\n", err)
				continue
			}
			if !match {
				klog.V(6).Infof("%s != %s\n", regex, followUp.Text)
				continue
			}
			klog.V(2).Infof("Match %s = %s\n", regex, followUp.Text)
			h.triggers[fur.ConversationID] = append(h.triggers[fur.ConversationID], fmt.Sprintf("FollowUp - %s", followUp.Text))
		}
	}

	return nil
}

func (h *Handler) ActionItemResult(air *shared.ActionItemResult) error {
	result := h.conversations[air.ConversationID]
	if result == nil {
		return ErrConversationNotFound
	}
	result.ActionItemResult = air.ActionItemResult

	for _, actionItem := range air.ActionItemResult.ActionItems {
		for _, regex := range h.config.FollowUpMatch {
			match, err := regexp.MatchString(regex, actionItem.Text)
			if err != nil {
				klog.V(6).Infof("MatchString failed. Err: %v\n", err)
				continue
			}
			if !match {
				klog.V(6).Infof("%s != %s\n", regex, actionItem.Text)
				continue
			}
			klog.V(2).Infof("Match %s = %s\n", regex, actionItem.Text)
			h.triggers[air.ConversationID] = append(h.triggers[air.ConversationID], fmt.Sprintf("ActionItem - %s", actionItem.Text))
		}
	}

	return nil
}

func (h *Handler) TopicResult(tr *shared.TopicResult) error {
	result := h.conversations[tr.ConversationID]
	if result == nil {
		return ErrConversationNotFound
	}
	result.TopicResult = tr.TopicResult

	for _, topic := range tr.TopicResult.Topics {
		for _, regex := range h.config.FollowUpMatch {
			match, err := regexp.MatchString(regex, topic.Text)
			if err != nil {
				klog.V(6).Infof("MatchString failed. Err: %v\n", err)
				continue
			}
			if !match {
				klog.V(6).Infof("%s != %s\n", regex, topic.Text)
				continue
			}
			klog.V(2).Infof("Match %s = %s\n", regex, topic.Text)
			h.triggers[tr.ConversationID] = append(h.triggers[tr.ConversationID], fmt.Sprintf("Topic - %s", topic.Text))
		}
	}

	return nil
}

func (h *Handler) TrackerResult(tr *shared.TrackerResult) error {
	result := h.conversations[tr.ConversationID]
	if result == nil {
		return ErrConversationNotFound
	}
	result.TrackerResult = tr.TrackerResult

	for _, trackerMatch := range tr.TrackerResult.Matches {
		for _, regex := range h.config.FollowUpMatch {
			match, err := regexp.MatchString(regex, tr.TrackerResult.Name)
			if err != nil {
				klog.V(6).Infof("MatchString failed. Err: %v\n", err)
				continue
			}
			if !match {
				klog.V(6).Infof("%s != %s/%s\n", regex, tr.TrackerResult.Name, trackerMatch.Value)
				continue
			}
			klog.V(2).Infof("Match %s = %s/%s\n", regex, tr.TrackerResult.Name, trackerMatch.Value)
			h.triggers[tr.ConversationID] = append(h.triggers[tr.ConversationID], fmt.Sprintf("Tracker - %s/%s", tr.TrackerResult.Name, trackerMatch.Value))
		}
	}

	return nil
}

func (h *Handler) EntityResult(er *shared.EntityResult) error {
	result := h.conversations[er.ConversationID]
	if result == nil {
		return ErrConversationNotFound
	}
	result.EntityResult = er.EntityResult

	for _, entity := range er.EntityResult.Entities {
		for _, entityMatch := range entity.Matches {
			for _, regex := range h.config.FollowUpMatch {
				match, err := regexp.MatchString(regex, entityMatch.DetectedValue)
				if err != nil {
					klog.V(6).Infof("MatchString failed. Err: %v\n", err)
					continue
				}
				if !match {
					klog.V(6).Infof("%s != %s\n", regex, entityMatch.DetectedValue)
					continue
				}
				klog.V(2).Infof("Match %s = %s\n", regex, entityMatch.DetectedValue)
				h.triggers[er.ConversationID] = append(h.triggers[er.ConversationID], fmt.Sprintf("Entity - %s", entityMatch.DetectedValue))
			}
		}
	}

	return nil
}

func (h *Handler) TeardownConversation(tm *shared.TeardownResult) error {
	klog.V(6).Infof("TeardownConversation ENTER\n")

	conversationId := tm.TeardownMessage.ConversationID
	klog.V(2).Infof("TeardownConversation - conversationID: %s\n", conversationId)

	conversation := h.conversations[conversationId]
	if conversation == nil {
		klog.V(1).Infof("conversations[%s] not found\n", conversationId)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return ErrConversationNotFound
	}
	triggers := h.triggers[conversationId]
	if conversation == nil {
		klog.V(1).Infof("triggers[%s] not found\n", conversationId)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return ErrConversationNotFound
	}

	// conversation of interest?
	klog.V(2).Infof("triggers matched:\n")
	for _, trigger := range triggers {
		klog.V(2).Infof("%s\n", trigger)
	}

	if len(triggers) == 0 {
		klog.V(3).Infof("No triggers in conversationId: %s\n", conversationId)
		return nil
	}

	// call webhook
	byData, err := json.Marshal(conversation)
	if err != nil {
		klog.V(1).Infof("json.Unmarshal failed. Err: %v\n", err)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	uri := fmt.Sprintf("%s/%s", h.config.WebhookURI, conversationId)
	klog.V(2).Infof("Webhook URI: %s\n", uri)
	req, err := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(byData))
	if err != nil {
		klog.V(1).Infof("http.NewRequestWithContext failed. Err: %v\n", err)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return err
	}

	// secret
	req.Header.Add("SYMBL-WEBHOOK-PLUGIN-SECRET", h.config.WebhookPassword)

	switch req.Method {
	case http.MethodPost, http.MethodPatch, http.MethodPut:
		klog.V(3).Infof("Content-Type = application/json\n")
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}

	if h.config.SkipServerAuth {
		// TODO: add verification later, pick up from ENV or FILE
		/* #nosec G402 */
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{
			Transport: tr,
		}
	}

	// execute!
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		klog.V(1).Infof("http.NewRequestWithContext failed. Err: %v\n", err)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return err
	}

	// clean up
	delete(h.cache, conversationId)
	delete(h.conversations, conversationId)
	delete(h.triggers, conversationId)

	// error handling
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusBadRequest:
		klog.V(1).Infof("HTTP Error Code: %d\n", resp.StatusCode)
		detail, err := io.ReadAll(resp.Body)
		if err != nil {
			klog.V(1).Infof("io.ReadAll failed. Err: %e\n", err)
		}
		return fmt.Errorf("%s: %s", resp.Status, bytes.TrimSpace(detail))
	default:
		klog.V(1).Infof("Unknown/Fatal Error")
		return fmt.Errorf("Unknown/Fatal Error")
	}

	klog.V(4).Infof("TeardownConversation Succeeded\n")
	klog.V(6).Infof("TeardownConversation LEAVE\n")
	return nil
}
