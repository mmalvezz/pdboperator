/*
** Copyright (c) 2022 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
**
** Subject to the condition set forth below, permission is hereby granted to any
** person obtaining a copy of this software, associated documentation and/or data
** (collectively the "Software"), free of charge and under any and all copyright
** rights in the Software, and any and all patent rights owned or freely
** licensable by each licensor hereunder covering either (i) the unmodified
** Software as contributed to or provided by such licensor, or (ii) the Larger
** Works (as defined below), to deal in both
**
** (a) the Software, and
** (b) any piece of software and/or hardware listed in the lrgrwrks.txt file if
** one is included with the Software (each a "Larger Work" to which the Software
** is contributed by such licensors),
**
** without restriction, including without limitation the rights to copy, create
** derivative works of, display, perform, and distribute the Software and make,
** use, sell, offer for sale, import, export, have made, and have sold the
** Software and the Larger Work(s), and to sublicense the foregoing rights on
** either these or other terms.
**
** This license is subject to the following condition:
** The above copyright notice and either this complete permission notice or at
** a minimum a reference to the UPL must be included in all copies or
** substantial portions of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
** AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
** LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
** OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
 */

package v4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PDBSpec struct {
	PDBName                   string  `json:"pdbName,omitempty"`
	CBDsys                    CDBSys  `json:"cdbSys,omitempty"`
	CBDpwd                    CDBPass `json:"cdbPwd,omitempty"`
	TNSstrg                   string  `json:"tnsstring,omitempty"`
	Pdbuser                   PDBsys  `json:"pdbSys,omitempty"`
	Pdbpass                   PDBPass `json:"pdbPwd,omitempty"`
	FileNameConversions       string  `json:"fileNameConversions,omitempty"`
	SourceFileNameConversions string  `json:"sourceFileNameConversions,omitempty"`
	XMLFileName               string  `json:"xmlFileName,omitempty"`
	// +kubebuilder:validation:Enum=COPY;NOCOPY;MOVE
	CopyAction string `json:"copyAction,omitempty"`
	// +kubebuilder:validation:Enum=INCLUDING;KEEP
	DropAction       string `json:"dropAction,omitempty"`
	SparseClonePath  string `json:"sparseClonePath,omitempty"`
	ReuseTempFile    *bool  `json:"reuseTempFile,omitempty"`
	UnlimitedStorage *bool  `json:"unlimitedStorage,omitempty"`
	AsClone          *bool  `json:"asClone,omitempty"`
	TotalSize        string `json:"totalSize,omitempty"`
	TempSize         string `json:"tempSize,omitempty"`
	// +kubebuilder:validation:Enum=OPEN;OPEN_READ_ONLY;CLOSE;DROP;
	PDBState  string `json:"pdbState,omitempty"`
	PDBState2 int    `json:"pdbState2,omitempty"`
	Paramerer string `json:"Parameter,omitempty"`
}

// PDBsys defines the secret containing Sys Admin User mapped to key 'adminName' for PDB
type PDBsys struct {
	Secret PDBSecret `json:"secret"`
}

// PDBPass defines the secret containing Sys Admin Password mapped to key 'adminPwd' for PDB
type PDBPass struct {
	Secret PDBSecret `json:"secret"`
}

type CDBSys struct {
	Secret PDBSecret `json:"secret"`
}

type CDBPass struct {
	Secret PDBSecret `json:"secret"`
}

// PDBSecret defines the secretName
type PDBSecret struct {
	SecretName string `json:"secretName"`
	Key        string `json:"key"`
}

// PDBStatus defines the observed state of PDB
type PDBStatus struct {
	PDBStatus  string `json:"pdbStatus,omitempty"`
	ConnString string `json:"connectString,omitempty"`
	// Total size of the PDB
	PDBBitMask    int    `json:"pdbBitMask,omitempty"`
	PDBBitMaskStr string `json:"pdbBitMaskStr,omitempty"`
	TotalSize     string `json:"totalSize,omitempty"`
	// Open mode of the PDB
	OpenMode string `json:"openMode,omitempty"`
	// Modify Option of the PDB
	ModifyOption string `json:"modifyOption,omitempty"`
	// Db handles address
	Dbhandle string `json:"dbHandle,omitempty"`
	// Last Completed Action
	Dbhandle64 uint64 `json:"dbHandle64,omitempty"`
	Connected  bool   `json:"connected,omitempty"`
	Message    string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.pdbName",name="PDB NAME",type="string",description="Name of the PDB"
// +kubebuilder:printcolumn:JSONPath=".status.openMode",name="OPENMODE",type="string",description="PDB Open Mode"
// +kubebuilder:printcolumn:JSONPath=".status.totalSize",name="PDB_SIZE",type="string",description="Total Size of the PDB"
// +kubebuilder:printcolumn:JSONPath=".status.message",name="MESSAGE",type="string",description="Error message, if any"
// +kubebuilder:printcolumn:JSONPath=".status.dbHandle",name="DBHANDLE",type="string",description="rdbms handle address"
// +kubebuilder:printcolumn:JSONPath=".status.pdbBitMaskStr",name="STATE_BITMASK_STR",type="string",description="bitmask status"
// +kubebuilder:printcolumn:JSONPath=".status.connectString",name="CONNECT_STRING",type="string",description="The connect string to be used"
// +kubebuilder:resource:path=pdbs,scope=Namespaced
// +kubebuilder:storageversion

// PDB is the Schema for the pdbs API
type PDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PDBSpec   `json:"spec,omitempty"`
	Status PDBStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PDBList contains a list of PDB
type PDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PDB{}, &PDBList{})
}
