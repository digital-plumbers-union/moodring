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

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/spf13/pflag"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"

	pipelinerun "github.com/digital-plumbers-union/tekton-commit-status-tracker/pkg/controller/pipelinerun"
	"github.com/digital-plumbers-union/tekton-commit-status-tracker/version"
)

// Controller configuration
var (
	gitBaseURL  string
	metricsHost string
	metricsPort int
	port        int
	namespace   string
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	pflag.StringVar(&gitBaseURL, "git-base-url", "https://api.github.com", "base URL for git API to use")
	pflag.StringVar(&metricsHost, "metrics-host", "0.0.0.0", "address to serve metrics on")
	pflag.IntVar(&metricsPort, "metrics-port", 8383, "port to serve metrics on")
	pflag.IntVar(&port, "port", 9443, "port to serve looks on")
	pflag.StringVar(&namespace, "watch-namespace", os.Getenv("WATCH_NAMESPACE"), "namespace to watch for pipeline runs")

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.New(func(o *zap.Options) {}))

	printVersion()

	// Get a config to talk to the apiserver
	cfg := config.GetConfigOrDie()

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		Port:               port,
	})
	if err != nil {
		log.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")
	if err := pipelinev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "Unable to register pipelinev1")
		os.Exit(1)
	}

	log.Info("Creating pipelinerun-controller")
	c, err := controller.New("pipeline-controller", mgr, controller.Options{
		Reconciler: pipelinerun.NewReconciler(mgr, gitBaseURL),
	})
	if err != nil {
		log.Error(err, "Unable to create pipelinerun-controller")
		os.Exit(1)
	}

	err = c.Watch(&source.Kind{Type: &pipelinev1.PipelineRun{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Unable to set up watches for pipelinerun-controller")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager failed to start")
	}
}
