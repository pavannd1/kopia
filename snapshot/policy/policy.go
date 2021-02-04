package policy

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/kopia/kopia/snapshot"
)

// ErrPolicyNotFound is returned when the policy is not found.
var ErrPolicyNotFound = errors.New("policy not found")

// Policy describes snapshot policy for a single source.
type Policy struct {
	Labels              map[string]string   `json:"-"`
	RetentionPolicy     RetentionPolicy     `json:"retention,omitempty"`
	FilesPolicy         FilesPolicy         `json:"files,omitempty"`
	ErrorHandlingPolicy ErrorHandlingPolicy `json:"errorHandling,omitempty"`
	SchedulingPolicy    SchedulingPolicy    `json:"scheduling,omitempty"`
	CompressionPolicy   CompressionPolicy   `json:"compression,omitempty"`
	Actions             ActionsPolicy       `json:"actions"`
	NoParent            bool                `json:"noParent,omitempty"`
}

func (p *Policy) String() string {
	var buf bytes.Buffer

	e := json.NewEncoder(&buf)
	e.SetIndent("", "  ")

	if err := e.Encode(p); err != nil {
		return "unable to policy as JSON: " + err.Error()
	}

	return buf.String()
}

// ID returns globally unique identifier of the policy.
func (p *Policy) ID() string {
	return p.Labels["id"]
}

// Target returns the snapshot.SourceInfo describing username, host and path targeted by the policy.
func (p *Policy) Target() snapshot.SourceInfo {
	return snapshot.SourceInfo{
		Host:     p.Labels["hostname"],
		UserName: p.Labels["username"],
		Path:     p.Labels["path"],
	}
}

// MergePolicies computes the policy by applying the specified list of policies in order.
func MergePolicies(policies []*Policy) *Policy {
	var merged Policy

	for _, p := range policies {
		if p.NoParent {
			return &merged
		}

		merged.RetentionPolicy.Merge(p.RetentionPolicy)
		merged.FilesPolicy.Merge(p.FilesPolicy)
		merged.ErrorHandlingPolicy.Merge(p.ErrorHandlingPolicy)
		merged.SchedulingPolicy.Merge(p.SchedulingPolicy)
		merged.CompressionPolicy.Merge(p.CompressionPolicy)
		merged.Actions.Merge(p.Actions)
	}

	// Merge default expiration policy.
	merged.RetentionPolicy.Merge(defaultRetentionPolicy)
	merged.FilesPolicy.Merge(defaultFilesPolicy)
	merged.ErrorHandlingPolicy.Merge(defaultErrorHandlingPolicy)
	merged.SchedulingPolicy.Merge(defaultSchedulingPolicy)
	merged.CompressionPolicy.Merge(defaultCompressionPolicy)
	merged.Actions.Merge(defaultActionsPolicy)

	if len(policies) > 0 {
		merged.Actions.MergeNonInheritable(policies[0].Actions)
	}

	return &merged
}

// ValidatePolicy returns error if the given policy is invalid.
// Currently, only SchedulingPolicy is validated.
func ValidatePolicy(pol *Policy) error {
	return ValidateSchedulingPolicy(pol.SchedulingPolicy)
}

func intPtr(n int) *int {
	return &n
}
