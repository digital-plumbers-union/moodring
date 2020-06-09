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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	notifiableName        string = "moodring.dpu.sh/git-status"
	statusContextName     string = "moodring.dpu.sh/status-context"
	statusTargetURLName   string = "moodring.dpu.sh/status-target-url"
	statusDescriptionName string = "moodring.dpu.sh/status-description"
	revisionName          string = "moodring.dpu.sh/commit"
	repoName              string = "moodring.dpu.sh/repo"
)

func getRepoAndSHAFromAnnotations(m metav1.ObjectMeta) (string, string, error) {
	// check for annotations
	if !(metav1.HasAnnotation(m, repoName) && metav1.HasAnnotation(m, revisionName)) {
		return "", "", fmt.Errorf("Annotations not present: %s and %s are required", repoName, revisionName)
	}
	repo, err := extractRepoFromGitHubURL(m.Annotations[repoName])
	if err != nil {
		return "", "", fmt.Errorf("getRepoAndSHAFromAnnotations failed: %w", err)
	}

	return strings.TrimSuffix(repo, ".git"), m.Annotations[revisionName], nil
}
