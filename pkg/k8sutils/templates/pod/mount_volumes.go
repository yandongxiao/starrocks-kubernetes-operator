// Copyright 2021-present, StarRocks Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pod

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	corev1 "k8s.io/api/core/v1"
)

func MountSecrets(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	secrets []v1.SecretInfo) ([]corev1.Volume, []corev1.VolumeMount) {
	for _, secret := range secrets {
		volumes = append(volumes, corev1.Volume{
			Name: secret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      secret.Name,
			MountPath: secret.MountPath,
		})
	}
	return volumes, volumeMounts
}

func MountConfigMap(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	cmInfo v1.ConfigMapInfo, mountPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	if cmInfo.ConfigMapName != "" && cmInfo.ResolveKey != "" {
		volumes = append(volumes, corev1.Volume{
			Name: cmInfo.ConfigMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cmInfo.ConfigMapName,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      cmInfo.ConfigMapName,
			MountPath: mountPath,
		})
	}
	return volumes, volumeMounts
}
