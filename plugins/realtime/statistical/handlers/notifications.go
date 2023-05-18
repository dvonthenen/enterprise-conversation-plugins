// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdkinterfaces "github.com/dvonthenen/symbl-go-sdk/pkg/api/streaming/v1/interfaces"
	neo4j "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	klog "k8s.io/klog/v2"

	interfacessdk "github.com/dvonthenen/enterprise-reference-implementation/pkg/middleware-plugin-sdk/interfaces"
	shared "github.com/dvonthenen/enterprise-reference-implementation/pkg/shared"
	utils "github.com/dvonthenen/enterprise-reference-implementation/pkg/utils"

	interfaces "github.com/dvonthenen/enterprise-conversation-plugins/plugins/realtime/statistical/interfaces"
)

func NewHandler(options HandlerOptions) *Handler {
	handler := Handler{
		session:     options.Session,
		symblClient: options.SymblClient,
		cache:       make(map[string]*utils.MessageCache),
	}
	return &handler
}

func (h *Handler) SetClientPublisher(mp *interfacessdk.MessagePublisher) {
	klog.V(4).Infof("SetClientPublisher called...\n")
	h.msgPublisher = mp
}

func (h *Handler) InitializedConversation(im *shared.InitializationResponse) error {
	conversationId := im.InitializationMessage.Message.Data.ConversationID
	klog.V(2).Infof("InitializedConversation - conversationID: %s\n", conversationId)
	h.cache[conversationId] = utils.NewMessageCache()
	return nil
}

func (h *Handler) RecognitionResultMessage(rr *shared.RecognitionResponse) error {
	// No implementation required. Return Succeess!
	return nil
}

func (h *Handler) MessageResponseMessage(mr *shared.MessageResponse) error {
	cache := h.cache[mr.ConversationID]
	if cache != nil {
		for _, msg := range mr.MessageResponse.Messages {
			cache.Push(msg.ID, msg.Payload.Content, msg.From.ID, msg.From.Name, msg.From.UserID)
		}
	} else {
		klog.V(1).Infof("MessageCache for ConversationID(%s) not found.", mr.ConversationID)
	}

	return nil
}

func (h *Handler) InsightResponseMessage(ir *shared.InsightResponse) error {
	// No implementation required. Return Succeess!
	return nil
}

func (h *Handler) TopicResponseMessage(tr *shared.TopicResponse) error {
	for _, curTopic := range tr.TopicResponse.Topics {

		last30Mins := h.topicStats(tr.ConversationID, &curTopic, StatType30Mins)
		lastHour := h.topicStats(tr.ConversationID, &curTopic, StatType1Hour)
		last4Hours := h.topicStats(tr.ConversationID, &curTopic, StatType4Hours)
		lastDay := h.topicStats(tr.ConversationID, &curTopic, StatType1Day)
		last2Days := h.topicStats(tr.ConversationID, &curTopic, StatType2Days)
		lastWeek := h.topicStats(tr.ConversationID, &curTopic, StatType1Week)
		lastMonth := h.topicStats(tr.ConversationID, &curTopic, StatType1Month)

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
			Messages:    h.convertMessageReferenceToSlice(tr.ConversationID, curTopic.MessageReferences),
		})

		// send the stat
		data, err := json.Marshal(*msg)
		if err != nil {
			klog.V(1).Infof("[Topic] json.Marshal failed. Err: %v\n", err)
		}

		err = (*h.msgPublisher).PublishMessage(tr.ConversationID, data)
		if err != nil {
			klog.V(1).Infof("[Topic] PublishMessage failed. Err: %v\n", err)
		}
	}

	return nil
}

