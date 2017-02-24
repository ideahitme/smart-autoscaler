package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/ideahitme/cluster-autoscaler/logger"
	"go.uber.org/zap"
)

// Manager implementation for accessing AWS API
type Manager struct {
	session   *session.Session
	asgClient *autoscaling.AutoScaling
}

// NewManager returns Manager object
func NewManager() (*Manager, error) {
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &Manager{
		session:   awsSession,
		asgClient: autoscaling.New(awsSession),
	}, nil
}

// DescribeASGByInstanceID returns ASG for a given node
func (m *Manager) DescribeASGByInstanceID(instanceID string) error {
	input := &autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}
	asg, err := m.asgClient.DescribeAutoScalingInstances(input)
	if err != nil {
		logger.Log.Error("error describing autoscaling instances", zap.Error(err))
		return err
	}
	logger.Log.Debug("asg description",
		zap.String("@aws_instance_id", instanceID),
		zap.String("@asg", *asg.AutoScalingInstances[0].AutoScalingGroupName))
	return nil
}
