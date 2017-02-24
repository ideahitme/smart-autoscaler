.PHONY: run

AWS_REGION?=eu-central-1

build:
	go build -o cluster-autoscaler

run: build
	zaws login teapot PowerUser
	zkubectl login kube-aws-test-1.teapot.zalan.do
	AWS_REGION=${AWS_REGION} ./cluster-autoscaler