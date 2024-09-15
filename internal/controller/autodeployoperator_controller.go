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

	"github.com/go-logr/logr"
	"github.com/wnguddn777/autoDeploy/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AutoDeployOperatorReconciler reconciles a AutoDeployOperator object
type AutoDeployOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile handles the deployment of the ML model to worker nodes
func (r *AutoDeployOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// log 선언은 함수 내부에서 이루어져야 함
	log := ctrl.Log.WithValues("autodeployoperator", req.NamespacedName)

	// Fetch the AutoDeployOperator instance
	var autoDeployOperator v1alpha1.AutoDeployOperator
	if err := r.Get(ctx, req.NamespacedName, &autoDeployOperator); err != nil {
		log.Error(err, "unable to fetch AutoDeployOperator")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Define the deployment for the ML model
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      autoDeployOperator.Name,
			Namespace: req.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": autoDeployOperator.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": autoDeployOperator.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mlmodel-container",
							Image: autoDeployOperator.Spec.Image, // 수정된 필드
						},
					},
					NodeSelector: map[string]string{"kubernetes.io/hostname": autoDeployOperator.Spec.NodeSelector}, // 수정된 필드
				},
			},
		},
	}

	// Set owner reference to the AutoDeployOperator instance for garbage collection
	if err := ctrl.SetControllerReference(&autoDeployOperator, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Deploy the model to the specified worker node
	if err := r.Create(ctx, deployment); err != nil {
		// client.IgnoreAlreadyExists(err) 함수는 AlreadyExists 에러를 무시함
		if err := client.IgnoreAlreadyExists(err); err != nil {
			return ctrl.Result{}, err
		}
		r.Log.Info("Deployment already exists, updating...", "Deployment", deployment.Name)
		if err := r.Update(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoDeployOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AutoDeployOperator{}). // 수정된 리소스 경로
		Complete(r)
}
