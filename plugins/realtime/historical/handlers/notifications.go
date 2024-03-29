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

	interfacessdk "github.com/dvonthenen/enterprise-conversation-application/pkg/middleware-plugin-sdk/interfaces"
	shared "github.com/dvonthenen/enterprise-conversation-application/pkg/shared"
	utils "github.com/dvonthenen/enterprise-conversation-application/pkg/utils"

	interfaces "github.com/dvonthenen/enterprise-conversation-plugins/plugins/realtime/historical/interfaces"
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
	ctx := context.Background()

	for _, curTopic := range tr.TopicResponse.Topics {
		// housekeeping
		atLeastOnce := false
		var msg *interfaces.AppSpecificHistorical

		// get past instances
		_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			myQuery := utils.ReplaceIndexes(`
				MATCH (t:Topic)-[x:TOPIC_MESSAGE_REF]-(m:Message)-[y:SPOKE]-(u:User)
				WHERE x.#conversation_index# <> $conversation_id AND y.#conversation_index# <> $conversation_id AND t.value = $topic_phrases
				RETURN t, x, m, y, u ORDER BY x.created DESC LIMIT 5`)
			result, err := tx.Run(ctx, myQuery, map[string]any{
				"conversation_id": tr.ConversationID,
				"topic_phrases":   strings.ToLower(curTopic.Phrases),
			})
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				// only init once!
				if !atLeastOnce {
					klog.V(2).Infof("Check Previous Topics\n")
					klog.V(2).Infof("----------------------------------------\n")

					msg = &interfaces.AppSpecificHistorical{
						Type: sdkinterfaces.MessageTypeUserDefined,
						Metadata: interfaces.Metadata{
							Type: interfaces.AppSpecificMessageTypeHistorical,
						},
						Historical: interfaces.Data{
							Type:        interfaces.UserHistoricalTypeTopic,
							Correlation: strings.ToLower(curTopic.Phrases),
							Current:     make([]interfaces.Insight, 0),
							Previous:    make([]interfaces.Insight, 0),
						},
					}

					msg.Historical.Current = append(msg.Historical.Current, interfaces.Insight{
						Correlation: strings.ToLower(curTopic.Phrases),
						Messages:    h.convertMessageReferenceToSlice(tr.ConversationID, curTopic.MessageReferences),
					})

					atLeastOnce = true
				}

				// prevTopic := result.Record().Values[0].(neo4j.Node)
				message := result.Record().Values[2].(neo4j.Node)
				relationship := result.Record().Values[1].(neo4j.Relationship)
				user := result.Record().Values[4].(neo4j.Node)

				klog.V(2).Infof("Previous Topic\n")
				klog.V(2).Infof("Author: %s / %s\n", user.Props["name"].(string), user.Props["email"].(string))
				klog.V(2).Infof("Topic Match: %s\n", relationship.Props["value"].(string))
				klog.V(2).Infof("Corresponding sentence: %s\n", message.Props["content"].(string))

				msg.Historical.Previous = append(msg.Historical.Previous, interfaces.Insight{
					Correlation: strings.ToLower(relationship.Props["value"].(string)),
					Messages: []interfaces.Message{
						interfaces.Message{
							ID:   message.Props["messageId"].(string),
							Text: message.Props["content"].(string),
							Author: interfaces.Author{
								ID:    user.Props["userId"].(string),
								Name:  user.Props["name"].(string),
								Email: user.Props["email"].(string),
							},
						},
					},
				})
			}

			if atLeastOnce {
				klog.V(2).Infof("----------------------------------------\n")
			}

			return nil, result.Err()
		})
		if err != nil {
			klog.V(1).Infof("[Topics] ExecuteRead failed. Err: %v\n", err)
			return err
		}

		/*
			If there is at least one Message or Insight that has triggered this Tracker, then
			send your High-level Application message back to the Dataminer component
		*/
		if atLeastOnce {
			data, err := json.Marshal(*msg)
			if err != nil {
				klog.V(1).Infof("[Topics] json.Marshal failed. Err: %v\n", err)
			}

			err = (*h.msgPublisher).PublishMessage(tr.ConversationID, data)
			if err != nil {
				klog.V(1).Infof("[Topics] PublishMessage failed. Err: %v\n", err)
			}
		}
	}

	return nil
}

