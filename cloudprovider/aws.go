package cloudprovider

import (
	"github.com/ideahitme/cluster-autoscaler/cloudprovider/aws"
	"github.com/ideahitme/cluster-autoscaler/logger"
	"github.com/ideahitme/cluster-autoscaler/pkg"
	"go.uber.org/zap"
)

// AWSProvider CloudProvider interface implementation
type AWSProvider struct {
	client *aws.Manager
}

// NewAWSCloudProvider returns aws provider
func NewAWSCloudProvider() (*AWSProvider, error) {
	client, err := aws.NewManager()
	if err != nil {
		return nil, err
	}
	return &AWSProvider{client: client}, nil
}

// NodeGroupForNode returns ASG for the given node
// nodeID is instance ID in AWS
func (ap *AWSProvider) NodeGroupForNode(nodeID pkg.NodeID) (*pkg.NodeGroup, error) {
	instanceID, err := nodeID.AWSInstanceID()
	if err != nil {
		return nil, err
	}
	logger.Log.Debug("conversion result", zap.String("@nodeid", string(nodeID)), zap.String("@instanceid", instanceID))
	ap.client.DescribeASGByInstanceID(instanceID)
	return nil, nil
}
