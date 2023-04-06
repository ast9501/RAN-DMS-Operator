/*
Copyright 2023.

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

// S-NSSAI
type Snssai struct {
	Sst int32  `json:"sst"`
	Sd  string `json:"sd"`
}

// OAIRanSliceSpec defines the desired state of OAIRanSlice
type OAIRanSliceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// support snssai list
	SnssaiList []Snssai `json:"snssaiList"`
	// specify OAI helm chart repo
	PackageUrl string `json:"packageUrl"`
	// specify OAI helm chart version
	CuPackageVersion string `json:"cuPackageVersion"`
	DuPackageVersion string `json:"duPackageVersion"`
	// AMF addr
	AMFAddr string `json:"amfAddr"`
}

// OAIRanSliceStatus defines the observed state of OAIRanSlice
type OAIRanSliceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// State of oai ran slice
	State string `json:"state"`
	// CUCP addr (secondary-interface)
	CUCPAddr string `json:"cucpAddr"`
	// CUUP-ngu interface addr (secondary-interface)
	NGUAddr string `json:"nguAddr"`
	// CUUP addr (secondary-interface)
	CUUPAddr string `json:"cuupAddr"`
	// DU addr (secondary-interface)
	DUAddr string `json:"duAddr"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OAIRanSlice is the Schema for the oairanslice API
type OAIRanSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OAIRanSliceSpec   `json:"spec,omitempty"`
	Status OAIRanSliceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OAIRanSliceList contains a list of OAIRanSlice
type OAIRanSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OAIRanSlice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OAIRanSlice{}, &OAIRanSliceList{})
}