func (h *Handler) TrackerResponseMessage(tr *shared.TrackerResponse) error {
	ctx := context.Background()

	for _, curTracker := range tr.TrackerResponse.Trackers {
		// housekeeping
		addCurrentMessageAlready := false
		atLeastOnceMessage := false
		atLeastOnceInsight := false
		var msg *interfaces.AppSpecificHistorical

		// get past messages
		_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			myQuery := utils.ReplaceIndexes(`
				MATCH (t:Tracker)-[x:TRACKER_MESSAGE_REF]-(m:Message)-[y:SPOKE]-(u:User)
				WHERE x.#conversation_index# <> $conversation_id AND y.#conversation_index# <> $conversation_id AND t.name = $tracker_name
				RETURN t, x, m, y, u ORDER BY x.created DESC LIMIT 5`)
			result, err := tx.Run(ctx, myQuery, map[string]any{
				"conversation_id": tr.ConversationID,
				"tracker_name":    strings.ToLower(curTracker.Name),
			})
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				// only init once!
				if !atLeastOnceMessage {
					klog.V(2).Infof("Check Previous Tracker [Message]\n")
					klog.V(2).Infof("----------------------------------------\n")

					msg = &interfaces.AppSpecificHistorical{
						Type: sdkinterfaces.MessageTypeUserDefined,
						Metadata: interfaces.Metadata{
							Type: interfaces.AppSpecificMessageTypeHistorical,
						},
						Historical: interfaces.Data{
							Type:        interfaces.UserHistoricalTypeTracker,
							Correlation: strings.ToLower(curTracker.Name),
							Current:     make([]interfaces.Insight, 0),
							Previous:    make([]interfaces.Insight, 0),
						},
					}
					atLeastOnceMessage = true
				}

				message := result.Record().Values[2].(neo4j.Node)
				relationship := result.Record().Values[1].(neo4j.Relationship)
				user := result.Record().Values[4].(neo4j.Node)

				klog.V(2).Infof("Previous Tracker [Message]\n")
				klog.V(2).Infof("Author: %s / %s\n", user.Props["name"].(string), user.Props["email"].(string))
				klog.V(2).Infof("Tracker Match: %s\n", relationship.Props["value"].(string))
				klog.V(2).Infof("Corresponding sentence: %s\n", message.Props["content"].(string))

				for _, match := range curTracker.Matches {
					if !addCurrentMessageAlready {
						msg.Historical.Current = append(msg.Historical.Current, interfaces.Insight{
							Correlation: strings.ToLower(match.Value),
							Messages:    h.convertMessageRefsToSlice(tr.ConversationID, match.MessageRefs),
						})
						addCurrentMessageAlready = true
					}

					msg.Historical.Previous = append(msg.Historical.Previous, interfaces.Insight{
						Correlation: strings.ToLower(relationship.Props["value"].(string)),
						Messages: []interfaces.Message{
							interfaces.Message{
								ID:   message.Props["messageId"].(string),
								Text: message.Props["content"].(string),
								Author: interfaces.Author{
									ID:    user.Props["userId"].(string),
									Name:  user.Props["name"].(string),
									Email: user.Props["email"].(string),
								},
							},
						},
					})
				}
			}

			if atLeastOnceMessage {
				klog.V(2).Infof("----------------------------------------\n")
			}

			return nil, result.Err()
		})
		if err != nil {
			klog.V(1).Infof("[Tracker] ExecuteRead failed. Err: %v\n", err)
			return err
		}

		// get past insights
		_, err = (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			myQuery := utils.ReplaceIndexes(`
				MATCH (t:Tracker)-[x:TRACKER_INSIGHT_REF]-(i:Insight)-[y:SPOKE]-(u:User)
				WHERE x.#conversation_index# <> $conversation_id AND y.#conversation_index# <> $conversation_id AND t.name = $tracker_name
				RETURN t, x, i, y, u ORDER BY x.created DESC LIMIT 5`)
			result, err := tx.Run(ctx, myQuery, map[string]any{
				"conversation_id": tr.ConversationID,
				"tracker_name":    strings.ToLower(curTracker.Name),
			})
			if err != nil {
				return nil, err
			}

			// insight
			for result.Next(ctx) {
				// only init once!
				if !atLeastOnceInsight {
					klog.V(2).Infof("Check Previous Tracker [Insight]\n")
					klog.V(2).Infof("----------------------------------------\n")

					msg = &interfaces.AppSpecificHistorical{
						Type: sdkinterfaces.MessageTypeUserDefined,
						Metadata: interfaces.Metadata{
							Type: interfaces.AppSpecificMessageTypeHistorical,
						},
						Historical: interfaces.Data{
							Type:        interfaces.UserHistoricalTypeTracker,
							Correlation: strings.ToLower(curTracker.Name),
							Current:     make([]interfaces.Insight, 0),
							Previous:    make([]interfaces.Insight, 0),
						},
					}
					atLeastOnceInsight = true
				}

				insight := result.Record().Values[2].(neo4j.Node)
				relationship := result.Record().Values[1].(neo4j.Relationship)
				user := result.Record().Values[4].(neo4j.Node)

				klog.V(2).Infof("Previous Tracker [Insight]\n")
				klog.V(2).Infof("Author: %s / %s\n", user.Props["name"].(string), user.Props["email"].(string))
				klog.V(2).Infof("Tracker Match: %s\n", relationship.Props["value"].(string))
				klog.V(2).Infof("Corresponding sentence: %s\n", insight.Props["content"].(string))

				for _, match := range curTracker.Matches {
					if !addCurrentMessageAlready {
						msg.Historical.Current = append(msg.Historical.Current, interfaces.Insight{
							Correlation: strings.ToLower(match.Value),
							Messages:    h.convertInsightRefsToSlice(tr.ConversationID, match.InsightRefs),
						})
						addCurrentMessageAlready = true
					}

					msg.Historical.Previous = append(msg.Historical.Previous, interfaces.Insight{
						Correlation: strings.ToLower(relationship.Props["value"].(string)),
						Messages: []interfaces.Message{
							interfaces.Message{
								ID:   insight.Props["insightId"].(string),
								Text: insight.Props["content"].(string),
								Author: interfaces.Author{
									ID:    user.Props["userId"].(string),
									Name:  user.Props["name"].(string),
									Email: user.Props["email"].(string),
								},
							},
						},
					})
				}
			}

			if atLeastOnceInsight {
				klog.V(2).Infof("----------------------------------------\n")
			}

			return nil, result.Err()
		})
		if err != nil {
			klog.V(1).Infof("[Tracker] ExecuteRead failed. Err: %v\n", err)
			return err
		}

		/*
			If there is at least one Message or Insight that has triggered this Tracker, then
			send your High-level Application message back to the Dataminer component
		*/
		if atLeastOnceMessage || atLeastOnceInsight {
			data, err := json.Marshal(*msg)
			if err != nil {
				klog.V(1).Infof("[Tracker] json.Marshal failed. Err: %v\n", err)
			}

			err = (*h.msgPublisher).PublishMessage(tr.ConversationID, data)
			if err != nil {
				klog.V(1).Infof("[Tracker] PublishMessage failed. Err: %v\n", err)
			}
		}
	}

	return nil
}

