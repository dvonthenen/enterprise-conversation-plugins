// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	interfacessdk "github.com/dvonthenen/enterprise-reference-implementation/pkg/middleware-plugin-sdk/interfaces"
	utils "github.com/dvonthenen/enterprise-reference-implementation/pkg/utils"
	sdkinterfaces "github.com/dvonthenen/symbl-go-sdk/pkg/api/streaming/v1/interfaces"
	neo4j "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	klog "k8s.io/klog/v2"

	interfaces "github.com/dvonthenen/enterprise-conversation-plugins/plugins/statistical/interfaces"
)

func NewHandler(options HandlerOptions) *Handler {
	handler := Handler{
		session:     options.Session,
		symblClient: options.SymblClient,
		cache:       utils.NewMessageCache(),
	}
	return &handler
}

func (h *Handler) SetClientPublisher(mp *interfacessdk.MessagePublisher) {
	klog.V(4).Infof("SetClientPublisher called...\n")
	h.msgPublisher = mp
}

func (h *Handler) InitializedConversation(im *sdkinterfaces.InitializationMessage) error {
	h.conversationID = im.Message.Data.ConversationID
	klog.V(2).Infof("conversationID: %s\n", h.conversationID)
	return nil
}

func (h *Handler) RecognitionResultMessage(rr *sdkinterfaces.RecognitionResult) error {
	// No implementation required. Return Succeess!
	return nil
}

func (h *Handler) MessageResponseMessage(mr *sdkinterfaces.MessageResponse) error {
	for _, msg := range mr.Messages {
		h.cache.Push(msg.ID, msg.Payload.Content, msg.From.ID, msg.From.Name, msg.From.UserID)
	}
	return nil
}

func (h *Handler) InsightResponseMessage(ir *sdkinterfaces.InsightResponse) error {
	// No implementation required. Return Succeess!
	return nil
}

func (h *Handler) TopicResponseMessage(tr *sdkinterfaces.TopicResponse) error {
	for _, curTopic := range tr.Topics {

		last30Mins := h.topicStats(&curTopic, StatType30Mins)
		lastHour := h.topicStats(&curTopic, StatType1Hour)
		last4Hours := h.topicStats(&curTopic, StatType4Hours)
		lastDay := h.topicStats(&curTopic, StatType1Day)
		last2Days := h.topicStats(&curTopic, StatType2Days)
		lastWeek := h.topicStats(&curTopic, StatType1Week)
		lastMonth := h.topicStats(&curTopic, StatType1Month)

		klog.V(2).Infof("Topic Stats: %s\n", curTopic.Phrases)
		klog.V(2).Infof("----------------------------------------\n")
		klog.V(2).Infof("last30Mins: %d\n", last30Mins)
		klog.V(2).Infof("lastHour: %d\n", lastHour)
		klog.V(2).Infof("last4Hours: %d\n", last4Hours)
		klog.V(2).Infof("lastDay: %d\n", lastDay)
		klog.V(2).Infof("last2Days: %d\n", last2Days)
		klog.V(2).Infof("lastWeek: %d\n", lastWeek)
		klog.V(2).Infof("lastMonth: %d\n", lastMonth)

		msg := &interfaces.AppSpecificStatistical{
			Type: sdkinterfaces.MessageTypeUserDefined,
			Metadata: interfaces.Metadata{
				Type: interfaces.AppSpecificMessageTypeStatistical,
			},
			Statistical: interfaces.Data{
				Type:     interfaces.UserStatisticalTypeTopic,
				Insights: make([]interfaces.Insight, 0),
				Stats: interfaces.Stats{
					Last30Mins: last30Mins,
					LastHour:   lastHour,
					Last4Hours: last4Hours,
					LastDay:    lastDay,
					Last2Days:  last2Days,
					LastWeek:   lastWeek,
					LastMonth:  lastMonth,
				},
			},
		}

		msg.Statistical.Insights = append(msg.Statistical.Insights, interfaces.Insight{
			Correlation: strings.ToLower(curTopic.Phrases),
			Messages:    h.convertMessageReferenceToSlice(curTopic.MessageReferences),
		})

		// send the stat
		data, err := json.Marshal(*msg)
		if err != nil {
			klog.V(1).Infof("[Topic] json.Marshal failed. Err: %v\n", err)
		}

		err = (*h.msgPublisher).PublishMessage(h.conversationID, data)
		if err != nil {
			klog.V(1).Infof("[Topic] PublishMessage failed. Err: %v\n", err)
		}
	}

	return nil
}

