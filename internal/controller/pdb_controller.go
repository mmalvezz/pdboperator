/*
Copyright 2025.

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

/*
#cgo CFLAGS: -I./../../common/ -I/usr/include/oracle/21/client64/  -I./ -I./include -DDEVPHASE -I./adepublic -Wimplicit-function-declaration
#cgo LDFLAGS:   -L/usr/lib64  -Wl,--warn-once  -L/usr/lib/oracle/21/client64/lib/ -lclntsh  -O2
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <dlfcn.h>
#include <string.h>
#include <oci.h>
#include <pdbtypes.h>
#include <pdbfunctions.c>

*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unsafe"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	databasev4 "github.com/oracle/pdboperator/api/v4"

	. "github.com/oracle/pdboperator/common/pdbutil"
)

const (
	recomp = 655322
)

const PDBFinalizer = "database.oracle.com/PDBFinalizer"

const (
	SQLEFE = 0
	PDBDET = 1
	PDBSPC = 2 /* Used by get /pdbname/status and get /pdbname */
	PDBVIL = 3 /* Plug in database violation */
	SQLERR = 4 /* just tyo test sql handle error */
	SQLPAR = 5
)

// PDBReconciler reconciles a PDB object
type PDBReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Interval time.Duration
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=database.oracle.com,resources=pdbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.oracle.com,resources=pdbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.oracle.com,resources=pdbs/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets;services;configmaps;namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.oracle.com,resources=secrets;events;deployments,verbs=get;list;watch;create;update;patch;delete

var cnt int

/*   STATE TABLE: is defined in util.go file

     PDBCRT = 0x00000001 - Create pdb
     PDBOPN = 0x00000002 - Open pdb read write
     PDBCLS = 0x00000004 - Close pdb
     PDBDIC = 0x00000008 - Drop pdb include datafiles
     OCIHDL = 0x00000010 - OCI handle allocation
     OCICON = 0x00000020 - Rdbms connection
     FNALAZ = 0x00000040 - Finalizer configured
     +---------- Error section -----------------+
     PDBCRE = 0x00001000 - PDB creation error
     PDBOPE = 0x00002000 - PDB open error
     PDBCLE = 0x00004000 - PDB close error
     OCIHDE = 0x00008000 - Allocation Handle Error
     OCICOE = 0x00010000 - CDD connection Error
     FNALAE = 0x00020000 - Finalizer error
*/

