// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

//nolint:tagliatelle
type TsharkOutput struct {
	Index  string       `json:"_index,omitempty"`
	Type   string       `json:"_type,omitempty"`
	Source TsharkSource `json:"_source,omitempty"`
}

type TsharkTraceDataTree struct {
	Filtered string `json:"filtered,omitempty"`
}

//nolint:tagliatelle
type TsharkTrace struct {
	Message  string              `json:"trace.message,omitempty"`
	Type     string              `json:"trace.type,omitempty"`
	Pid      string              `json:"trace.pid,omitempty"`
	Function string              `json:"trace.function,omitempty"`
	File     string              `json:"trace.file,omitempty"`
	Line     string              `json:"trace.line,omitempty"`
	Data     string              `json:"trace.data,omitempty"`
	DataTree TsharkTraceDataTree `json:"trace.data_tree,omitempty"`
}
type TsharkLayers struct {
	Trace TsharkTrace `json:"trace,omitempty"`
}

type TsharkSource struct {
	Layers TsharkLayers `json:"layers,omitempty"`
}