func (h *Handler) TrackerResponseMessage(tr *shared.TrackerResponse) error {
	for _, curTracker := range tr.TrackerResponse.Trackers {
		last30Mins := h.trackerStats(tr.ConversationID, &curTracker, StatType30Mins)
		lastHour := h.trackerStats(tr.ConversationID, &curTracker, StatType1Hour)
		last4Hours := h.trackerStats(tr.ConversationID, &curTracker, StatType4Hours)
		lastDay := h.trackerStats(tr.ConversationID, &curTracker, StatType1Day)
		last2Days := h.trackerStats(tr.ConversationID, &curTracker, StatType2Days)
		lastWeek := h.trackerStats(tr.ConversationID, &curTracker, StatType1Week)
		lastMonth := h.trackerStats(tr.ConversationID, &curTracker, StatType1Month)

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

		err = (*h.msgPublisher).PublishMessage(tr.ConversationID, data)
		if err != nil {
			klog.V(1).Infof("[Topic] PublishMessage failed. Err: %v\n", err)
		}
	}

	return nil
}

func (h *Handler) EntityResponseMessage(er *shared.EntityResponse) error {
	for _, curEntity := range er.EntityResponse.Entities {
		for _, curMatch := range curEntity.Matches {
			last30Mins := h.entityStats(er.ConversationID, &curEntity, &curMatch, StatType30Mins)
			lastHour := h.entityStats(er.ConversationID, &curEntity, &curMatch, StatType1Hour)
			last4Hours := h.entityStats(er.ConversationID, &curEntity, &curMatch, StatType4Hours)
			lastDay := h.entityStats(er.ConversationID, &curEntity, &curMatch, StatType1Day)
			last2Days := h.entityStats(er.ConversationID, &curEntity, &curMatch, StatType2Days)
			lastWeek := h.entityStats(er.ConversationID, &curEntity, &curMatch, StatType1Week)
			lastMonth := h.entityStats(er.ConversationID, &curEntity, &curMatch, StatType1Month)

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

			err = (*h.msgPublisher).PublishMessage(er.ConversationID, data)
			if err != nil {
				klog.V(1).Infof("[Entities] PublishMessage failed. Err: %v\n", err)
			}
		}
	}

	return nil
}

func (h *Handler) TeardownConversation(tm *shared.TeardownResponse) error {
	conversationId := tm.TeardownMessage.Message.Data.ConversationID
	klog.V(2).Infof("TeardownConversation - conversationID: %s\n", conversationId)
	delete(h.cache, conversationId)
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

func (h *Handler) convertMessageReferenceToSlice(conversationId string, msgRefs []sdkinterfaces.MessageReference) []interfaces.Message {
	tmp := make([]interfaces.Message, 0)

	cache := h.cache[conversationId]
	if cache == nil {
		tmp = append(tmp, interfaces.Message{
			Text: interfaces.MessageNotFound,
		})
		return tmp
	}

	for _, msgRef := range msgRefs {
		cacheMessage, err := cache.Find(msgRef.ID)
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

func (h *Handler) topicStats(conversationId string, curTopic *sdkinterfaces.Topic, length StatisticType) int64 {
	ctx := context.Background()

	var retValue int64

	_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		myQuery := utils.ReplaceIndexes(h.topicQuery(length))
		result, err := tx.Run(ctx, myQuery, map[string]any{
			"conversation_id": conversationId,
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

func (h *Handler) trackerStats(conversationId string, curTracker *sdkinterfaces.Tracker, length StatisticType) int64 {
	ctx := context.Background()

	var retValue int64

	_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := h.trackerQuery(length)
		klog.V(2).Infof("QUERY: \n%s\n", query)
		myQuery := utils.ReplaceIndexes(query)
		result, err := tx.Run(ctx, myQuery, map[string]any{
			"conversation_id": conversationId,
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

func (h *Handler) entityStats(conversationId string, entity *sdkinterfaces.Entity, match *sdkinterfaces.EntityMatch, length StatisticType) int64 {
	ctx := context.Background()

	var retValue int64

	_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := h.entityQuery(length)
		klog.V(2).Infof("QUERY: \n%s\n", query)
		myQuery := utils.ReplaceIndexes(query)
		result, err := tx.Run(ctx, myQuery, map[string]any{
			"conversation_id": conversationId,
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
