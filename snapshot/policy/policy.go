package policy

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/kopia/kopia/snapshot"
)

// ErrPolicyNotFound is returned when the policy is not found.
var ErrPolicyNotFound = errors.New("policy not found")

// TargetWithPolicy wraps a policy with its target and ID.
type TargetWithPolicy struct {
	ID     string              `json:"id"`
	Target snapshot.SourceInfo `json:"target"`
	*Policy
}

// Policy describes snapshot policy for a single source.
type Policy struct {
	Labels              map[string]string   `json:"-"`
	RetentionPolicy     RetentionPolicy     `json:"retention,omitempty"`
	FilesPolicy         FilesPolicy         `json:"files,omitempty"`
	ErrorHandlingPolicy ErrorHandlingPolicy `json:"errorHandling,omitempty"`
	SchedulingPolicy    SchedulingPolicy    `json:"scheduling,omitempty"`
	CompressionPolicy   CompressionPolicy   `json:"compression,omitempty"`
	Actions             ActionsPolicy       `json:"actions"`
	LoggingPolicy       LoggingPolicy       `json:"logging"`
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
		merged.LoggingPolicy.Merge(p.LoggingPolicy)
	}

	// Merge default expiration policy.
	merged.RetentionPolicy.Merge(defaultRetentionPolicy)
	merged.FilesPolicy.Merge(defaultFilesPolicy)
	merged.ErrorHandlingPolicy.Merge(defaultErrorHandlingPolicy)
	merged.SchedulingPolicy.Merge(defaultSchedulingPolicy)
	merged.CompressionPolicy.Merge(defaultCompressionPolicy)
	merged.Actions.Merge(defaultActionsPolicy)
	merged.LoggingPolicy.Merge(defaultLoggingPolicy)

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

// validatePolicyPath validates that the provided policy path is valid and the path exists.
func validatePolicyPath(p string) error {
	if isSlashOrBackslash(p[len(p)-1]) && !isRootPath(p) {
		return errors.Errorf("path cannot end with a slash or a backslash")
	}

	return nil
}

func intPtr(n int) *int {
	return &n
}
