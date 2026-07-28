// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchScope_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		jsonInput   string
		wantSystem  *SystemSearchScope
		wantComp    *ComponentSearchScope
		wantWf      *WorkflowSearchScope
		wantErr     bool
		errContains string
	}{
		{
			name:      "system search scope",
			jsonInput: `{"plane":"control-plane","cluster":"c1","namespace":"ns1","workload":"w1","container":"c1"}`,
			wantSystem: &SystemSearchScope{
				Plane:     "control-plane",
				Cluster:   "c1",
				Namespace: "ns1",
				Workload:  "w1",
				Container: "c1",
			},
		},
		{
			name:      "workflow search scope",
			jsonInput: `{"namespace":"ns1","workflowRunName":"run-1","taskName":"task-1"}`,
			wantWf: &WorkflowSearchScope{
				Namespace:       "ns1",
				WorkflowRunName: "run-1",
				TaskName:        "task-1",
			},
		},
		{
			name:      "component search scope with project",
			jsonInput: `{"namespace":"ns1","project":"p1","component":"c1","environment":"e1"}`,
			wantComp: &ComponentSearchScope{
				Namespace:   "ns1",
				Project:     "p1",
				Component:   "c1",
				Environment: "e1",
			},
		},
		{
			name:      "namespace only defaults to component search scope",
			jsonInput: `{"namespace":"ns1"}`,
			wantComp: &ComponentSearchScope{
				Namespace: "ns1",
			},
		},
		{
			name:        "mixed plane with project fails",
			jsonInput:   `{"plane":"control-plane","project":"p1"}`,
			wantErr:     true,
			errContains: "searchScope cannot mix plane",
		},
		{
			name:        "mixed plane with workflowRunName fails",
			jsonInput:   `{"plane":"control-plane","workflowRunName":"r1"}`,
			wantErr:     true,
			errContains: "searchScope cannot mix plane",
		},
		{
			name:        "mixed workflowRunName with project fails",
			jsonInput:   `{"workflowRunName":"r1","project":"p1"}`,
			wantErr:     true,
			errContains: "searchScope cannot mix workflowRunName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var scope SearchScope
			err := json.Unmarshal([]byte(tt.jsonInput), &scope)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantSystem, scope.System)
				assert.Equal(t, tt.wantComp, scope.Component)
				assert.Equal(t, tt.wantWf, scope.Workflow)
			}
		})
	}
}

func TestSearchScope_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		scope       SearchScope
		wantJSON    string
		wantErr     bool
		errContains string
	}{
		{
			name: "marshal system scope",
			scope: SearchScope{
				System: &SystemSearchScope{Plane: "control-plane"},
			},
			wantJSON: `{"plane":"control-plane"}`,
		},
		{
			name: "marshal component scope",
			scope: SearchScope{
				Component: &ComponentSearchScope{Namespace: "ns1"},
			},
			wantJSON: `{"namespace":"ns1"}`,
		},
		{
			name: "marshal workflow scope",
			scope: SearchScope{
				Workflow: &WorkflowSearchScope{Namespace: "ns1", WorkflowRunName: "r1"},
			},
			wantJSON: `{"namespace":"ns1","workflowRunName":"r1"}`,
		},
		{
			name:        "marshal zero scope set fails",
			scope:       SearchScope{},
			wantErr:     true,
			errContains: "must contain exactly one",
		},
		{
			name: "marshal multiple scopes set fails",
			scope: SearchScope{
				System:    &SystemSearchScope{Plane: "control-plane"},
				Component: &ComponentSearchScope{Namespace: "ns1"},
			},
			wantErr:     true,
			errContains: "must contain exactly one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(&tt.scope)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.JSONEq(t, tt.wantJSON, string(data))
			}
		})
	}
}
