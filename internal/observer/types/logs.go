// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/json"
	"fmt"
)

// ComponentSearchScope defines the search scope for component logs
// Matches OpenAPI ComponentSearchScope schema
type ComponentSearchScope struct {
	Namespace   string `json:"namespace" validate:"required"`
	Project     string `json:"project,omitempty"`
	Component   string `json:"component,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// WorkflowSearchScope defines the search scope for workflow run logs
// Matches OpenAPI WorkflowSearchScope schema
type WorkflowSearchScope struct {
	Namespace       string `json:"namespace" validate:"required"`
	WorkflowRunName string `json:"workflowRunName,omitempty"`
	TaskName        string `json:"taskName,omitempty"`
}

// SystemSearchScope defines the search scope for OpenChoreo's own system
// components (control-plane, data-plane, workflow-plane, and
// observability-plane infrastructure), as opposed to tenant workloads
// (ComponentSearchScope) or workflow runs (WorkflowSearchScope).
// Matches OpenAPI SystemSearchScope schema.
type SystemSearchScope struct {
	Plane     string `json:"plane" validate:"required"`
	Cluster   string `json:"cluster,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Workload  string `json:"workload,omitempty"`
	Container string `json:"container,omitempty"`
}

// LogsQueryRequest represents the request body for POST /api/v1/logs/query
// Matches OpenAPI LogsQueryRequest schema
type LogsQueryRequest struct {
	// SearchScope defines where to search for logs (component or workflow)
	SearchScope *SearchScope `json:"searchScope" validate:"required"`

	// Time range for the query (required)
	StartTime string `json:"startTime" validate:"required"`
	EndTime   string `json:"endTime" validate:"required"`

	// Optional filters
	SearchPhrase string   `json:"searchPhrase,omitempty"`
	LogLevels    []string `json:"logLevels,omitempty"`

	// Pagination and sorting
	Limit     int    `json:"limit,omitempty"`
	SortOrder string `json:"sortOrder,omitempty"` // asc or desc, default: desc
}

// SearchScope is a union type for component, workflow, or system search scope.
// Implements oneOf from OpenAPI spec - can be ComponentSearchScope, WorkflowSearchScope, or SystemSearchScope.
type SearchScope struct {
	Component *ComponentSearchScope `json:"-"`
	Workflow  *WorkflowSearchScope  `json:"-"`
	System    *SystemSearchScope    `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling to handle oneOf
// The JSON can be a ComponentSearchScope, WorkflowSearchScope, or SystemSearchScope directly.
func (s *SearchScope) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a map to check for distinguishing fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse searchScope: %w", err)
	}

	// Check for distinguishing fields:
	// - plane indicates SystemSearchScope
	// - workflowRunName indicates WorkflowSearchScope
	// - project, component, or environment indicates ComponentSearchScope
	_, hasPlane := raw["plane"]
	_, hasWorkflowRunName := raw["workflowRunName"]
	_, hasProject := raw["project"]
	_, hasComponent := raw["component"]
	_, hasEnvironment := raw["environment"]

	// Reject mixed oneOf: plane cannot coexist with the other scopes' distinguishing fields, and workflowRunName cannot coexist with component-specific fields.
	if hasPlane && (hasWorkflowRunName || hasProject || hasComponent || hasEnvironment) {
		return fmt.Errorf("searchScope cannot mix plane with workflowRunName/project/component/environment")
	}
	if hasWorkflowRunName && (hasProject || hasComponent || hasEnvironment) {
		return fmt.Errorf("searchScope cannot mix workflowRunName with project/component/environment")
	}

	if hasPlane {
		var systemScope SystemSearchScope
		if err := json.Unmarshal(data, &systemScope); err != nil {
			return fmt.Errorf("failed to unmarshal as SystemSearchScope: %w", err)
		}
		s.System = &systemScope
		return nil
	}

	if hasWorkflowRunName {
		var workflowScope WorkflowSearchScope
		if err := json.Unmarshal(data, &workflowScope); err != nil {
			return fmt.Errorf("failed to unmarshal as WorkflowSearchScope: %w", err)
		}
		s.Workflow = &workflowScope
		return nil
	}

	// Check for component-specific fields
	if hasProject || hasComponent || hasEnvironment {
		var componentScope ComponentSearchScope
		if err := json.Unmarshal(data, &componentScope); err != nil {
			return fmt.Errorf("failed to unmarshal as ComponentSearchScope: %w", err)
		}
		s.Component = &componentScope
		return nil
	}

	// If only namespace is present, default to ComponentSearchScope
	// (both types require namespace, but component scope is more common for namespace-only queries)
	var componentScope ComponentSearchScope
	if err := json.Unmarshal(data, &componentScope); err != nil {
		return fmt.Errorf("failed to unmarshal searchScope: %w", err)
	}
	s.Component = &componentScope
	return nil
}

// MarshalJSON implements custom JSON marshaling
func (s *SearchScope) MarshalJSON() ([]byte, error) {
	set := 0
	if s.Component != nil {
		set++
	}
	if s.Workflow != nil {
		set++
	}
	if s.System != nil {
		set++
	}
	if set != 1 {
		return nil, fmt.Errorf("searchScope must contain exactly one of component, workflow, or system")
	}
	if s.Component != nil {
		return json.Marshal(s.Component)
	}
	if s.Workflow != nil {
		return json.Marshal(s.Workflow)
	}
	return json.Marshal(s.System)
}

// LogMetadata contains metadata for a log entry
// Used for both component and workflow logs
// Matches OpenAPI ComponentLogEntry.metadata schema (workflow logs use a subset)
type LogMetadata struct {
	// Component-specific fields (empty for workflow logs)
	ComponentName   string `json:"componentName,omitempty"`
	ProjectName     string `json:"projectName,omitempty"`
	EnvironmentName string `json:"environmentName,omitempty"`
	NamespaceName   string `json:"namespaceName,omitempty"`
	ComponentUID    string `json:"componentUid,omitempty"`
	ProjectUID      string `json:"projectUid,omitempty"`
	EnvironmentUID  string `json:"environmentUid,omitempty"`
	ContainerName   string `json:"containerName,omitempty"`
	PodName         string `json:"podName,omitempty"`
	PodNamespace    string `json:"podNamespace,omitempty"`
}

// LogEntry represents a single log entry in the response
// Used for both component and workflow logs
// Matches OpenAPI ComponentLogEntry/WorkflowLogEntry schemas
type LogEntry struct {
	Timestamp string       `json:"timestamp"`
	Log       string       `json:"log"`
	Level     string       `json:"level,omitempty"`
	Metadata  *LogMetadata `json:"metadata,omitempty"`
}

// LogsQueryResponse represents the response for POST /api/v1/logs/query
// Matches OpenAPI LogsQueryResponse schema
type LogsQueryResponse struct {
	Logs   []LogEntry `json:"logs"`
	Total  int        `json:"total"`
	TookMs int        `json:"tookMs"`
}
