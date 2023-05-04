// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import (
	interfacessdk "github.com/dvonthenen/enterprise-reference-implementation/pkg/middleware-plugin-sdk/interfaces"
	utils "github.com/dvonthenen/enterprise-reference-implementation/pkg/utils"
	symbl "github.com/dvonthenen/symbl-go-sdk/pkg/client"
	neo4j "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

/*
	Handler for messages
*/
type HandlerOptions struct {
	Session     *neo4j.SessionWithContext // retrieve insights
	SymblClient *symbl.RestClient
}

type Handler struct {
	// properties
	cache map[string]*utils.MessageCache

	// housekeeping
	session      *neo4j.SessionWithContext
	symblClient  *symbl.RestClient
	msgPublisher *interfacessdk.MessagePublisher
}
