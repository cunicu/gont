// Package tc contains the options for configuring per-interface Traffic Control (TC) queuing disciplines
package tc

type Probability struct {
	Probability float32
	Correlation float32
}