func (r *PDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues("RECONCILIATION_LOOP", req.NamespacedName)
	var err error

	pdb := &databasev4.PDB{}

	reconcilePeriod := r.Interval * time.Second
	requeueY := ctrl.Result{Requeue: true, RequeueAfter: reconcilePeriod}
	requeueN := ctrl.Result{}

	err = r.Client.Get(context.TODO(), req.NamespacedName, pdb)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("PDB resource not found", "Pdb", pdb.Spec.PDBName)
			return requeueN, nil
		}
		log.Info("Client.Get Error")
		return requeueN, err
	}

	//* ORACLE SESSION *//
	if Bit(pdb.Status.PDBBitMask, OCIHDL|OCIHDE|OCICOE|PDBDIC) == false {
		err = r.RDBMSHandles(ctx, req, pdb)
		if err != nil {
			log.Error(err, err.Error())
			pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, OCIHDE)
			r.UpdateStatus(ctx, pdb)
		}
		pdb.Status.PDBBitMaskStr = Bitmaskprint(pdb.Status.PDBBitMask)
		r.UpdateStatus(ctx, pdb)
	}

	//* CREATE *//
	if Bit(pdb.Status.PDBBitMask, OCICON) == true && Bit(pdb.Status.PDBBitMask, PDBCRT|PDBCRE) == false {
		log.Info("pdb creation")
		err = r.CreatePDB(ctx, req, pdb)
		if err != nil {
			log.Error(err, err.Error())
		}
		pdb.Status.PDBBitMaskStr = Bitmaskprint(pdb.Status.PDBBitMask)
		r.UpdateStatus(ctx, pdb)
	}

	//* SET FINALIZER *//
	if Bit(pdb.Status.PDBBitMask, FNALAZ|FNALAE) == false && Bit(pdb.Status.PDBBitMask, PDBCRE|OCIHDL) == true && Bit(pdb.Status.PDBBitMask, OCICOE) == false {
		err = r.SetFinalizer(ctx, req, pdb)
		if err != nil {
			log.Error(err, err.Error())
			pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, FNALAE)
			r.UpdateStatus(ctx, pdb)
		}
		//pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, FNALAZ)
		//pdb.Status.PDBBitMaskStr = Bitmaskprint(pdb.Status.PDBBitMask)
	}

	//* CREATE *//
	if Bit(pdb.Status.PDBBitMask, OCICON) == true && Bit(pdb.Status.PDBBitMask, PDBCRT|PDBCRE) == false {
		log.Info("pdb creation")
		err = r.CreatePDB(ctx, req, pdb)
		if err != nil {
			log.Error(err, err.Error())
		}
		pdb.Status.PDBBitMaskStr = Bitmaskprint(pdb.Status.PDBBitMask)
		r.UpdateStatus(ctx, pdb)
	}

	//* OPEN *//
	if pdb.Spec.PDBState == "OPEN" && Bit(pdb.Status.PDBBitMask, OCICON) == true && Bit(pdb.Status.PDBBitMask, PDBOPN|PDBOPE) == false {
		log.Info("pdb opening")
		err = r.OpenPDB(ctx, req, pdb)
		if err != nil {
			log.Error(err, err.Error())
		}
		pdb.Status.PDBBitMaskStr = Bitmaskprint(pdb.Status.PDBBitMask)
		r.UpdateStatus(ctx, pdb)
	}

	//* CLOSE *//
	if pdb.Spec.PDBState == "CLOSE" && Bit(pdb.Status.PDBBitMask, OCICON) == true && Bit(pdb.Status.PDBBitMask, PDBCLS|PDBCLE) == false {
		log.Info("pdb closing")
		err = r.ClosePDB(ctx, req, pdb)
		if err != nil {
			log.Error(err, err.Error())
		}
		pdb.Status.PDBBitMaskStr = Bitmaskprint(pdb.Status.PDBBitMask)
		r.UpdateStatus(ctx, pdb)
	}

	//* DELETE  *//
	if pdb.ObjectMeta.DeletionTimestamp.IsZero() == false && Bit(pdb.Status.PDBBitMask, PDBCRT) == true && Bit(pdb.Status.PDBBitMask, PDBDIC) == false {
		log.Info(" ObjectMeta.DeletionTimestamp.IsZero is not null")
		if controllerutil.ContainsFinalizer(pdb, PDBFinalizer) {
			if Bit(pdb.Status.PDBBitMask, OCIHDL) == true {
				err = r.DropPDB(ctx, req, pdb)
				if err != nil {
					log.Info("Cannot drop database")
					r.UpdateStatus(ctx, pdb)
					return requeueN, err
				}
			}
			r.FreePdbHandle(ctx, req, pdb)
			pdb.Status.PDBBitMask = Bid(pdb.Status.PDBBitMask, OCIHDL)
			r.UpdateStatus(ctx, pdb)
			controllerutil.RemoveFinalizer(pdb, PDBFinalizer)
			if err := r.Update(ctx, pdb); err != nil {
				return requeueN, err
			}

		}
	}

	//* MONITOR  RESOURCE STATUS *//

	log.Info("STATEBITMASK:" + pdb.Status.PDBBitMaskStr)
	return requeueY, nil

}

func (r *PDBReconciler) UpdateStatus(ctx context.Context, pdb *databasev4.PDB) {
	log := logf.FromContext(ctx).WithValues("Update Status", "=====")
	err := r.Status().Update(ctx, pdb)
	if err != nil {
		fmt.Printf("[1]Error updating status\n")
		log.Error(err, err.Error())
	}
}

func (r *PDBReconciler) getSecret(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB, secretName string, keyName string) (string, error) {

	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: pdb.Namespace}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			pdb.Status.Message = "Secret not found:" + secretName
			return "", err
		}
		return "", err
	}

	return string(secret.Data[keyName]), nil
}

