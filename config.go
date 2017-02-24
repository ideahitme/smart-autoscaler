package main

import (
	"fmt"

	"os"

	"github.com/ideahitme/cluster-autoscaler/logger"
	flags "github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

type config struct {
	InCluster  bool `long:"in-cluster" description:"set if running outside the cluster"`
	kubeConfig string
}

func newConfig() *config {
	return &config{}
}

func (cfg *config) parse() {
	parser := flags.NewParser(cfg, flags.Default)
	if _, err := parser.Parse(); err != nil {
		logger.Log.Fatal("incorrect configuration", zap.Error(err))
	}
	if !cfg.InCluster {
		cfg.kubeConfig = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	}
}

func (cfg *config) validate() error {
	return nil
}

func (cfg *config) String() string {
	return "In cluster: " + fmt.Sprintf("%t", cfg.InCluster)
}
