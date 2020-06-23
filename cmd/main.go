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

	"github.com/peterbourgon/ff"
	"github.com/spf13/pflag"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/digital-plumbers-union/tekton-commit-status-tracker/pkg/controller"
	"github.com/digital-plumbers-union/tekton-commit-status-tracker/version"
)

// Change below variables to serve metrics on different host or port.
var ()
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	fs := flag.NewFlagSet("moodring", flag.ExitOnError)
	// Controller configuration
	var (
		gitBaseURL        = fs.String("git-base-url", "https://api.github.com", "base URL for git API to use")
		metricsHost       = fs.String("metrics-host", "0.0.0.0", "address to serve metrics on")
		metricsPort int32 = fs.Int("metrics-port", 8383, "port to serve metrics on")
		port              = fs.Int("port", 9443)
		namespace         = fs.String("watch-namespace", "", "namespace to watch for pipeline runs")

		// Move rest of var declared above, check rest of files for config
	)

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	fs.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix())

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
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Failed to get kubeconfig")
		os.Exit(1)
	}

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

	// replace this with a direct invocation to controller.New ?
	// c, err := controller.New("foo-controller", mgr, controller.Options{
	// 	Reconciler: &reconcileReplicaSet{client: mgr.GetClient(), log: log.WithName("reconciler")},
	// })
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "Unable to register controller")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}
