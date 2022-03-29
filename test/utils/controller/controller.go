/*
 * Copyright (c) 2022, Nadun De Silva. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *   http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/test/utils/cleanup"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

const (
	kustomizeDirName           = "kustomize"
	defaulControllerNamespace  = "k8s-replicator"
	defaultTestNamespacePrefix = "replicator-e2e"
	defaultImage               = "nadunrds/k8s-replicator:test"

	controllerNamespaceContextKey = "__controller_namespace__"
)

var (
	image = os.Getenv("CONTROLLER_IMAGE")
)

func GetImage() string {
	if image == "" {
		image = defaultImage
	}
	return image
}

func GetNamspace(ctx context.Context) string {
	namespace := ctx.Value(controllerNamespaceContextKey)
	return namespace.(string)
}

func SetupReplicator(ctx context.Context, t *testing.T, cfg *envconf.Config, options ...Option) context.Context {
	opts := &Options{
		labels: map[string]string{},
	}
	for _, option := range options {
		option(opts)
	}

	namespace := envconf.RandomName(defaultTestNamespacePrefix, 32)
	ctx = context.WithValue(ctx, controllerNamespaceContextKey, namespace)

	kustomizeDir, err := filepath.Abs(filepath.Join("..", "..", kustomizeDirName))
	if err != nil {
		t.Fatalf("failed to resolve kustomize dir %s: %v", kustomizeDirName, err)
	}
	t.Logf("creating controller from kustomize dir: %s", kustomizeDir)

	fileSys := filesys.MakeFsOnDisk()
	if !fileSys.Exists(kustomizeDir) {
		t.Fatalf("kustomization dir %s does not exist on file system", kustomizeDir)
	}

	k := krusty.MakeKustomizer(&krusty.Options{
		AddManagedbyLabel: true,
		PluginConfig: &types.PluginConfig{
			FnpLoadingOptions: types.FnPluginLoadingOptions{},
		},
	})
	m, err := k.Run(fileSys, kustomizeDir)
	if err != nil {
		t.Fatalf("failed build kustomization: %v", err)
	}

	var controllerDeployment *appsv1.Deployment
	for _, resource := range m.Resources() {
		yaml, err := resource.AsYAML()
		if err != nil {
			t.Fatalf("failed get kustomization output yaml: %v", err)
		}
		obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(yaml, nil, nil)
		if err != nil {
			t.Fatalf("failed parse kustomization output yaml: %v", err)
		}

		setNamespace := func(object metav1.Object, namespace string) {
			if object.GetNamespace() == defaulControllerNamespace {
				object.SetNamespace(namespace)
			}
		}

		kind := groupVersionKind.String()
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			setNamespace(deployment, namespace)
			deployment.Spec.Template.Spec.Containers[0].Image = image
			controllerDeployment = deployment
		} else if cm, ok := obj.(*corev1.ConfigMap); ok {
			setNamespace(cm, namespace)
		} else if sa, ok := obj.(*corev1.ServiceAccount); ok {
			setNamespace(sa, namespace)
		} else if ns, ok := obj.(*corev1.Namespace); ok {
			if ns.GetName() == defaulControllerNamespace {
				ns.SetName(namespace)
				for k, v := range opts.labels {
					ns.GetLabels()[k] = v
				}
			}
			ctx = cleanup.AddTestObjectToContext(ctx, t, ns)
		} else if clusterrole, ok := obj.(*rbacv1.ClusterRole); ok {
			ctx = cleanup.AddTestObjectToContext(ctx, t, clusterrole)
		} else if clusterrolebinding, ok := obj.(*rbacv1.ClusterRoleBinding); ok {
			newSubjs := []rbacv1.Subject{}
			for _, subject := range clusterrolebinding.Subjects {
				if subject.Namespace == defaulControllerNamespace {
					subject.Namespace = namespace
				}
				newSubjs = append(newSubjs, subject)
			}
			clusterrolebinding.Subjects = newSubjs
			ctx = cleanup.AddTestObjectToContext(ctx, t, clusterrolebinding)
		}
		err = cfg.Client().Resources().Create(ctx, obj.(k8s.Object))
		if err != nil {
			t.Fatalf("failed to create controller resource of kind %s: %v", kind, err)
		}
	}
	if controllerDeployment == nil {
		t.Fatalf("controller deployment not found in controller kustomize files")
	}

	err = wait.For(conditions.New(cfg.Client().Resources()).ResourceMatch(controllerDeployment, func(object k8s.Object) bool {
		d := object.(*appsv1.Deployment)
		return d.Status.AvailableReplicas > 0 && d.Status.ReadyReplicas > 0
	}), wait.WithTimeout(time.Minute))
	if err != nil {
		t.Fatalf("failed to wait for controller deployment to be ready: %v", err)
	}
	return ctx
}