func (h *Handler) EntityResponseMessage(er *shared.EntityResponse) error {
	ctx := context.Background()

	for _, entity := range er.EntityResponse.Entities {
		for _, match := range entity.Matches {
			// housekeeping
			atLeastOnce := false
			addCurrentMessageAlready := false
			var msg *interfaces.AppSpecificHistorical

			// get past instances
			_, err := (*h.session).ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
				myQuery := utils.ReplaceIndexes(`
				MATCH (e:Entity)-[x:ENTITY_MESSAGE_REF]-(m:Message)-[y:SPOKE]-(u:User)
					WHERE x.#conversation_index# <> $conversation_id AND y.#conversation_index# <> $conversation_id AND e.category = $entity_category AND e.type = $entity_type AND e.subType = $entity_subtype AND e.value = $entity_value
				RETURN e, x, m, y, u ORDER BY x.created DESC LIMIT 5`)
				result, err := tx.Run(ctx, myQuery, map[string]any{
					"conversation_id": er.ConversationID,
					"entity_category": strings.ToLower(entity.Category),
					"entity_type":     strings.ToLower(entity.Type),
					"entity_subtype":  strings.ToLower(entity.SubType),
					"entity_value":    strings.ToLower(match.DetectedValue),
				})
				if err != nil {
					return nil, err
				}

				for result.Next(ctx) {
					// only init once!
					if !atLeastOnce {
						klog.V(2).Infof("Check Previous Entities\n")
						klog.V(2).Infof("----------------------------------------\n")

						msg = &interfaces.AppSpecificHistorical{
							Type: sdkinterfaces.MessageTypeUserDefined,
							Metadata: interfaces.Metadata{
								Type: interfaces.AppSpecificMessageTypeHistorical,
							},
							Historical: interfaces.Data{
								Type:        interfaces.UserHistoricalTypeEntity,
								Correlation: fmt.Sprintf("%s/%s/%s/%s", strings.ToLower(entity.Category), strings.ToLower(entity.Type), strings.ToLower(entity.SubType), strings.ToLower(match.DetectedValue)),
								Current:     make([]interfaces.Insight, 0),
								Previous:    make([]interfaces.Insight, 0),
							},
						}
						atLeastOnce = true
					}

					message := result.Record().Values[2].(neo4j.Node)
					relationship := result.Record().Values[1].(neo4j.Relationship)
					user := result.Record().Values[4].(neo4j.Node)

					klog.V(2).Infof("Previous Entity\n")
					klog.V(2).Infof("Author: %s / %s\n", user.Props["name"].(string), user.Props["email"].(string))
					klog.V(2).Infof("Entity Match: %s\n", relationship.Props["value"].(string))
					klog.V(2).Infof("Corresponding sentence: %s\n", message.Props["content"].(string))

					for _, match := range entity.Matches {
						if !addCurrentMessageAlready {
							msg.Historical.Current = append(msg.Historical.Current, interfaces.Insight{
								Correlation: strings.ToLower(match.DetectedValue),
								Messages:    h.convertMessageRefsToSlice(er.ConversationID, match.MessageRefs),
							})
							addCurrentMessageAlready = true
						}

						msg.Historical.Previous = append(msg.Historical.Previous, interfaces.Insight{
							Correlation: strings.ToLower(relationship.Props["value"].(string)),
							Messages: []interfaces.Message{
								interfaces.Message{
									ID:   message.Props["messageId"].(string),
									Text: message.Props["content"].(string),
									Author: interfaces.Author{
										ID:    user.Props["userId"].(string),
										Name:  user.Props["name"].(string),
										Email: user.Props["email"].(string),
									},
								},
							},
						})
					}
				}

				if atLeastOnce {
					klog.V(2).Infof("----------------------------------------\n")
				}

				return nil, result.Err()
			})
			if err != nil {
				klog.V(1).Infof("[Entities] ExecuteRead failed. Err: %v\n", err)
				return err
			}

			/*
				If there is at least one Message or Insight that has triggered this Tracker, then
				send your High-level Application message back to the Dataminer component
			*/
			if atLeastOnce {
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

func (h *Handler) convertInsightRefsToSlice(conversationId string, inRefs []sdkinterfaces.InsightRef) []interfaces.Message {
	tmp := make([]interfaces.Message, 0)

	cache := h.cache[conversationId]
	if cache == nil {
		tmp = append(tmp, interfaces.Message{
			ID:   interfaces.MessageNotFound,
			Text: interfaces.MessageNotFound,
		})
		return tmp
	}

	for _, inRef := range inRefs {
		cacheMessage, err := cache.Find(inRef.ID)
		if err != nil {
			klog.V(4).Infof("Msg ID not found: %s\n", inRef.ID)
			tmp = append(tmp, interfaces.Message{
				ID:   inRef.ID,
				Text: interfaces.MessageNotFound,
			})
			continue
		}

		tmp = append(tmp, interfaces.Message{
			ID:   inRef.ID,
			Text: cacheMessage.Text,
			Author: interfaces.Author{
				ID:    cacheMessage.Author.ID,
				Name:  cacheMessage.Author.Name,
				Email: cacheMessage.Author.Email,
			},
		})
	}

	return tmp
}

func (h *Handler) convertMessageRefsToSlice(conversationId string, msgRefs []sdkinterfaces.MessageRef) []interfaces.Message {
	tmp := make([]interfaces.Message, 0)

	cache := h.cache[conversationId]
	if cache == nil {
		tmp = append(tmp, interfaces.Message{
			ID:   interfaces.MessageNotFound,
			Text: interfaces.MessageNotFound,
		})
		return tmp
	}

	for _, msgRef := range msgRefs {
		cacheMessage, err := cache.Find(msgRef.ID)
		if err != nil {
			klog.V(4).Infof("Msg ID not found: %s\n", msgRef.ID)
			tmp = append(tmp, interfaces.Message{
				ID:   msgRef.ID,
				Text: interfaces.MessageNotFound,
			})
			continue
		}

		tmp = append(tmp, interfaces.Message{
			ID:   msgRef.ID,
			Text: cacheMessage.Text,
			Author: interfaces.Author{
				ID:    cacheMessage.Author.ID,
				Name:  cacheMessage.Author.Name,
				Email: cacheMessage.Author.Email,
			},
		})
	}

	return tmp
}

func (h *Handler) convertMessageReferenceToSlice(conversationId string, msgRefs []sdkinterfaces.MessageReference) []interfaces.Message {
	tmp := make([]interfaces.Message, 0)

	cache := h.cache[conversationId]
	if cache == nil {
		tmp = append(tmp, interfaces.Message{
			ID:   interfaces.MessageNotFound,
			Text: interfaces.MessageNotFound,
			Author: interfaces.Author{
				ID:    interfaces.MessageNotFound,
				Name:  interfaces.MessageNotFound,
				Email: interfaces.MessageNotFound,
			},
		})
		return tmp
	}

	for _, msgRef := range msgRefs {
		cacheMessage, err := cache.Find(msgRef.ID)
		if err != nil {
			klog.V(4).Infof("Msg ID not found: %s\n", msgRef.ID)
			tmp = append(tmp, interfaces.Message{
				ID:   msgRef.ID,
				Text: interfaces.MessageNotFound,
			})
			continue
		}

		tmp = append(tmp, interfaces.Message{
			ID:   msgRef.ID,
			Text: cacheMessage.Text,
			Author: interfaces.Author{
				ID:    cacheMessage.Author.ID,
				Name:  cacheMessage.Author.Name,
				Email: cacheMessage.Author.Email,
			},
		})
	}

	return tmp
}
