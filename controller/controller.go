package controller

import (
	"github.com/ideahitme/cluster-autoscaler/cloudprovider"
	"github.com/ideahitme/cluster-autoscaler/kube"
	"github.com/ideahitme/cluster-autoscaler/logger"
	"github.com/ideahitme/cluster-autoscaler/pkg"
	"go.uber.org/zap"
)

// Run main controlling mechanism
func Run(stopCh <-chan struct{}, cloudmanager cloudprovider.CloudProvider, kubemanager *kube.Manager) {
	nodelist, err := kubemanager.NodeList()
	for _, node := range nodelist.Items {
		logger.Log.Info("@node resource capacity/utilization",
			zap.String("@node_name", node.Spec.ExternalID),
			zap.String("@node_allocatable_memory", node.Status.Allocatable.Memory().String()),
			zap.String("@node_allocatable_cpu", node.Status.Allocatable.Cpu().String()),
		)
	}
	if err != nil {
		logger.Log.Fatal("error in controller run", zap.Error(err))
	}
	for _, node := range nodelist.Items {
		_, err := cloudmanager.NodeGroupForNode(pkg.NodeID(node.Spec.ProviderID))
		if err != nil {
			logger.Log.Error("error node group for node", zap.String("@nodeid", node.Spec.ProviderID), zap.Error(err))
		}
	}
	kubemanager.RunController(stopCh)
}