func (h *Handler) TrackerResponseMessage(tr *sdkinterfaces.TrackerResponse) error {
	for _, curTracker := range tr.Trackers {
		last30Mins := h.trackerStats(&curTracker, StatType30Mins)
		lastHour := h.trackerStats(&curTracker, StatType1Hour)
		last4Hours := h.trackerStats(&curTracker, StatType4Hours)
		lastDay := h.trackerStats(&curTracker, StatType1Day)
		last2Days := h.trackerStats(&curTracker, StatType2Days)
		lastWeek := h.trackerStats(&curTracker, StatType1Week)
		lastMonth := h.trackerStats(&curTracker, StatType1Month)

		klog.V(2).Infof("Topic Stats: %s\n", curTracker.Name)
		klog.V(2).Infof("----------------------------------------\n")
		klog.V(2).Infof("last30Mins: %d\n", last30Mins)
		klog.V(2).Infof("lastHour: %d\n", lastHour)
		klog.V(2).Infof("last4Hours: %d\n", last4Hours)
		klog.V(2).Infof("lastDay: %d\n", lastDay)
		klog.V(2).Infof("last2Days: %d\n", last2Days)
		klog.V(2).Infof("lastWeek: %d\n", lastWeek)
		klog.V(2).Infof("lastMonth: %d\n", lastMonth)

		msg := &interfaces.AppSpecificStatistical{
			Type: sdkinterfaces.MessageTypeUserDefined,
			Metadata: interfaces.Metadata{
				Type: interfaces.AppSpecificMessageTypeStatistical,
			},
			Statistical: interfaces.Data{
				Type:     interfaces.UserStatisticalTypeTracker,
				Insights: make([]interfaces.Insight, 0),
				Stats: interfaces.Stats{
					Last30Mins: last30Mins,
					LastHour:   lastHour,
					Last4Hours: last4Hours,
					LastDay:    lastDay,
					Last2Days:  last2Days,
					LastWeek:   lastWeek,
					LastMonth:  lastMonth,
				},
			},
		}

		for _, match := range curTracker.Matches {
			msg.Statistical.Insights = append(msg.Statistical.Insights, interfaces.Insight{
				Correlation: strings.ToLower(match.Value),
				Messages:    h.convertMessageAndInsightRefsToSlice(match.MessageRefs, match.InsightRefs),
			})
		}

		// send the stat
		data, err := json.Marshal(*msg)
		if err != nil {
			klog.V(1).Infof("[Topic] json.Marshal failed. Err: %v\n", err)
		}

		err = (*h.msgPublisher).PublishMessage(h.conversationID, data)
		if err != nil {
			klog.V(1).Infof("[Topic] PublishMessage failed. Err: %v\n", err)
		}
	}

	return nil
}

