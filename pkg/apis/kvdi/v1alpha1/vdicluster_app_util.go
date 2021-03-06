package v1alpha1

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/version"
	corev1 "k8s.io/api/core/v1"
)

func (c *VDICluster) GetAppName() string {
	return fmt.Sprintf("%s-app", c.GetName())
}

func (c *VDICluster) GetAppReplicas() *int32 {
	if c.Spec.App != nil && c.Spec.App.Replicas != 0 {
		return &c.Spec.App.Replicas
	}
	return &defaultReplicas
}

func (c *VDICluster) GetAppResources() corev1.ResourceRequirements {
	if c.Spec.App != nil {
		return c.Spec.App.Resources
	}
	return corev1.ResourceRequirements{}
}

func (c *VDICluster) GetAppImage() string {
	if c.Spec.App != nil && c.Spec.App.Image != "" {
		return c.Spec.App.Image
	}
	return fmt.Sprintf("quay.io/tinyzimmer/kvdi:app-%s", version.Version)
}

func (c *VDICluster) GetAppPullPolicy() corev1.PullPolicy {
	return corev1.PullIfNotPresent
}

func (c *VDICluster) GetAppSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    &defaultUser,
		RunAsGroup:   &defaultUser,
		RunAsNonRoot: &trueVal,
	}
}

func (c *VDICluster) EnableCORS() bool {
	if c.Spec.App != nil {
		return c.Spec.App.CORSEnabled
	}
	return false
}

func (c *VDICluster) AuditLogEnabled() bool {
	if c.Spec.App != nil {
		return c.Spec.App.AuditLog
	}
	return false
}

func (c *VDICluster) GetAppSecretsName() string {
	if c.Spec.Secrets != nil && c.Spec.Secrets.K8SSecret != nil && c.Spec.Secrets.K8SSecret.SecretName != "" {
		return c.Spec.Secrets.K8SSecret.SecretName
	}
	return fmt.Sprintf("%s-app-secrets", c.GetName())
}
