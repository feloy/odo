package storage

import (
	"fmt"
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	devfileParser "github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func TestGetPVC(t *testing.T) {

	tests := []struct {
		pvc        string
		volumeName string
	}{
		{
			pvc:        "mypvc",
			volumeName: "myvolume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.volumeName, func(t *testing.T) {
			volume := getPVC(tt.volumeName, tt.pvc)

			if volume.Name != tt.volumeName {
				t.Errorf("TestGetPVC error: volume name does not match; expected %s got %s", tt.volumeName, volume.Name)
			}

			if volume.PersistentVolumeClaim.ClaimName != tt.pvc {
				t.Errorf("TestGetPVC error: pvc name does not match; expected %s got %s", tt.pvc, volume.PersistentVolumeClaim.ClaimName)
			}
		})
	}
}

func TestGetVolumeInfos(t *testing.T) {
	tests := []struct {
		name                 string
		pvcs                 []corev1.PersistentVolumeClaim
		wantOdoSourcePVCName string
		wantInfos            map[string]VolumeInfo
		wantErr              bool
	}{
		{
			name: "odo-projects is not found",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc1",
						Labels: map[string]string{
							"app.kubernetes.io/storage-name": "a-name",
						},
					},
				},
			},
			wantOdoSourcePVCName: "",
			wantInfos: map[string]VolumeInfo{
				"a-name": {
					PVCName:    "pvc1",
					VolumeName: "pvc1-vol",
				},
			},
		},
		{
			name: "odo-projects is found",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc1",
						Labels: map[string]string{
							"app.kubernetes.io/storage-name": "odo-projects",
						},
					},
				},
			},
			wantOdoSourcePVCName: "pvc1",
			wantInfos:            map[string]VolumeInfo{},
		},
		{
			name: "odo-projects is found and other pvcs",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc1",
						Labels: map[string]string{
							"app.kubernetes.io/storage-name": "odo-projects",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc2",
						Labels: map[string]string{
							"app.kubernetes.io/storage-name": "name2",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc3",
						Labels: map[string]string{
							"app.kubernetes.io/storage-name": "name3",
						},
					},
				},
			},
			wantOdoSourcePVCName: "pvc1",
			wantInfos: map[string]VolumeInfo{
				"name2": {
					PVCName:    "pvc2",
					VolumeName: "pvc2-vol",
				},
				"name3": {
					PVCName:    "pvc3",
					VolumeName: "pvc3-vol",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			odoSourcePVCName, infos, err := GetVolumeInfos(tt.pvcs)
			if err != nil != tt.wantErr {
				t.Errorf("Got error %v, expected %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.wantOdoSourcePVCName, odoSourcePVCName); diff != "" {
				t.Errorf("GetVolumeInfos() wantOdoSourcePVCName mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantInfos, infos); diff != "" {
				t.Errorf("GetVolumeInfos() wantInfos mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAddVolumeMountToContainers(t *testing.T) {

	tests := []struct {
		podName                string
		namespace              string
		serviceAccount         string
		pvc                    string
		volumeName             string
		containerMountPathsMap map[string][]string
		container              v1.Container
		labels                 map[string]string
		wantErr                bool
	}{
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			pvc:            "mypvc",
			volumeName:     "myvolume",
			containerMountPathsMap: map[string][]string{
				"container1": {"/tmp/path1", "/tmp/path2"},
			},
			container: v1.Container{
				Name:            "container1",
				Image:           "image1",
				ImagePullPolicy: v1.PullAlways,

				Command: []string{"tail"},
				Args:    []string{"-f", "/dev/null"},
				Env:     []v1.EnvVar{},
			},
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
			wantErr: false,
		},
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			pvc:            "mypvc",
			volumeName:     "myvolume",
			containerMountPathsMap: map[string][]string{
				"container1": {"/tmp/path1", "/tmp/path2"},
			},
			container: v1.Container{},
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.podName, func(t *testing.T) {
			containers := []v1.Container{tt.container}
			initContainers := []v1.Container{}
			addVolumeMountToContainers(containers, initContainers, tt.volumeName, tt.containerMountPathsMap)

			mountPathCount := 0
			for _, container := range containers {
				if container.Name == tt.container.Name {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == tt.volumeName {
							for _, mountPath := range tt.containerMountPathsMap[tt.container.Name] {
								if volumeMount.MountPath == mountPath {
									mountPathCount++
								}
							}
						}
					}
				}
			}

			if mountPathCount != len(tt.containerMountPathsMap[tt.container.Name]) {
				t.Errorf("Volume Mounts for %s have not been properly mounted to the container", tt.volumeName)
			}
		})
	}
}

func TestGetVolumesAndVolumeMounts(t *testing.T) {

	type testVolumeMountInfo struct {
		mountPath  string
		volumeName string
	}

	tests := []struct {
		name                string
		components          []devfilev1.Component
		volumeNameToVolInfo map[string]VolumeInfo
		wantContainerToVol  map[string][]testVolumeMountInfo
		wantErr             bool
	}{
		{
			name:       "One volume mounted",
			components: []devfilev1.Component{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("comp2")},
			volumeNameToVolInfo: map[string]VolumeInfo{
				"myvolume1": {
					PVCName:    "volume1-pvc",
					VolumeName: "volume1-pvc-vol",
				},
			},
			wantContainerToVol: map[string][]testVolumeMountInfo{
				"comp1": {
					{
						mountPath:  "/my/volume/mount/path1",
						volumeName: "volume1-pvc-vol",
					},
				},
				"comp2": {
					{
						mountPath:  "/my/volume/mount/path1",
						volumeName: "volume1-pvc-vol",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "One volume mounted at diff locations",
			components: []devfilev1.Component{
				{
					Name: "container1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "volume1",
										Path: "/path1",
									},
									{
										Name: "volume1",
										Path: "/path2",
									},
								},
							},
						},
					},
				},
			},
			volumeNameToVolInfo: map[string]VolumeInfo{
				"volume1": {
					PVCName:    "volume1-pvc",
					VolumeName: "volume1-pvc-vol",
				},
			},
			wantContainerToVol: map[string][]testVolumeMountInfo{
				"container1": {
					{
						mountPath:  "/path1",
						volumeName: "volume1-pvc-vol",
					},
					{
						mountPath:  "/path2",
						volumeName: "volume1-pvc-vol",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "One volume mounted at diff container components",
			components: []devfilev1.Component{
				{
					Name: "container1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "volume1",
										Path: "/path1",
									},
								},
							},
						},
					},
				},
				{
					Name: "container2",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "volume1",
										Path: "/path2",
									},
								},
							},
						},
					},
				},
			},
			volumeNameToVolInfo: map[string]VolumeInfo{
				"volume1": {
					PVCName:    "volume1-pvc",
					VolumeName: "volume1-pvc-vol",
				},
			},
			wantContainerToVol: map[string][]testVolumeMountInfo{
				"container1": {
					{
						mountPath:  "/path1",
						volumeName: "volume1-pvc-vol",
					},
				},
				"container2": {
					{
						mountPath:  "/path2",
						volumeName: "volume1-pvc-vol",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid case",
			components: []devfilev1.Component{
				{
					Name: "container1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
					}),
					ComponentUnion: devfilev1.ComponentUnion{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(tt.components)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			podTemplateSpec, err := generator.GetPodTemplateSpec(devObj, generator.PodTemplateParams{})
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetVolumesAndVolumeMounts error - %v", err)
				return
			}

			if tt.wantErr && err != nil {
				return
			}

			containers := podTemplateSpec.Spec.Containers

			var options parsercommon.DevfileOptions
			if tt.wantErr {
				options = parsercommon.DevfileOptions{
					Filter: map[string]interface{}{
						"firstString": "firstStringValue",
					},
				}
			}

			initContainers := []v1.Container{}
			pvcVols, err := GetPersistentVolumesAndVolumeMounts(devObj, containers, initContainers, tt.volumeNameToVolInfo, options)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetVolumesAndVolumeMounts unexpected error: %v", err)
				return
			} else if tt.wantErr && err != nil {
				return
			} else if tt.wantErr && err == nil {
				t.Error("TestGetVolumesAndVolumeMounts expected error but got nil")
				return
			}

			// check if the pvc volumes returned are correct
			for _, volInfo := range tt.volumeNameToVolInfo {
				matched := false
				for _, pvcVol := range pvcVols {
					if volInfo.VolumeName == pvcVol.Name && pvcVol.PersistentVolumeClaim != nil && volInfo.PVCName == pvcVol.PersistentVolumeClaim.ClaimName {
						matched = true
					}
				}

				if !matched {
					t.Errorf("TestGetVolumesAndVolumeMounts error - could not find volume details %s in the actual result", volInfo.VolumeName)
				}
			}

			// check the volume mounts of the containers
			for _, container := range containers {
				if volMounts, ok := tt.wantContainerToVol[container.Name]; !ok {
					t.Errorf("TestGetVolumesAndVolumeMounts error - did not find the expected container %s", container.Name)
					return
				} else {
					for _, expectedVolMount := range volMounts {
						matched := false
						for _, actualVolMount := range container.VolumeMounts {
							if expectedVolMount.volumeName == actualVolMount.Name && expectedVolMount.mountPath == actualVolMount.MountPath {
								matched = true
							}
						}

						if !matched {
							t.Errorf("TestGetVolumesAndVolumeMounts error - could not find volume mount details for path %s in the actual result for container %s", expectedVolMount.mountPath, container.Name)
						}
					}
				}
			}
		})
	}
}

