package kube

import (
	"github.com/ideahitme/cluster-autoscaler/logger"
	"go.uber.org/zap"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

func isMirrorPod(pod *v1.Pod) bool {
	return true
}

func isDaemonSetPod(pod *v1.Pod) (bool, error) {
	ref, err := getRef(pod)
	if ref == nil || err != nil {
		return false, nil
	}
	return ref.Reference.Kind == "DaemonSet", nil
}

func (km *Manager) isCronJobPod(pod *v1.Pod) (bool, error) {
	ref, err := getRef(pod)
	if ref == nil || err != nil {
		return false, nil
	}
	if ref.Reference.Kind != "Job" {
		return false, nil
	}
	cronjobs, err := km.client.BatchV2alpha1().CronJobs(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
	if err != nil {
		logger.Log.Error("failed to list cronjobs", zap.Error(err))
		return true, nil
	}
	for _, job := range cronjobs.Items {
		logger.Log.Info("job uids", zap.String("@cron_uid", string(job.UID)),
			zap.String("@pod_ref_uid", string(ref.Reference.UID)))
		if job.UID == ref.Reference.UID {
			return true, nil
		}
	}
	return false, nil
}

func getRef(pod *v1.Pod) (*v1.SerializedReference, error) {
	ref, exist := pod.ObjectMeta.Annotations[v1.CreatedByAnnotation]
	if !exist {
		return nil, nil
	}
	var serRef v1.SerializedReference
	if err := runtime.DecodeInto(api.Codecs.UniversalDecoder(), []byte(ref), &serRef); err != nil {
		logger.Log.Error("failed to decode object", zap.Error(err))
		return nil, err
	}
	return &serRef, nil
}
