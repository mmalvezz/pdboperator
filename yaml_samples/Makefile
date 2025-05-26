# Copyright (c) 2022, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at http://oss.oracle.com/licenses/upl.
#   __  __       _         __ _ _
#  |  \/  | __ _| | _____ / _(_) | ___
#  | |\/| |/ _` | |/ / _ \ |_| | |/ _ \
#  | |  | | (_| |   <  __/  _| | |  __/
#  |_|  |_|\__,_|_|\_\___|_| |_|_|\___|
#  | | | | ___| |_ __   ___ _ __
#  | |_| |/ _ \ | '_ \ / _ \ '__|
#  |  _  |  __/ | |_) |  __/ |
#  |_| |_|\___|_| .__/ \___|_|
#               |_|


export PARAMETERS=parameters.txt
export TNSSTR=$(shell cat $(PARAMETERS)  |grep -v ^\#|grep TNSSTR|cut -d :  -f 2)
export SYSUSR=$(shell cat $(PARAMETERS)|grep -v ^\#|grep SYSUSR|cut -d :  -f 2)
export SYSPWD=$(shell cat $(PARAMETERS)|grep -v ^\#|grep SYSPWD|cut -d :  -f 2)
export PDBUSR=$(shell cat $(PARAMETERS)|grep -v ^\#|grep PDBUSR|cut -d :  -f 2)
export PDBPWD=$(shell cat $(PARAMETERS)|grep -v ^\#|grep PDBPWD|cut -d :  -f 2)
export PDBNMS=$(shell cat $(PARAMETERS)|grep -v ^\#|grep PDB_NAMESPACE|cut -d :  -f 2)
export OPRNMS=$(shell cat $(PARAMETERS)|grep -v ^\#|grep OPERATOR_NAMESPACE|cut -d :  -f 2)
export FILCVP=$(shell cat $(PARAMETERS)|grep -v ^\#|grep FILECONVERSIONPATH1|cut -d :  -f 2)
export FILCVQ=$(shell cat $(PARAMETERS)|grep -v ^\#|grep FILECONVERSIONPATH2|cut -d :  -f 2)
export TEST_EXEC_TIMEOUT=3m


## BIN ##
export BASE64=/usr/bin/base64
export SED=/usr/bin/sed
export KUBECTL=/usr/bin/kubectl
export SQLPLUS=$(ORACLE_HOME)/bin/sqlplus -s '/as sysdba' 

## FILENAME ##
export SYSSEC=sys_secrets.yaml
export PDBSEC=pdb_secrets.yaml
export OPRFIL=../pdboperator.yaml
export PDBSM1=pdb_sample_01.yaml
export PDBSM2=pdb_sample_02.yaml
export PDBBND=./pdbnamespace_binding.yaml

## OBJECTS ##
export SYSSECNM=syssecret
export PDBSECNM=pdbsecret
export PDBNAME1=pdb1
export PDBNAME2=pdb2

define _genyaml

echo "Writing files:"
echo "------------------------"

export B64SYSUSR=`echo ${SYSUSR}|${BASE64}` 
export B64SYSPWD=`echo ${SYSPWD}|${BASE64}`
export B64PDBUSR=`echo ${PDBUSR}|${BASE64}`
export B64PDBPWD=`echo ${PDBPWD}|${BASE64}`

echo "${SYSSEC}"
echo "${PDBSEC}"
echo "${PDBSM1}"
echo `basename ${OPRFIL}`


cat <<EOF >${SYSSEC}
apiVersion: v1
kind: Secret
metadata:
  name: ${SYSSECNM}
  namespace: ${PDBNMS}
type: Opaque
data:
  sysusr: ${B64SYSUSR}
  syspwd: ${B64SYSPWD}
EOF

cat <<EOF >${PDBSEC}
apiVersion: v1
kind: Secret
metadata:
  name: ${PDBSECNM}
  namespace: ${PDBNMS}
type: Opaque
data:
  sysusr: ${B64PDBUSR}
  syspwd: ${B64PDBPWD}
EOF



cat <<EOF > ${PDBSM1}
apiVersion: database.oracle.com/v4
kind: PDB
metadata:
  name: ${PDBNAME1}
  namespace: ${PDBNMS}
spec:
  pdbName: ${PDBNAME1}_tnt
  tnsstring: ${TNSSTR}
  totalSize: "2G"
  tempSize: "800M"
  fileNameConversions: "${FILCVP}"
  reuseTempFile: true
  unlimitedStorage: true
  cdbSys:
    secret:
      secretName: "${SYSSECNM}"
      key: "sysusr"
  cdbPwd:
    secret:
      secretName: "${SYSSECNM}"
      key: "syspwd"
  pdbSys:
    secret:
      secretName: "${PDBSECNM}"
      key: "sysusr"
  pdbPwd:
    secret:
      secretName: "${PDBSECNM}"
      key: "syspwd"
EOF


cat <<EOF > ${PDBSM2}
apiVersion: database.oracle.com/v4
kind: PDB
metadata:
  name: ${PDBNAME2}
  namespace: ${PDBNMS}
spec:
  pdbName: ${PDBNAME2}_tnt
  tnsstring: ${TNSSTR}
  totalSize: "2G"
  tempSize: "800M"
  fileNameConversions: "${FILCVQ}"
  reuseTempFile: true
  unlimitedStorage: true
  cdbSys:
    secret:
      secretName: "${SYSSECNM}"
      key: "sysusr"
  cdbPwd:
    secret:
      secretName: "${SYSSECNM}"
      key: "syspwd"
  pdbSys:
    secret:
      secretName: "${PDBSECNM}"
      key: "sysusr"
  pdbPwd:
    secret:
      secretName: "${PDBSECNM}"
      key: "syspwd"
EOF


cat <<EOF >${PDBBND}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pdboperator-pdboperator-manager-rolebinding1
  namespace: pdbnamespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pdboperator-manager-role
subjects:
- kind: ServiceAccount
  name: default
  namespace: pdboperator-system
EOF


#cat ${OPRFIL} |sed 's/value: ""/value: pdboperator,"${PDBNMS}"/g'  > `basename ${OPRFIL}`
cat ${OPRFIL}  > `basename ${OPRFIL}`

endef

export gyaml = $(value _genyaml)

genyaml:
	 @ eval "$$gyaml"
	 $(SED) -i  's/value: ""/value: $(OPRNMS),$(PDBNMS)/g'   `basename ${OPRFIL}`

logplane:
	$(KUBECTL) logs -f -l control-plane=controller-manager -n $(OPRNMS)
	


clean:
	-- $(KUBECTL) delete -f $(PDBSM1)
	-- $(KUBECTL) delete -f $(OPRFIL)

clean1:
	-- $(KUBECTL) delete -f $(OPRFIL)

info:
	$(KUBECTL) describe clusterrole pdboperator-manager-role

login:
	$(KUBECTL) exec `$(KUBECTL) get pods -n pdboperator-system --no-headers|awk ' { print $1 }'` -n pdboperator-system -it -- /usr/bin/bash

run01: drop
	$(KUBECTL) apply -f $(PDBSM1)
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="MOUNT" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) patch pdb pdb1 -n $(PDBNMS) -p '{"spec":{"pdbState":"OPEN"}}' --type=merge
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="READ WRITE" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) patch pdb pdb1 -n $(PDBNMS) -p '{"spec":{"pdbState":"CLOSE"}}' --type=merge
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="MOUNT" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) delete pdb pdb1 -n $(PDBNMS)

run02: drop
	$(KUBECTL) apply -f $(PDBSM1)
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="MOUNT" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) delete pdb pdb1 -n $(PDBNMS)

run03: drop 
	-- $(KUBECTL) delete pdb pdb1 -n $(PDBNMS)
	$(KUBECTL) apply -f $(PDBSM1)
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="MOUNT" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) describe pdb pdb1 -n $(PDBNMS)

run04: drop
	$(KUBECTL) apply -f $(PDBSM1)
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="MOUNT" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) apply -f $(PDBSM2)
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="MOUNT" pdb pdb2 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) patch pdb pdb1 -n $(PDBNMS) -p '{"spec":{"pdbState":"OPEN"}}' --type=merge
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="READ WRITE" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) patch pdb pdb1 -n $(PDBNMS) -p '{"spec":{"pdbState":"CLOSE"}}' --type=merge
	$(KUBECTL) wait  --for jsonpath='{.status.openMode}'="MOUNT" pdb pdb1 -n $(PDBNMS) --timeout=$(TEST_EXEC_TIMEOUT)
	$(KUBECTL) get pdb -n $(PDBNMS)
	$(KUBECTL) delete pdb pdb1 -n $(PDBNMS)
	$(KUBECTL) delete pdb pdb2 -n $(PDBNMS)


rmfn:
	@echo "Clear finalizer"
	-- $(KUBECTL) patch pdb pdb1 -n $(PDBNMS) -p '{"metadata":{"finalizers":[]}}' --type=merge
	-- $(KUBECTL) patch pdb pdb2 -n $(PDBNMS) -p '{"metadata":{"finalizers":[]}}' --type=merge

diag:
	@echo "Dump operator trace file"
	@$(eval TMPSP := $(shell date "+%y%m%d%H%M%S" ))
	@$(eval DIAGFILE := ./opdmp.$(TMPSP))
	@>$(DIAGFILE)
	$(KUBECTL) logs `$(KUBECTL) get pods -n pdboperator-system --no-headers |  cut -d ' ' -f 1` -n pdboperator-system >>$(DIAGFILE)

drop:
	-- $(SQLPLUS) @drop.sql
