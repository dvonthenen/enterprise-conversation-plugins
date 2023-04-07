// Copyright 2023 Enterprise Conversation Plugins contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import "errors"

var (
	// ErrUnhandledMessage runhandled message from symbl-proxy-dataminer
	ErrUnhandledMessage = errors.New("unhandled message from symbl-proxy-dataminer")
)