func (r *PDBReconciler) RDBMSHandles(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB) error {
	log := r.Log.WithValues("RDBMSHandles", req.NamespacedName)
	//var rc C.uchar
	var CredPtr (*C.struct_OCICred) /* Database credential */
	var ConnErr (*C.struct_OCIErrmsg)
	var dbhandle (*C.struct_OCIHandlePool)

	CredPtr = C.MallocOCIcred()
	tnsal := pdb.Spec.TNSstrg
	uname, _ := r.getSecret(ctx, req, pdb, pdb.Spec.CBDsys.Secret.SecretName, pdb.Spec.CBDsys.Secret.Key)
	passw, _ := r.getSecret(ctx, req, pdb, pdb.Spec.CBDpwd.Secret.SecretName, pdb.Spec.CBDpwd.Secret.Key)
	uname = strings.TrimSpace(uname)
	passw = strings.TrimSpace(passw)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(CredPtr.uname)), len(uname)), uname)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(CredPtr.passw)), len(passw)), passw)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(CredPtr.tnsal)), len(tnsal)), tnsal)
	dbhandle = C.OCIConnectRDBMS(CredPtr.uname, CredPtr.passw, CredPtr.tnsal)
	ConnErr = dbhandle.errmsg /* Get the error */
	C.FreeOCIcred(CredPtr)
	fmt.Printf("+----------------------------------------------------------------------------+\n")
	fmt.Printf("SIZE:ADDR 0x%x 0x%x \n", unsafe.Sizeof(dbhandle), unsafe.Pointer(dbhandle))
	fmt.Printf("TNS:%s\n", tnsal)
	fmt.Printf("USR:%s\n", uname)
	fmt.Printf(" sqlca.sqlcode=%d\n", ConnErr.errcode)
	fmt.Printf(" sqlca.errmsg=%s\n", ConnErr.errbuf)
	fmt.Printf("+----------------------------------------------------------------------------+\n")
	pdb.Status.Dbhandle = fmt.Sprintf("0x%X", unsafe.Pointer(dbhandle)) /* Copy DB Handles address into status struct */
	fmt.Sscanf(pdb.Status.Dbhandle, "0x%X", &pdb.Status.Dbhandle64)
	pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, OCIHDL)

	if ConnErr.errcode == 0 {
		log.Info("Connection completed")
		fmt.Printf("%s:%d\n", pdb.Status.Dbhandle, pdb.Status.Dbhandle64)
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, OCICON)
		r.UpdateStatus(ctx, pdb)
		return nil
	} else {
		pdb.Status.Message = fmt.Sprintf("ORA-%d", ConnErr.errcode)
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, OCICOE)
		r.UpdateStatus(ctx, pdb)
		return errors.New(pdb.Status.Message)
	}

}

func (r *PDBReconciler) ClosePDB(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB) error {
	log := logf.FromContext(ctx).WithValues("ClosePDB", req.NamespacedName)
	var p1 C.struct_pdbinfo
	var CloseErr (*C.struct_OCIErrmsg)
	var dbhandle (*C.struct_OCIHandlePool)
	dbhandle = C.CastPtr(C.ulong(pdb.Status.Dbhandle64)) /* get handle address */
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.pdbname)), len(pdb.Spec.PDBName)), pdb.Spec.PDBName)
	C.OCIClosePdb(dbhandle, SQLEFE, (*C.pdbinfo)(&p1))
	CloseErr = dbhandle.errmsg

	log.Info("SQLCA.SQLCODE:" + fmt.Sprintf("%d", CloseErr.errcode))
	log.Info("SQLCA.ERRMSG:" + fmt.Sprintf("%s", CloseErr.errbuf))

	if CloseErr.errcode != 0 {
		pdb.Status.Message = fmt.Sprintf("ORA-%d", CloseErr.errcode)
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBCLE)
		r.UpdateStatus(ctx, pdb)
		return errors.New(pdb.Status.Message)
	} else {
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBCLS)
		if Bit(pdb.Status.PDBBitMask, PDBOPN) == true {
			pdb.Status.PDBBitMask = Bid(pdb.Status.PDBBitMask, PDBOPN)
		}

		pdb.Status.Message = "CLOSE:OK"
		pdb.Status.OpenMode = "MOUNT"
		pdb.Status.PDBStatus = "CLOSE"
		r.UpdateStatus(ctx, pdb)
	}

	return nil

}

