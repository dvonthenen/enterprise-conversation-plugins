// Copyright 2023 Symbl.ai SDK contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package handlers

import "errors"

type StatisticType int

const (
	StatType1Month StatisticType = iota
	StatType1Week
	StatType2Days
	StatType1Day
	StatType4Hours
	StatType1Hour
	StatType30Mins
)

var (
	// ErrUnhandledMessage runhandled message from symbl-proxy-dataminer
	ErrUnhandledMessage = errors.New("unhandled message from symbl-proxy-dataminer")
)
