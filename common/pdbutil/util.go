package pdbutil

import (
	"fmt"
	"regexp"
	"strings"
)

// * STATE TABLE *//
const (
	PDBCRT = 0x00000001 /* Create pdb */
	PDBOPN = 0x00000002 /* Open pdb read write */
	PDBCLS = 0x00000004 /* Close pdb */
	PDBDIC = 0x00000008 /* Drop pdb include datafiles */
	OCIHDL = 0x00000010 /* OCI handle allocation */
	OCICON = 0x00000020 /* Rdbms connection */
	FNALAZ = 0x00000040 /* Finalizer configured */
	/* Error section */
	PDBCRE = 0x00001000 /* PDB creation error */
	PDBOPE = 0x00002000 /* PDB open error */
	PDBCLE = 0x00004000 /* PDB close error */
	OCIHDE = 0x00008000 /* Allocation Handle Error */
	OCICOE = 0x00010000 /* CDD connection Error */
	FNALAE = 0x00020000
)

func ParseTnsAlias(tns *string, lrpdbsrv *string) {
	fmt.Printf("Analyzing string [%s]\n", *tns)
	fmt.Printf("Relacing  srv [%s]\n", *lrpdbsrv)
	var swaptns string

	if strings.Contains(strings.ToUpper(*tns), "SERVICE_NAME") == false {
		fmt.Print("Cannot generate tns alias for pdb")
		return
	}

	if strings.Contains(strings.ToUpper(*tns), "ORACLE_SID") == true {
		fmt.Print("Cannot generate tns alias for pdb")
		return
	}

	*tns = strings.ReplaceAll(*tns, " ", "")

	swaptns = fmt.Sprintf("SERVICE_NAME=%s", *lrpdbsrv)
	tnsreg := regexp.MustCompile(`SERVICE_NAME=\w+`)
	*tns = tnsreg.ReplaceAllString(*tns, swaptns)

	fmt.Printf("Newstring [%s]\n", *tns)

}

func Bid(bitmask int, bitval int) int {
	bitmask ^= ((bitval) & (bitmask))
	return bitmask
}

func Bit(bitmask int, bitval int) bool {
	if bitmask&bitval != 0 {
		return true
	} else {
		return false
	}
}

func Bis(bitmask int, bitval int) int {
	bitmask = ((bitmask) | (bitval))
	return bitmask
}

func Bitmaskprint(bitmask int) string {
	BitRead := "|"
	if Bit(bitmask, PDBCRT) {
		BitRead = strings.Join([]string{BitRead, "PDBCRT|"}, "")
	}
	if Bit(bitmask, PDBOPN) {
		BitRead = strings.Join([]string{BitRead, "PDBOPN|"}, "")
	}
	if Bit(bitmask, PDBCLS) {
		BitRead = strings.Join([]string{BitRead, "PDBCLS|"}, "")
	}
	if Bit(bitmask, PDBDIC) {
		BitRead = strings.Join([]string{BitRead, "PDBDIC|"}, "")
	}
	if Bit(bitmask, OCIHDL) {
		BitRead = strings.Join([]string{BitRead, "OCIHDL|"}, "")
	}
	if Bit(bitmask, OCICON) {
		BitRead = strings.Join([]string{BitRead, "OCICON|"}, "")
	}
	if Bit(bitmask, FNALAZ) {
		BitRead = strings.Join([]string{BitRead, "FNALAZ|"}, "")
	}

	if Bit(bitmask, PDBCRE) {
		BitRead = strings.Join([]string{BitRead, "PDBCRE|"}, "")
	}
	if Bit(bitmask, PDBOPE) {
		BitRead = strings.Join([]string{BitRead, "PDBOPE|"}, "")
	}
	if Bit(bitmask, PDBCLE) {
		BitRead = strings.Join([]string{BitRead, "PDBCLE|"}, "")
	}
	if Bit(bitmask, OCIHDE) {
		BitRead = strings.Join([]string{BitRead, "OCIHDE|"}, "")
	}
	if Bit(bitmask, OCICOE) {
		BitRead = strings.Join([]string{BitRead, "OCICOE|"}, "")
	}
	if Bit(bitmask, FNALAE) {
		BitRead = strings.Join([]string{BitRead, "FNALAE|"}, "")
	}

	BitRead = fmt.Sprintf("[%d]%s", bitmask, BitRead)
	return BitRead
}
