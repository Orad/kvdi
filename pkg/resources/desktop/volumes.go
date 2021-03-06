package desktop

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *DesktopReconciler) reconcileVolumes(reqLogger logr.Logger, cluster *v1alpha1.VDICluster, instance *v1alpha1.Desktop) error {
	volMapCM, err := f.getVolMapForCluster(cluster)
	if err != nil {
		return err
	}
	var existingVol string
	var ok bool
	if existingVol, ok = volMapCM.Data[instance.GetUser()]; ok {
		if err := f.client.Get(context.TODO(), types.NamespacedName{Name: existingVol, Namespace: metav1.NamespaceAll}, &corev1.PersistentVolume{}); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			reqLogger.Info("The volume referenced in the userdata configmap no longer exists, creating a new one")
			existingVol = ""
		}
	}
	pvc := newPVCForUser(cluster, instance, existingVol)
	return reconcile.ReconcilePersistentVolumeClaim(reqLogger, f.client, pvc)
}

func (f *DesktopReconciler) reconcileUserdataMapping(reqLogger logr.Logger, cluster *v1alpha1.VDICluster, instance *v1alpha1.Desktop) error {

	pvc, err := f.getPVCForInstance(cluster, instance)
	if err != nil {
		return err
	}

	if pvc.Spec.VolumeName == "" {
		return errors.NewRequeueError("PVC has not had its volume provisioned yet", 3)
	}

	pvName := pvc.Spec.VolumeName

	pv, err := f.getPV(pvName)
	if err != nil {
		return err
	}

	// it won't harm the running instance and the storage class provider may
	// leave us alone
	if _, err := f.freePV(pv); err != nil {
		return err
	}

	volMapCM, err := f.getVolMapForCluster(cluster)
	if err != nil {
		return err
	}

	if volMapCM.Data == nil {
		volMapCM.Data = make(map[string]string)
	}

	if pv, ok := volMapCM.Data[instance.GetUser()]; !ok || pv != pvName {
		volMapCM.Data[instance.GetUser()] = pvName
		if err := f.client.Update(context.TODO(), volMapCM); err != nil {
			return err
		}
	}

	return nil
}

func newConfigMapForCluster(cluster *v1alpha1.VDICluster) *corev1.ConfigMap {
	nn := cluster.GetUserdataVolumeMapName()
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nn.Name,
			Namespace:       nn.Namespace,
			Labels:          cluster.GetComponentLabels("userdata-map"),
			OwnerReferences: cluster.OwnerReferences(),
		},
		Data: make(map[string]string),
	}
}

func newPVCForUser(cluster *v1alpha1.VDICluster, instance *v1alpha1.Desktop, existingPVName string) *corev1.PersistentVolumeClaim {
	spec := cluster.GetUserdataVolumeSpec()
	if existingPVName != "" {
		spec.VolumeName = existingPVName
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cluster.GetUserdataVolumeName(instance.GetUser()),
			Namespace:       instance.GetNamespace(),
			Labels:          cluster.GetUserDesktopLabels(instance.GetUser()),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: *spec,
	}
}
