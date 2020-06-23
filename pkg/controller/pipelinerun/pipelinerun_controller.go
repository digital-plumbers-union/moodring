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
	"context"
	"crypto/sha1"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

var log = logf.Log.WithName("controller_pipelinerun")

// Add creates a new PipelineRun Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// used as an in-memory store to track pending runs.
type pipelineRunTracker map[string]State

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, gitBaseURL string) reconcile.Reconciler {
	return &ReconcilePipelineRun{client: mgr.GetClient(), scheme: mgr.GetScheme(), gitBaseURL: gitBaseURL, scmFactory: createClient, pipelineRuns: make(pipelineRunTracker)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("pipelinerun-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &pipelinev1.PipelineRun{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

// ReconcilePipelineRun reconciles a PipelineRun object
type ReconcilePipelineRun struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client       client.Client
	gitBaseURL   string
	scheme       *runtime.Scheme
	scmFactory   scmClientFactory
	pipelineRuns pipelineRunTracker
}

// Reconcile reads that state of the cluster for a PipelineRun object and makes changes based on the state read
// and what is in the PipelineRun.Spec
func (r *ReconcilePipelineRun) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PipelineRun")
	ctx := context.Background()

	// Fetch the PipelineRun instance
	pipelineRun := &pipelinev1.PipelineRun{}
	err := r.client.Get(ctx, request.NamespacedName, pipelineRun)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if !isNotifiablePipelineRun(pipelineRun) {
		reqLogger.Info("not a notifiable pipeline run")
		return reconcile.Result{}, nil
	}

	var repo string
	var sha string

	res, err := findGitResource(pipelineRun)
	if err != nil {
		reqLogger.Info("failed to find a git resource")
	}
	if res != nil {
		repo, sha, err = getRepoAndSHAFromResource(res)
		if err != nil {
			reqLogger.Error(err, "failed to parse the URL and SHA from git resource")
			return reconcile.Result{}, nil
		}
	} else {
		repo, sha, err = getRepoAndSHAFromAnnotations(pipelineRun.ObjectMeta)
		if err != nil {
			reqLogger.Error(err, "failed to parse the URL and SHA from annotations")
			return reconcile.Result{}, nil
		}
		reqLogger.Info("parsed repo and sha from annotations")
	}

	key := keyForCommit(repo, sha)
	status := getPipelineRunState(pipelineRun)
	last, ok := r.pipelineRuns[key]
	// This uses the in-memory state to retain an original pending state to
	// avoid duplicate API calls.
	if ok {
		if status == last {
			return reconcile.Result{}, nil
		}
	}

	secret, err := getAuthSecret(r.client, request.Namespace)
	if err != nil {
		reqLogger.Error(err, "failed to get an authSecret")
		return reconcile.Result{}, nil
	}

	client := r.scmFactory(secret)
	commitStatusInput := getCommitStatusInput(pipelineRun)
	reqLogger.Info("creating a commit status for", "resource", res, "repo", repo, "sha", sha, "status", commitStatusInput)
	_, _, err = client.Repositories.CreateStatus(ctx, repo, sha, commitStatusInput)
	if err != nil {
		reqLogger.Error(err, "failed to create the commit status")
		return reconcile.Result{}, err
	}
	r.pipelineRuns[key] = status
	return reconcile.Result{}, nil
}

func keyForCommit(repo, sha string) string {
	return sha1String(fmt.Sprintf("%s:%s", repo, sha))
}

func sha1String(s string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))
}
