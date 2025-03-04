// This file is part of MinIO Direct CSI
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package installer

import (
	"context"
	"testing"

	"github.com/minio/direct-csi/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

func init() {
	client.FakeInit()
}

func TestInstaller(t1 *testing.T) {
	installConfig := &Config{
		Identity:                   "direct-csi-min-io",
		DirectCSIContainerImage:    "test-image",
		DirectCSIContainerOrg:      "test-org",
		DirectCSIContainerRegistry: "test-registry",
		AdmissionControl:           false,
		LoopbackMode:               false,
		NodeSelector:               nil,
		Tolerations:                nil,
		SeccompProfile:             "",
		ApparmorProfile:            "",
		DynamicDriveDiscovery:      true,
		DryRun:                     false,
	}

	getDiscoveryGroupsAndMethodsFn := func() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
		return []*metav1.APIGroup{
				{
					Name: "policy",
					Versions: []metav1.GroupVersionForDiscovery{
						{
							GroupVersion: "policy/v1beta1",
							Version:      "v1beta1",
						},
					},
				},
				{
					Name: "storage.k8s.io",
					Versions: []metav1.GroupVersionForDiscovery{
						{
							GroupVersion: "storage.k8s.io/v1",
							Version:      "v1",
						},
					},
				},
			}, []*metav1.APIResourceList{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "policy/v1beta1",
						Kind:       "PodSecurityPolicy",
					},
					GroupVersion: "policy/v1beta1",
					APIResources: []metav1.APIResource{
						{
							Name:       "policy",
							Group:      "policy",
							Version:    "v1beta1",
							Namespaced: false,
							Kind:       "PodSecurityPolicy",
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "storage.k8s.io/v1",
						Kind:       "CSIDriver",
					},
					GroupVersion: "storage.k8s.io/v1",
					APIResources: []metav1.APIResource{
						{
							Name:       "CSIDriver",
							Group:      "storage.k8s.io",
							Version:    "v1",
							Namespaced: false,
							Kind:       "CSIDriver",
						},
					},
				},
			}, nil
	}

	testVersions := []version.Info{
		{
			Major: "1",
			Minor: "18",
		},
		{
			Major: "1",
			Minor: "19",
		},
		{
			Major: "1",
			Minor: "20",
		},
		{
			Major: "1",
			Minor: "21",
		},
		{
			Major: "1",
			Minor: "22",
		},
	}

	for _, testVersion := range testVersions {
		client.SetFakeDiscoveryClient(getDiscoveryGroupsAndMethodsFn, &testVersion)
		ctx := context.TODO()
		if err := Install(ctx, installConfig); err != nil {
			t1.Fatalf("install failed: %v", err)
		}
		installConfig.ForceRemove = true
		installConfig.UninstallCRD = true
		if err := Uninstall(ctx, installConfig); err != nil {
			t1.Fatalf("uninstall failed: %v", err)
		}
		if _, err := client.GetKubeClient().CoreV1().Namespaces().Get(ctx, "direct-csi-min-io", metav1.GetOptions{}); err == nil {
			t1.Errorf("namespace not removed upon uninstallation. version: %s.%s", testVersion.Major, testVersion.Minor)
		}
	}
}
