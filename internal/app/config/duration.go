package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a time.Duration that marshals to and from Go duration strings
// such as "250ms" or "15s" in YAML.
type Duration time.Duration

// UnmarshalYAML decodes a duration string into the receiver.
func (d *Duration) UnmarshalYAML(node *yaml.Node) (err error) {
	var raw string
	if err = node.Decode(&raw); err != nil {
		return
	}

	var parsed time.Duration
	if parsed, err = time.ParseDuration(raw); err != nil {
		err = fmt.Errorf("invalid duration %q: %w", raw, err)
		return
	}
	*d = Duration(parsed)
	return
}

// MarshalYAML encodes the duration as a Go duration string.
func (d Duration) MarshalYAML() (value any, err error) {
	value = time.Duration(d).String()
	return
}

// Std converts the wrapper back to a time.Duration for use by the runtime.
func (d Duration) Std() (std time.Duration) {
	std = time.Duration(d)
	return
}
