## [WIP] - Cluster Autoscaler for Kubernetes on AWS

## Limitations and things to be improved

- does not respect node labels as specified by pods
- does not respect critical pods
- does not consider jobs and daemonset pods
- does not respect pods with EBS volumes
- using taints - https://kubernetes.io/docs/user-guide/kubectl/kubectl_taint/, https://github.com/kubernetes/kubernetes/issues/17190
- write proper tests before it can be used anywhere

## Flows

### Scale up 
 - Identify unschedulable pods
 - Get the nodes
 - Identify autoscaling groups
 - Kill necessary number of fake pods and put the pods in its place
 - Create new node by increasing autoscaling group size
 - Deploy fake (see below) pods else where 

 ### Scale down
 - Compute theoretical possibility of rescheduling pods elsewhere from least loaded node elsewhere (consider only worker nodes)
 - If yes drain the node
 - Terminate the instance
 - Observe the pods rescheduling
 - If rescheduling fails scale back up
 - This part needs additional thinking

 ### Periods

 Scale up check is frequent and runs every minute
 Scale down is rare and runs every 30 minutes

## Fake pods

 - For now pods with hardcoded resource requests 500m cpu 100mb 
 - Randomly distributed 
 - Can be killed on demand and recreated on a new node

## TODOs
  - Add exponential back off for sorter on expanding fail