func (r *PDBReconciler) OpenPDB(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB) error {
	log := logf.FromContext(ctx).WithValues("OpenPDB", req.NamespacedName)
	var p1 C.struct_pdbinfo
	var OpenErr (*C.struct_OCIErrmsg)
	var dbhandle (*C.struct_OCIHandlePool)
	dbhandle = C.CastPtr(C.ulong(pdb.Status.Dbhandle64)) /* get handle address */
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.pdbname)), len(pdb.Spec.PDBName)), pdb.Spec.PDBName)
	C.OCIOpenPdb(dbhandle, SQLEFE, (*C.pdbinfo)(&p1))
	OpenErr = dbhandle.errmsg

	log.Info("SQLCA.SQLCODE:" + fmt.Sprintf("%d", OpenErr.errcode))
	log.Info("SQLCA.ERRMSG:" + fmt.Sprintf("%s", OpenErr.errbuf))

	if OpenErr.errcode != 0 {
		pdb.Status.Message = fmt.Sprintf("ORA-%d", OpenErr.errcode)
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBOPE)
		r.UpdateStatus(ctx, pdb)
		return errors.New(pdb.Status.Message)
	} else {
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBOPN)
		if Bit(pdb.Status.PDBBitMask, PDBCLS) == true {
			pdb.Status.PDBBitMask = Bid(pdb.Status.PDBBitMask, PDBCLS)
		}
		pdb.Status.Message = "OPEN:OK"
		pdb.Status.OpenMode = "READ WRITE"
		pdb.Status.PDBStatus = "OPEN"
		r.UpdateStatus(ctx, pdb)
	}

	return nil

}

func (r *PDBReconciler) CreatePDB(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB) error {
	log := logf.FromContext(ctx).WithValues("CreatePDB", req.NamespacedName)

	var p1 C.struct_pdbinfo
	var addr uint64 /* addr of db handle */
	var CreationErr (*C.struct_OCIErrmsg)
	var dbhandle (*C.struct_OCIHandlePool)

	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.pdbname)), len(pdb.Spec.PDBName)), pdb.Spec.PDBName)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.totalSize)), len(pdb.Spec.TotalSize)), pdb.Spec.TotalSize)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.tempSize)), len(pdb.Spec.TempSize)), pdb.Spec.TempSize)

	if pdb.Spec.FileNameConversions != "" {
		NmConvert := "(" + pdb.Spec.FileNameConversions + ")"
		copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.filenamecon)), len(NmConvert)), NmConvert)
	} else {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.filenamecon)), len("NONE")), "NONE")
	}

	if *pdb.Spec.ReuseTempFile == true {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.reusetmp)), len("TRUE")), "TRUE")
		log.Info("resuse tempfile=true")
	}
	if *pdb.Spec.UnlimitedStorage == true {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.unlimitedstg)), len("TRUE")), "TRUE")
	}

	uname, _ := r.getSecret(ctx, req, pdb, pdb.Spec.Pdbuser.Secret.SecretName, pdb.Spec.Pdbuser.Secret.Key)
	passw, _ := r.getSecret(ctx, req, pdb, pdb.Spec.Pdbpass.Secret.SecretName, pdb.Spec.Pdbpass.Secret.Key)
	uname = strings.TrimSpace(uname)
	passw = strings.TrimSpace(passw)

	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.adminName)), len(uname)), uname)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.adminPwd)), len(passw)), passw)

	fmt.Sscanf(pdb.Status.Dbhandle, "0x%X", &addr)
	dbhandle = C.CastPtr(C.ulong(pdb.Status.Dbhandle64))

	//C.Create_Pdb(C.CastPtr(C.ulong(addr)), SQLEFE, (*C.pdbinfo)(&p1))
	C.OCICreatePdb(dbhandle, SQLEFE, (*C.pdbinfo)(&p1))

	CreationErr = dbhandle.errmsg

	log.Info("SQLCA.SQLCODE:" + fmt.Sprintf("%d", CreationErr.errcode))
	log.Info("SQLCA.ERRMSG=" + fmt.Sprintf("%s", CreationErr.errbuf))

	if CreationErr.errcode != 0 {
		pdb.Status.Message = fmt.Sprintf("ORA-%d", CreationErr.errcode)
		r.Recorder.Eventf(pdb, corev1.EventTypeWarning, "creation issue ", "Error:%s", pdb.Status.Message)
		r.UpdateStatus(ctx, pdb)
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBCRE)
		return errors.New(pdb.Status.Message)
	} else {
		pdb.Status.TotalSize = pdb.Spec.TotalSize
		pdb.Status.Message = "CREATE:OK"
		pdb.Status.OpenMode = "MOUNT"
		pdb.Status.ConnString = pdb.Spec.TNSstrg
		ParseTnsAlias(&(pdb.Status.ConnString), &(pdb.Spec.PDBName))
		r.Recorder.Eventf(pdb, corev1.EventTypeNormal, "create pdb", "pdbname=%s", pdb.Spec.PDBName)
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBCRT)
		r.UpdateStatus(ctx, pdb)
	}

	return nil
}

