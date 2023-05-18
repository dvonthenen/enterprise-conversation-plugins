// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"text/template"

	gomail "gopkg.in/mail.v2"
	klog "k8s.io/klog/v2"

	interfacessdk "github.com/dvonthenen/enterprise-conversation-application/pkg/middleware-plugin-sdk/interfaces"
	shared "github.com/dvonthenen/enterprise-conversation-application/pkg/shared"
	utils "github.com/dvonthenen/enterprise-conversation-application/pkg/utils"
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
	var stmpPassword string
	if v := os.Getenv("EMAIL_SMTP_PASSWORD"); v != "" {
		klog.V(4).Info("EMAIL_SMTP_PASSWORD found")
		stmpPassword = v
	} else {
		klog.Errorf("EMAIL_SMTP_PASSWORD not found\n")
		klog.V(6).Infof("ParseConfig LEAVE\n")
		return ErrInvalidInput
	}
	h.config.EmailSmtpPassword = stmpPassword

	// template
	h.template, err = template.ParseFiles(h.config.Template)
	if err != nil {
		klog.V(1).Infof("template.ParseFiles failed. Err: %v\n", err)
		klog.V(6).Infof("ParseConfig LEAVE\n")
		return err
	}

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
	klog.V(3).Infof("TeardownConversation - conversationID: %s\n", conversationId)

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
	klog.V(5).Infof("triggers matched:\n")
	for _, trigger := range triggers {
		klog.V(5).Infof("%s\n", trigger)
	}

	if len(triggers) == 0 {
		klog.V(3).Infof("No triggers in conversationId: %s\n", conversationId)
		return nil
	}

	// email body
	data, err := json.Marshal(conversation)
	if err != nil {
		klog.V(1).Infof("InitializedConversation json.Marshal failed. Err: %v\n", err)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return err
	}

	// convert string port to int
	ismtpPort, err := strconv.Atoi(h.config.EmailSmtpPort)
	if err != nil {
		klog.V(1).Infof("strconv.Atoi failed. Err: %v\n", err)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return err
	}
	// body
	var body bytes.Buffer
	h.template.Execute(&body, struct {
		Triggers []string
		Dump     string
	}{
		Triggers: triggers,
		Dump:     string(data),
	})

	// setup and send email
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", h.config.EmailFrom)

	// Set E-Mail receivers
	m.SetHeader("To", h.config.EmailTo)

	// Set E-Mail subject
	m.SetHeader("Subject", h.config.EmailSubject)

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", body.String())

	// Settings for SMTP server
	d := gomail.NewDialer(h.config.EmailSmtpAddr, ismtpPort, h.config.EmailSmtpUsername, h.config.EmailSmtpPassword)

	// skip server auth
	if h.config.SkipServerAuth {
		// TODO: add verification later, pick up from ENV or FILE
		/* #nosec G402 */
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		klog.V(1).Infof("DialAndSend failed. Err: %v\n", err)
		klog.V(6).Infof("TeardownConversation LEAVE\n")
		return err
	}

	// clean up
	delete(h.cache, conversationId)
	delete(h.conversations, conversationId)
	delete(h.triggers, conversationId)

	klog.V(4).Infof("TeardownConversation Succeeded\n")
	klog.V(6).Infof("TeardownConversation LEAVE\n")
	return nil
}
