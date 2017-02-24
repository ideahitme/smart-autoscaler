package pkg

import (
	"errors"
	"regexp"
)

var (
	awsInstanceIDRegexp = regexp.MustCompile(`aws:///[^/]+/(.+)`)
)

// NodeGroup - ASG in AWS
type NodeGroup struct {
	ID string
}

// NodeID as provided by Kube API Server
type NodeID string

// AWSInstanceID converts to AWS instance ID
func (id NodeID) AWSInstanceID() (string, error) {
	matches := awsInstanceIDRegexp.FindStringSubmatch(string(id))
	if len(matches) != 2 {
		return "", errors.New("Node ID cannot be converted to AWS instance ID")
	}
	return matches[1], nil
}