func (r *PDBReconciler) DropPDB(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB) error {
	log := logf.FromContext(ctx).WithValues("DropPDB", req.NamespacedName)
	log.Info("Deleting PDB")

	var p1 C.struct_pdbinfo
	var DeletionErr (*C.struct_OCIErrmsg)
	var dbhandle (*C.struct_OCIHandlePool)
	dbhandle = C.CastPtr(C.ulong(pdb.Status.Dbhandle64)) /* get handle address */
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p1.pdbname)), len(pdb.Spec.PDBName)), pdb.Spec.PDBName)
	C.OCIDropPdb(dbhandle, SQLEFE, (*C.pdbinfo)(&p1))
	DeletionErr = dbhandle.errmsg

	log.Info("SQLCA.SQLCODE:" + fmt.Sprintf("%d", DeletionErr.errcode))
	log.Info("SQLCA.ERRMSG=" + fmt.Sprintf("%s", DeletionErr.errbuf))

	if DeletionErr.errcode != 0 {
		pdb.Status.Message = fmt.Sprintf("ORA-%d", DeletionErr.errcode)
		r.UpdateStatus(ctx, pdb)
		return errors.New(pdb.Status.Message)
	} else {
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBDIC)
		pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, PDBCLS)
		if Bit(pdb.Status.PDBBitMask, PDBOPN) == true {
			pdb.Status.PDBBitMask = Bid(pdb.Status.PDBBitMask, PDBOPN)
		}
		pdb.Status.TotalSize = pdb.Spec.TotalSize
		pdb.Status.Message = "CREATE:OK"
		pdb.Status.OpenMode = "MOUNT"
		r.UpdateStatus(ctx, pdb)
	}

	return nil

}

func (r *PDBReconciler) FreePdbHandle(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB) error {
	log := logf.FromContext(ctx).WithValues("FreePdbHandle", req.NamespacedName)
	log.Info("Closeing session and releasing memory")

	var dbhandle (*C.struct_OCIHandlePool)
	dbhandle = C.CastPtr(C.ulong(pdb.Status.Dbhandle64))
	C.OCIFreeHandle(dbhandle)
	return nil
}

func (r *PDBReconciler) SetFinalizer(ctx context.Context, req ctrl.Request, pdb *databasev4.PDB) error {
	log := logf.FromContext(ctx).WithValues("SetFinalizer", req.NamespacedName)
	if pdb.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(pdb, PDBFinalizer) {
			log.Info("Add finalizer:" + PDBFinalizer)
			controllerutil.AddFinalizer(pdb, PDBFinalizer)
			if err := r.Update(ctx, pdb); err != nil {
				return errors.New("Update finalizer error")

			}
			pdb.Status.PDBBitMask = Bis(pdb.Status.PDBBitMask, FNALAZ)
			pdb.Status.PDBBitMaskStr = Bitmaskprint(pdb.Status.PDBBitMask)
			r.UpdateStatus(ctx, pdb)
		}

	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev4.PDB{}).
		Named("pdb").
		Complete(r)
}
