// Copyright 2023 Symbl.ai SDK contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package interfaces

/*
	Higher Level Application Message
*/
type Message struct {
	Text string `json:"text,omitempty"`
}

type Insight struct {
	Correlation string    `json:"correlation,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
}

type Stats struct {
	Last30Mins int64 `json:"last30Mins"`
	LastHour   int64 `json:"lastHour"`
	Last4Hours int64 `json:"last4Hour"`
	LastDay    int64 `json:"lastDay"`
	Last2Days  int64 `json:"last2Days"`
	LastWeek   int64 `json:"lastWeek"`
	LastMonth  int64 `json:"lastMonth"`
}

type Data struct {
	Type     string    `json:"type,omitempty"`
	Insights []Insight `json:"insights,omitempty"`
	Stats    Stats     `json:"stats"`
}

/*
	Please see github.com/dvonthenen/enterprise-reference-implementation/pkg/interfaces
	for required definition of this common part of the struct
*/
type Metadata struct {
	Type string `json:"type"`
}

type AppSpecificStatistical struct {
	Type        string   `json:"type"`
	Metadata    Metadata `json:"metadata"`
	Statistical Data     `json:"statistical,omitempty"`
}
