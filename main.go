package main

// TODO: improve comments and integrate with coveralls/goreport
import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"math/rand"

	"github.com/ideahitme/cluster-autoscaler/cloudprovider"
	"github.com/ideahitme/cluster-autoscaler/controller"
	"github.com/ideahitme/cluster-autoscaler/kube"
	"github.com/ideahitme/cluster-autoscaler/logger"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	stopCh := make(chan struct{}, 1)

	cfg := newConfig()
	cfg.parse()

	if err := cfg.validate(); err != nil {
		logger.Log.Fatal("invalid configuration", zap.Error(err))
	}

	config, err := clientcmd.BuildConfigFromFlags("", cfg.kubeConfig)
	if err != nil {
		logger.Log.Fatal("failed to create k8s config", zap.Error(err))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Log.Fatal("failed to create k8s client", zap.Error(err))
	}

	kubemanager := kube.NewManager(clientset)

	cloudmanager, err := cloudprovider.NewAWSCloudProvider()
	if err != nil {
		logger.Log.Fatal("failed to create cloud provider manager", zap.Error(err))
	}

	go controller.Run(stopCh, cloudmanager, kubemanager)

	handleSIGTERM(stopCh)
}

func handleSIGTERM(stopCh chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	<-signals
	close(stopCh)
}