func (h *Handler) EntityResponseMessage(er *sdkinterfaces.EntityResponse) error {
	for _, curEntity := range er.Entities {
		for _, curMatch := range curEntity.Matches {
			last30Mins := h.entityStats(&curEntity, &curMatch, StatType30Mins)
			lastHour := h.entityStats(&curEntity, &curMatch, StatType1Hour)
			last4Hours := h.entityStats(&curEntity, &curMatch, StatType4Hours)
			lastDay := h.entityStats(&curEntity, &curMatch, StatType1Day)
			last2Days := h.entityStats(&curEntity, &curMatch, StatType2Days)
			lastWeek := h.entityStats(&curEntity, &curMatch, StatType1Week)
			lastMonth := h.entityStats(&curEntity, &curMatch, StatType1Month)

			klog.V(2).Infof("Topic Stats: %s\n", curMatch.DetectedValue)
			klog.V(2).Infof("----------------------------------------\n")
			klog.V(2).Infof("last30Mins: %d\n", last30Mins)
			klog.V(2).Infof("lastHour: %d\n", lastHour)
			klog.V(2).Infof("last4Hours: %d\n", last4Hours)
			klog.V(2).Infof("lastDay: %d\n", lastDay)
			klog.V(2).Infof("last2Days: %d\n", last2Days)
			klog.V(2).Infof("lastWeek: %d\n", lastWeek)
			klog.V(2).Infof("lastMonth: %d\n", lastMonth)

			msg := &interfaces.AppSpecificStatistical{
				Type: sdkinterfaces.MessageTypeUserDefined,
				Metadata: interfaces.Metadata{
					Type: interfaces.AppSpecificMessageTypeStatistical,
				},
				Statistical: interfaces.Data{
					Type:     interfaces.UserStatisticalTypeEntity,
					Insights: make([]interfaces.Insight, 0),
					Stats: interfaces.Stats{
						Last30Mins: last30Mins,
						LastHour:   lastHour,
						Last4Hours: last4Hours,
						LastDay:    lastDay,
						Last2Days:  last2Days,
						LastWeek:   lastWeek,
						LastMonth:  lastMonth,
					},
				},
			}

			msg.Statistical.Insights = append(msg.Statistical.Insights, interfaces.Insight{
				Correlation: fmt.Sprintf("%s/%s/%s/%s", strings.ToLower(curEntity.Category), strings.ToLower(curEntity.Type), strings.ToLower(curEntity.SubType), strings.ToLower(curMatch.DetectedValue)),
				Messages:    h.convertMessageRefsToSlice(curMatch.MessageRefs),
			})

			// send the stat
			data, err := json.Marshal(*msg)
			if err != nil {
				klog.V(1).Infof("[Entities] json.Marshal failed. Err: %v\n", err)
			}

			err = (*h.msgPublisher).PublishMessage(h.conversationID, data)
			if err != nil {
				klog.V(1).Infof("[Entities] PublishMessage failed. Err: %v\n", err)
			}
		}
	}

	return nil
}

func (h *Handler) TeardownConversation(tm *sdkinterfaces.TeardownMessage) error {
	// No implementation required. Return Succeess!
	return nil
}

func (h *Handler) UserDefinedMessage(data []byte) error {
	// No implementation required. Return Succeess!
	return nil
}

func (h *Handler) UnhandledMessage(byMsg []byte) error {
	klog.Errorf("\n\n-------------------------------\n")
	klog.Errorf("UnhandledMessage:\n%v\n", string(byMsg))
	klog.Errorf("-------------------------------\n\n")
	return ErrUnhandledMessage
}

func (h *Handler) convertMessageAndInsightRefsToSlice(msgRefs []sdkinterfaces.MessageRef, inRefs []sdkinterfaces.InsightRef) []interfaces.Message {
	tmp := make([]interfaces.Message, 0)

	for _, inRef := range inRefs {
		tmp = append(tmp, interfaces.Message{
			Text: inRef.Text,
		})
	}
	for _, msgRef := range msgRefs {
		tmp = append(tmp, interfaces.Message{
			Text: msgRef.Text,
		})
	}

	return tmp
}

func (h *Handler) convertMessageRefsToSlice(msgRefs []sdkinterfaces.MessageRef) []interfaces.Message {
	tmp := make([]interfaces.Message, 0)

	for _, msgRef := range msgRefs {
		tmp = append(tmp, interfaces.Message{
			Text: msgRef.Text,
		})
	}

	return tmp
}

func (h *Handler) convertMessageReferenceToSlice(msgRefs []sdkinterfaces.MessageReference) []interfaces.Message {
	tmp := make([]interfaces.Message, 0)

	for _, msgRef := range msgRefs {
		cacheMessage, err := h.cache.Find(msgRef.ID)
		if err != nil {
			klog.V(4).Infof("Msg ID not found: %s\n", msgRef.ID)
			tmp = append(tmp, interfaces.Message{
				Text: interfaces.MessageNotFound,
			})
			continue
		}

		tmp = append(tmp, interfaces.Message{
			Text: cacheMessage.Text,
		})
	}

	return tmp
}

func (h *Handler) statTypeToString(length StatisticType) string {
	switch length {
	case StatType1Month:
		return "duration({months: 1})"
	case StatType1Week:
		return "duration({days: 7})"
	case StatType2Days:
		return "duration({days: 2})"
	case StatType1Day:
		return "duration({days: 1})"
	case StatType4Hours:
		return "duration({hours: 4})"
	case StatType1Hour:
		return "duration({hours: 1})"
	case StatType30Mins:
		return "duration({minutes: 30})"
	default:
		return "duration({seconds: 1})" // dont return anything
	}
}