func TestGetAutomountVolumes(t *testing.T) {

	container1 := corev1.Container{
		Name:  "container1",
		Image: "image1",
	}
	container2 := corev1.Container{
		Name:  "container2",
		Image: "image2",
	}
	initContainer1 := corev1.Container{
		Name:  "initContainer1",
		Image: "image1",
	}
	initContainer2 := corev1.Container{
		Name:  "initContainer2",
		Image: "image2",
	}

	type args struct {
		configAutomountClient func(ctrl *gomock.Controller) configAutomount.Client
		containers            []corev1.Container
		initContainers        []corev1.Container
	}
	tests := []struct {
		name             string
		args             args
		want             []corev1.Volume
		wantVolumeMounts []corev1.VolumeMount
		wantEnvFroms     []corev1.EnvFromSource
		wantErr          bool
	}{
		{
			name: "No automounting volume",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want:             nil,
			wantVolumeMounts: nil,
			wantErr:          false,
		},
		{
			name: "One PVC",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					info1 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypePVC,
						VolumeName: "pvc1",
						MountPath:  "/path/to/mount1",
						MountAs:    configAutomount.MountAsFile,
					}
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{info1}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want: []v1.Volume{
				{
					Name: "auto-pvc-pvc1",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc1",
						},
					},
				},
			},
			wantVolumeMounts: []v1.VolumeMount{
				{
					Name:      "auto-pvc-pvc1",
					MountPath: "/path/to/mount1",
				},
			},
			wantErr: false,
		},
		{
			name: "One PVC and one secret",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					info1 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypePVC,
						VolumeName: "pvc1",
						MountPath:  "/path/to/mount1",
						MountAs:    configAutomount.MountAsFile,
					}
					info2 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypeSecret,
						VolumeName: "secret2",
						MountPath:  "/path/to/mount2",
						MountAs:    configAutomount.MountAsFile,
					}
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{info1, info2}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want: []v1.Volume{
				{
					Name: "auto-pvc-pvc1",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc1",
						},
					},
				},
				{
					Name: "auto-secret-secret2",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "secret2",
						},
					},
				},
			},
			wantVolumeMounts: []v1.VolumeMount{
				{
					Name:      "auto-pvc-pvc1",
					MountPath: "/path/to/mount1",
				},
				{
					Name:      "auto-secret-secret2",
					MountPath: "/path/to/mount2",
				},
			},
			wantErr: false,
		},
		{
			name: "One PVC, one secret and one configmap",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					info1 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypePVC,
						VolumeName: "pvc1",
						MountPath:  "/path/to/mount1",
						MountAs:    configAutomount.MountAsFile,
					}
					info2 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypeSecret,
						VolumeName: "secret2",
						MountPath:  "/path/to/mount2",
						MountAs:    configAutomount.MountAsFile,
					}
					info3 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypeConfigmap,
						VolumeName: "cm3",
						MountPath:  "/path/to/mount3",
						MountAs:    configAutomount.MountAsFile,
					}
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{info1, info2, info3}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want: []v1.Volume{
				{
					Name: "auto-pvc-pvc1",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc1",
						},
					},
				},
				{
					Name: "auto-secret-secret2",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "secret2",
						},
					},
				},
				{
					Name: "auto-cm-cm3",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm3",
							},
						},
					},
				},
			},
			wantVolumeMounts: []v1.VolumeMount{
				{
					Name:      "auto-pvc-pvc1",
					MountPath: "/path/to/mount1",
				},
				{
					Name:      "auto-secret-secret2",
					MountPath: "/path/to/mount2",
				},
				{
					Name:      "auto-cm-cm3",
					MountPath: "/path/to/mount3",
				},
			},
			wantErr: false,
		},
		{
			name: "One secret and one configmap mounted as Env",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					info1 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypeSecret,
						VolumeName: "secret1",
						MountAs:    configAutomount.MountAsEnv,
					}
					info2 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypeConfigmap,
						VolumeName: "cm2",
						MountAs:    configAutomount.MountAsEnv,
					}
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{info1, info2}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want:             nil,
			wantVolumeMounts: nil,
			wantEnvFroms: []corev1.EnvFromSource{
				{
					SecretRef: &v1.SecretEnvSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret1",
						},
					},
				},
				{
					ConfigMapRef: &v1.ConfigMapEnvSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm2",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "One secret and one configmap mounted as Subpath",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					info1 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypeSecret,
						VolumeName: "secret1",
						MountPath:  "/path/to/secret1",
						MountAs:    configAutomount.MountAsSubpath,
						Keys:       []string{"secretKey1", "secretKey2"},
					}
					info2 := configAutomount.AutomountInfo{
						VolumeType: configAutomount.VolumeTypeConfigmap,
						VolumeName: "cm2",
						MountPath:  "/path/to/cm2",
						MountAs:    configAutomount.MountAsSubpath,
						Keys:       []string{"cmKey1", "cmKey2"},
					}
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{info1, info2}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want: []corev1.Volume{
				{
					Name: "auto-secret-secret1",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "secret1",
						},
					},
				},
				{
					Name: "auto-cm-cm2",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm2",
							},
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "auto-secret-secret1",
					MountPath: "/path/to/secret1/secretKey1",
					SubPath:   "secretKey1",
				},
				{
					Name:      "auto-secret-secret1",
					MountPath: "/path/to/secret1/secretKey2",
					SubPath:   "secretKey2",
				},
				{
					Name:      "auto-cm-cm2",
					MountPath: "/path/to/cm2/cmKey1",
					SubPath:   "cmKey1",
				},
				{
					Name:      "auto-cm-cm2",
					MountPath: "/path/to/cm2/cmKey2",
					SubPath:   "cmKey2",
				},
			},
			wantErr: false,
		},
		{
			name: "One secret and one configmap mounted as file with access mode",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					info1 := configAutomount.AutomountInfo{
						VolumeType:      configAutomount.VolumeTypeSecret,
						VolumeName:      "secret1",
						MountPath:       "/path/to/secret1",
						MountAs:         configAutomount.MountAsFile,
						Keys:            []string{"secretKey1", "secretKey2"},
						MountAccessMode: pointer.Int32(0400),
					}
					info2 := configAutomount.AutomountInfo{
						VolumeType:      configAutomount.VolumeTypeConfigmap,
						VolumeName:      "cm2",
						MountPath:       "/path/to/cm2",
						MountAs:         configAutomount.MountAsFile,
						Keys:            []string{"cmKey1", "cmKey2"},
						MountAccessMode: pointer.Int32(0444),
					}
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{info1, info2}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want: []corev1.Volume{
				{
					Name: "auto-secret-secret1",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							DefaultMode: pointer.Int32(0400),
							SecretName:  "secret1",
						},
					},
				},
				{
					Name: "auto-cm-cm2",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							DefaultMode: pointer.Int32(0444),
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm2",
							},
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "auto-secret-secret1",
					MountPath: "/path/to/secret1",
				},
				{
					Name:      "auto-cm-cm2",
					MountPath: "/path/to/cm2",
				},
			},
			wantErr: false,
		},
		{
			name: "One secret and one configmap mounted as Subpath with access mode",
			args: args{
				configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
					info1 := configAutomount.AutomountInfo{
						VolumeType:      configAutomount.VolumeTypeSecret,
						VolumeName:      "secret1",
						MountPath:       "/path/to/secret1",
						MountAs:         configAutomount.MountAsSubpath,
						Keys:            []string{"secretKey1", "secretKey2"},
						MountAccessMode: pointer.Int32(0400),
					}
					info2 := configAutomount.AutomountInfo{
						VolumeType:      configAutomount.VolumeTypeConfigmap,
						VolumeName:      "cm2",
						MountPath:       "/path/to/cm2",
						MountAs:         configAutomount.MountAsSubpath,
						Keys:            []string{"cmKey1", "cmKey2"},
						MountAccessMode: pointer.Int32(0444),
					}
					client := configAutomount.NewMockClient(ctrl)
					client.EXPECT().GetAutomountingVolumes().Return([]configAutomount.AutomountInfo{info1, info2}, nil)
					return client
				},
				containers:     []corev1.Container{container1, container2},
				initContainers: []corev1.Container{initContainer1, initContainer2},
			},
			want: []corev1.Volume{
				{
					Name: "auto-secret-secret1",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName:  "secret1",
							DefaultMode: pointer.Int32(0400),
						},
					},
				},
				{
					Name: "auto-cm-cm2",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm2",
							},
							DefaultMode: pointer.Int32(0444),
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "auto-secret-secret1",
					MountPath: "/path/to/secret1/secretKey1",
					SubPath:   "secretKey1",
				},
				{
					Name:      "auto-secret-secret1",
					MountPath: "/path/to/secret1/secretKey2",
					SubPath:   "secretKey2",
				},
				{
					Name:      "auto-cm-cm2",
					MountPath: "/path/to/cm2/cmKey1",
					SubPath:   "cmKey1",
				},
				{
					Name:      "auto-cm-cm2",
					MountPath: "/path/to/cm2/cmKey2",
					SubPath:   "cmKey2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			got, err := GetAutomountVolumes(tt.args.configAutomountClient(ctrl), tt.args.containers, tt.args.initContainers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAutomountVolumes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetAutomountVolumes() mismatch (-want +got):\n%s", diff)
			}

			checkContainers := func(containers, initContainers []corev1.Container) error {
				allContainers := containers
				allContainers = append(allContainers, initContainers...)
				for _, container := range allContainers {
					if diff := cmp.Diff(tt.wantVolumeMounts, container.VolumeMounts); diff != "" {
						return fmt.Errorf(diff)
					}
					if diff := cmp.Diff(tt.wantEnvFroms, container.EnvFrom); diff != "" {
						return fmt.Errorf(diff)
					}
				}
				return nil
			}

			if err := checkContainers(tt.args.containers, tt.args.initContainers); err != nil {
				t.Errorf("GetAutomountVolumes() containers error: %v", err)
			}
		})
	}
}
