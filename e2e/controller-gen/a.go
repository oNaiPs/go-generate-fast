package controller_gen

//go:generate controller-gen paths=a.go output:dir:=out crd object:headerFile=./header.txt webhook schemapatch:manifests=./manifests rbac:roleName=operator-role output:crd:artifacts:config=out/crd/artifacts output:crd:dir:=out/crd/base

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ProjectSpec struct {
	Replicas int `json:"replicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProjectSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Project `json:"items"`
}
