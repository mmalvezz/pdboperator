typedef struct OCIErrmsg
{
  text errbuf[512];
  ub4  errcode;
  ub2  lenmsg;
} OCIErrmsg;

#define STRLEN(A)    (strlen((text *)A))

#define XORL(A,B)    (STRLEN(A) ^ STRLEN(B))

#define STRNCMP(A,B) ((!strncmp((text *) A,(const text *) B,strlen((const text*)B))) && (XORL(A,B) == 0))

#define STRCAT(a,b) memcpy((text *)&a[strlen((text *)a)], \
                          (const text *)b, \
                          strlen((const text *)b)) \

#define NT(A) (A[strlen((const text*)A)]='\0')

#define MEMSET(A)    (memset(A,0x00,strlen(A)))

#define CMP(A,B) (!strncmp((text *) A,(const text *) B,strlen((const text*)B)))

#define NVL(a) (( strlen(a) == 0  ) ? (const text *)"NULL" : a )


typedef struct OCIHandlePool
{
#define OCIMALLOC(ocim) (ocim=(OCIHandlePool *)malloc(sizeof(OCIHandlePool)))
#define OCIMERROR(ocim) (ocim->errmsg=malloc(sizeof(OCIErrmsg)))
#define MAXOPENCRS 100
  OCIEnv       *envhp;
  OCIError     *errhp;
  OCIServer    *srvhp;
  OCISvcCtx    *svchp;
  OCISession   *usrhp;
  OCIStmt      *stmthp;
  OCIStmt      *stmthpidx[MAXOPENCRS];
  OCIDefine    *def;
  OCIBind      *bnd;
  OCIErrmsg *errmsg;

  ub2 spare;
} OCIHandlePool;

typedef struct OCICred 
{
#define OCICREDMALLOC(ocic) (ocic=(OCICred *)malloc(sizeof(OCICred)))
 text  *uname;
 text  *passw;
 text  *tnsal;
} OCICred;


#define SHRSTR 30
#define LNGSTR 1024
#define SQLDDLLEN 512                   /* Max ddl statement length */


typedef struct pdbinfo
{
  ub2 id;
  text pdbname[SHRSTR];
  text srcPdbName[SHRSTR];
  text clonePDBName[SHRSTR];
  text method[SHRSTR];
  text adminName[SHRSTR];
  text adminPwd[SHRSTR];
  text filenamecon[LNGSTR];
  text reusetmp[SHRSTR];
  text unlimitedstg[SHRSTR];
  text totalSize[SHRSTR];
  text tempSize[SHRSTR];
  text getScript[SHRSTR];
  text state[SHRSTR];
  text action[SHRSTR];
  text tdeExport[SHRSTR];
  text copyAction[SHRSTR];
  text xmlFileName[LNGSTR];
  text tdeKeystorePath[LNGSTR];
  text tdepassword[SHRSTR];
  text tdesecret[SHRSTR];
  text modifyOption[SHRSTR];
  text modifyOption2[SHRSTR];
  text sourceFileNameConversions[LNGSTR];
  text serviceNameConvert[LNGSTR];
  text alterSystem[LNGSTR];
  text altSysPrm[LNGSTR];
  text altSysVal[LNGSTR];
  text parameterScope[SHRSTR];
  int unplug;
// This is a quick and dirty solution
// if other state , whatever reason , will be required then we
// have to manage with a bitmask
#define      OFF 0
#define      ON  1
#define      SETOFF(A) (A->unplug=OFF)
#define      SETON(A)   (A->unplug=ON)
} pdbinfo;

/* Cursor Index */

#define  PDBNAM   0 /* Select pdbname and open_mode etc ttc from v$pdb */
#define  PDBSPC   1 /* select * from v$pdb  */
#define  PDBDET   2 /* select * from v$pdbs where name = :b1 */
#define  PDBVIL   3 /* Spare 2 */
#define  SQLEFE   4 /* Ephemeral: used for ddl */


