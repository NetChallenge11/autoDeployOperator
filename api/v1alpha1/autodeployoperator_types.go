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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutoDeployOperatorSpec defines the desired state of AutoDeployOperator
type AutoDeployOperatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	ModelName string `json:"modelName"`

	Image string `json:"image"`

	NodeSelector string `json:"nodeSelector"`
}

// AutoDeployOperatorStatus defines the observed state of AutoDeployOperator
type AutoDeployOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// 배포가 성공했는지 여부를 나타내는 상태 정보
	Ready bool `json:"ready"`

	// 배포된 파드의 이름
	DeploymentName string `json:"deploymentName,omitempty"`

	// 모델이 배포된 노드의 이름
	NodeName string `json:"nodeName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AutoDeployOperator is the Schema for the autodeployoperators API
type AutoDeployOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoDeployOperatorSpec   `json:"spec,omitempty"`
	Status AutoDeployOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutoDeployOperatorList contains a list of AutoDeployOperator
type AutoDeployOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoDeployOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoDeployOperator{}, &AutoDeployOperatorList{})
}
