// Copyright 2020 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipelinerun

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tb "github.com/tektoncd/pipeline/test/builder"
)

func TestGetPipelineRunState(t *testing.T) {
	statusTests := []struct {
		conditionType   apis.ConditionType
		conditionStatus corev1.ConditionStatus
		reason          string
		want            State
	}{
		{apis.ConditionSucceeded, corev1.ConditionTrue, "Succeeded", Successful},
		{apis.ConditionSucceeded, corev1.ConditionUnknown, "Running", Pending},
		{apis.ConditionSucceeded, corev1.ConditionFalse, "Failed", Failed},
		{apis.ConditionSucceeded, corev1.ConditionFalse, pipelinev1.PipelineRunSpecStatusCancelled, Cancelled},
		{apis.ConditionSucceeded, corev1.ConditionUnknown, pipelinev1.PipelineRunSpecStatusCancelled, Cancelled},
	}

	for _, tt := range statusTests {
		s := getPipelineRunState(makePipelineRunWithCondition(tt.conditionType, tt.conditionStatus, tt.reason))
		if s != tt.want {
			t.Errorf("getPipelineRunState(%s) got %v, want %v", tt.conditionStatus, s, tt.want)
		}
	}
}

// TODO: use non-string type for Reason, seems its not exposed on v1alpha1
func makePipelineRunWithCondition(s apis.ConditionType, c corev1.ConditionStatus, r string) *pipelinev1.PipelineRun {
	return tb.PipelineRun(pipelineRunName,
		tb.PipelineRunNamespace(testNamespace), tb.PipelineRunSpec("tomatoes"),
		tb.PipelineRunStatus(tb.PipelineRunStatusCondition(apis.Condition{Type: s, Status: c, Reason: r}),
			tb.PipelineRunTaskRunsStatus("trname", &pipelinev1.PipelineRunTaskRunStatus{PipelineTaskName: "task-1"})))
}
