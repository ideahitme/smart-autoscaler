package cloudprovider

import (
	"github.com/ideahitme/cluster-autoscaler/pkg"
)

// CloudProvider required interface for each cloud provider
type CloudProvider interface {
	NodeGroupForNode(nodeID pkg.NodeID) (*pkg.NodeGroup, error)
}
