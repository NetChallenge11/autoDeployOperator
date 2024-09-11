/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	operatorv1alpha1 "github.com/wnguddn777/autoDeploy/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AutoDeployOperatorReconciler reconciles an AutoDeployOperator object
type AutoDeployOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is the core logic of the operator, triggered when CR is created or updated
func (r *AutoDeployOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the AutoDeployOperator instance
	autoDeploy := &operatorv1alpha1.AutoDeployOperator{}
	if err := r.Get(ctx, req.NamespacedName, autoDeploy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // client.IgnoreNotFound를 사용합니다
	}

	// Deploy models to Edge and Core servers using different kubeconfigs
	if err := r.deployModelToServer(ctx, autoDeploy.Spec.EdgeServer, autoDeploy.Spec.EdgeModel, "edge-model-deployment", autoDeploy.Spec.EdgeKubeConfig); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.deployModelToServer(ctx, autoDeploy.Spec.CoreServer, autoDeploy.Spec.CoreModel, "core-model-deployment", autoDeploy.Spec.CoreKubeConfig); err != nil {
		return ctrl.Result{}, err
	}

	// Update status after successful deployment
	autoDeploy.Status.Phase = "Deployed"
	if err := r.Status().Update(ctx, autoDeploy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // client.IgnoreNotFound를 사용합니다
	}

	return ctrl.Result{}, nil
}

// deployModelToServer creates a Deployment for the model on the specified server with kubeconfig
func (r *AutoDeployOperatorReconciler) deployModelToServer(ctx context.Context, server, model, deploymentName, kubeconfigPath string) error {
	// Create a new client using the kubeconfig for the specified server
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig for server %s: %v", server, err)
	}

	newClient, err := client.New(kubeconfig, client.Options{})
	if err != nil {
		return fmt.Errorf("failed to create client for server %s: %v", server, err)
	}

	// Define a Deployment spec
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": deploymentName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deploymentName,
							Image: model,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							// LivenessProbe 설정 (Handler로 다시 설정)
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.IntOrString{IntVal: 8080},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}

	// Check if the deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: "default"}, found)
	if err != nil && !errors.IsNotFound(err) { // client.IgnoreNotFound를 사용합니다
		return fmt.Errorf("failed to get Deployment: %v", err)
	}

	// Create or Update the Deployment
	if found.Name == "" {
		fmt.Printf("Creating deployment %s on server %s\n", deploymentName, server)
		if err := newClient.Create(ctx, deployment); err != nil {
			return fmt.Errorf("failed to create Deployment: %v", err)
		}
	} else {
		fmt.Printf("Updating deployment %s on server %s\n", deploymentName, server)
		if err := newClient.Update(ctx, deployment); err != nil {
			return fmt.Errorf("failed to update Deployment: %v", err)
		}
	}
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

// SetupWithManager sets up the controller with the Manager
func (r *AutoDeployOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AutoDeployOperator{}).
		Complete(r)
}
