package resource

import (
	"context"
	"time"

	"github.com/tiffanyfay/aws-k8s-test-framework/test/e2e/framework/utils"

	log "github.com/cihub/seelog"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

type DaemonSetManager struct {
	cs kubernetes.Interface
}

func NewDaemonSetManager(cs kubernetes.Interface) *DaemonSetManager {
	return &DaemonSetManager{
		cs: cs,
	}
}

func (m *DaemonSetManager) WaitDaemonSetReady(ctx context.Context, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	var (
		observedDS *appsv1.DaemonSet
		err        error
	)
	start := time.Now()

	return observedDS, wait.PollImmediateUntil(utils.PollIntervalShort, func() (bool, error) {
		observedDS, err = m.cs.AppsV1().DaemonSets(ds.Namespace).Get(ds.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		log.Debugf("%d / %d pods ready in namespace '%s' in daemonset '%s' (%d seconds elapsed)",
			observedDS.Status.NumberReady, observedDS.Status.DesiredNumberScheduled, ds.Namespace,
			observedDS.ObjectMeta.Name, int(time.Since(start).Seconds()))
		if observedDS.Status.DesiredNumberScheduled == observedDS.Status.NumberReady {
			return true, nil
		}
		return false, nil
	}, ctx.Done())
}
