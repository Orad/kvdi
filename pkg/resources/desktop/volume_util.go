package desktop

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *DesktopReconciler) freePV(pv *corev1.PersistentVolume) (bool, error) {
	var changed bool
	if pv.Spec.PersistentVolumeReclaimPolicy != corev1.PersistentVolumeReclaimRetain {
		pv.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimRetain
		changed = true
	}
	if pv.Spec.ClaimRef != nil {
		pv.Spec.ClaimRef = nil
		changed = true
	}
	if changed {
		return changed, f.client.Update(context.TODO(), pv)
	}
	return changed, nil
}

func (f *DesktopReconciler) getPVCForInstance(cluster *v1alpha1.VDICluster, instance *v1alpha1.Desktop) (*corev1.PersistentVolumeClaim, error) {
	pvcNN := types.NamespacedName{
		Name:      cluster.GetUserdataVolumeName(instance.GetUser()),
		Namespace: instance.GetNamespace(),
	}
	pvc := &corev1.PersistentVolumeClaim{}
	return pvc, f.client.Get(context.TODO(), pvcNN, pvc)
}

func (f *DesktopReconciler) getPV(name string) (*corev1.PersistentVolume, error) {
	pv := &corev1.PersistentVolume{}
	return pv, f.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: metav1.NamespaceAll}, pv)
}

func (f *DesktopReconciler) getVolMapForCluster(cluster *v1alpha1.VDICluster) (*corev1.ConfigMap, error) {
	cmName := cluster.GetUserdataVolumeMapName()
	volMapCM := &corev1.ConfigMap{}
	if err := f.client.Get(context.TODO(), cmName, volMapCM); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		volMapCM = newConfigMapForCluster(cluster)
		if err := f.client.Create(context.TODO(), volMapCM); err != nil {
			return nil, err
		}
		newCM := &corev1.ConfigMap{}
		return newCM, f.client.Get(context.TODO(), cmName, newCM)
	}
	return volMapCM, nil
}
