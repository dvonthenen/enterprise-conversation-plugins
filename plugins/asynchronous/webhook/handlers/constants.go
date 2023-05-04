// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import "errors"

var (
	// ErrInvalidInput required input was not found
	ErrInvalidInput = errors.New("required input was not found")

	// ErrUnhandledMessage unhandled message from asynchronous email plugin
	ErrUnhandledMessage = errors.New("unhandled message from asynchronous email plugin")

	// ErrConversationNotFound conversation not found
	ErrConversationNotFound = errors.New("conversation not found")
)