func (h *Handler) topicQuery(length StatisticType) string {
	const queryPre = `
	MATCH (t:Topic)-[x:TOPIC_MESSAGE_REF]-(m:Message)
	WHERE x.#conversation_index# <> $conversation_id AND x.value = $topic_phrases AND x.created > datetime() - `
	const queryPost = `
	RETURN count(x)`

	duration := h.statTypeToString(length)

	return fmt.Sprintf("%s%s%s", queryPre, duration, queryPost)
}

func (h *Handler) trackerQuery(length StatisticType) string {
	const queryPre = `
	MATCH (t:Tracker)-[x:TRACKER_MESSAGE_REF]-(m:Message)
	WHERE x.#conversation_index# <> $conversation_id AND x.name = $tracker_name AND x.createdAt > datetime() - `
	const queryPost = `
	RETURN count(x)`

	duration := h.statTypeToString(length)

	return fmt.Sprintf("%s%s%s", queryPre, duration, queryPost)
}

func (h *Handler) entityQuery(length StatisticType) string {
	const queryPre = `
	MATCH (e:Entity)-[x:ENTITY_MESSAGE_REF]-(m:Message)
	WHERE x.#conversation_index# <> $conversation_id AND e.category = $entity_category AND e.type = $entity_type AND e.subType = $entity_subtype AND x.value = $entity_value AND x.createdAt > datetime() - `
	const queryPost = `
	RETURN count(x)`

	duration := h.statTypeToString(length)

	return fmt.Sprintf("%s%s%s", queryPre, duration, queryPost)
}

func (h *Handler) topicStats(curTopic *sdkinterfaces.Topic, length StatisticType) int64 {
	ctx := context.Background()

	var retValue int64

	_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		myQuery := utils.ReplaceIndexes(h.topicQuery(length))
		result, err := tx.Run(ctx, myQuery, map[string]any{
			"conversation_id": h.conversationID,
			"topic_phrases":   strings.ToLower(curTopic.Phrases),
		})
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			retValue = result.Record().Values[0].(int64)
		}

		return nil, result.Err()
	})
	if err != nil {
		klog.V(1).Infof("[Entities] ExecuteRead failed. Err: %v\n", err)
		return 0
	}

	return retValue
}

func (h *Handler) trackerStats(curTracker *sdkinterfaces.Tracker, length StatisticType) int64 {
	ctx := context.Background()

	var retValue int64

	_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := h.trackerQuery(length)
		klog.V(2).Infof("QUERY: \n%s\n", query)
		myQuery := utils.ReplaceIndexes(query)
		result, err := tx.Run(ctx, myQuery, map[string]any{
			"conversation_id": h.conversationID,
			"tracker_name":    strings.ToLower(curTracker.Name),
		})
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			retValue = result.Record().Values[0].(int64)
		}

		return nil, result.Err()
	})
	if err != nil {
		klog.V(1).Infof("[Tracker] ExecuteRead failed. Err: %v\n", err)
		return 0
	}

	return retValue
}

func (h *Handler) entityStats(entity *sdkinterfaces.Entity, match *sdkinterfaces.EntityMatch, length StatisticType) int64 {
	ctx := context.Background()

	var retValue int64

	_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := h.entityQuery(length)
		klog.V(2).Infof("QUERY: \n%s\n", query)
		myQuery := utils.ReplaceIndexes(query)
		result, err := tx.Run(ctx, myQuery, map[string]any{
			"conversation_id": h.conversationID,
			"entity_category": strings.ToLower(entity.Category),
			"entity_type":     strings.ToLower(entity.Type),
			"entity_subtype":  strings.ToLower(entity.SubType),
			"entity_value":    strings.ToLower(match.DetectedValue),
		})
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			retValue = result.Record().Values[0].(int64)
		}

		return nil, result.Err()
	})
	if err != nil {
		klog.V(1).Infof("entityStats ExecuteRead failed. Err: %v\n", err)
		return 0
	}

	return retValue
}
