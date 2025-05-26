#include <oci.h>

OCIHandlePool *OpenConnection (text * username, text * password,
			       text * tnsalias, FILE * logfile);
OCIHandlePool *OCIConnectRDBMS (text * uname, text * psswd, text * tnsal);
void OCICheckConnection (OCIHandlePool * Hp, ub1 * status);
void FreeOCIcred (OCICred * ptr);
void HandleCursor (OCIHandlePool * Hp, OraText * sqltext, ub1 OpenIndex);
int ExecuteCrsidx ( /*... Execute cursor ... */ );
OCIHandlePool *CastPtr ( /*... cast to pointer ... */ );
void SetApplicationInfo ( /*... Set dbms application info .. */ );
ub1 OCICreatePdb (OCIHandlePool * db, ub1 crsidx, pdbinfo * ptr);
ub1 OCIOpenPdb (OCIHandlePool * db, ub1 crsidx, pdbinfo * ptr);
ub1 OCIClosePdb (OCIHandlePool * db, ub1 crsidx, pdbinfo * ptr);
ub1 OCIDropPdb (OCIHandlePool * db, ub1 crsidx, pdbinfo * ptr);
void CloseCursor ( /*.... close cursor .... */ );
OCIHandlePool *CastPtr (ub8 addr);
void OCIFreeHandle (OCIHandlePool * Hp);


static void checkerr ( /* check error */ );
static void checkerr2 ( /* check error */ );
static void checkerr3 ( /* check error */ );

/* WIP */
OraText *sqltext[] = { "select ........",
  "select ........",
  "insert ........"
};

#define LOGCRS 0
#define CHKWTR 1
#define UPTHST 2


OCIHandlePool *
OCIConnectRDBMS (uname, psswd, tnsal)
     text *uname;
     text *psswd;
     text *tnsal;
{
  FILE *logfile;
  logfile = fopen ("/tmp/testfile", "a+");
  OCIHandlePool *Hp = OpenConnection (uname, psswd, tnsal, logfile);
  fflush (logfile);
  return ((OCIHandlePool *) Hp);
}

OCIHandlePool *
OpenConnection (username, password, tnsalias, logfile)
     text *username;
     text *password;
     text *tnsalias;
     FILE *logfile;
{
  OCIHandlePool *Hp;
  text ServerVersion[1024];
  text buffver[100];
  ub4 version;

  OCIMALLOC (Hp);
  OCIMERROR (Hp);
  fprintf (stdout, "starting connection to database as %s  \n",
	   (char *) username);

  OCIInitialize ((ub4) OCI_OBJECT, (dvoid *) 0, (dvoid * (*)())0,
		 (dvoid * (*)())0, (void (*)()) 0);

  OCIEnvInit ((OCIEnv **) & Hp->envhp, (ub4) OCI_DEFAULT, (size_t) 0,
	      (dvoid **) 0);

  OCIHandleAlloc ((dvoid *) Hp->envhp, (dvoid **) & Hp->errhp,
		  (ub4) OCI_HTYPE_ERROR, (size_t) 0, (dvoid **) 0);

  OCIHandleAlloc ((dvoid *) Hp->envhp, (dvoid **) & Hp->srvhp,
		  (ub4) OCI_HTYPE_SERVER, (size_t) 0, (dvoid **) 0);

  checkerr (Hp->errhp, OCIServerAttach ((OCIServer *) Hp->srvhp,
					(OCIError *) Hp->errhp,
					(dvoid *) ((text *) tnsalias),
					strlen ((text *) tnsalias),
					OCI_DEFAULT));

  OCIHandleAlloc ((dvoid *) Hp->envhp, (dvoid **) & Hp->svchp,
		  (ub4) OCI_HTYPE_SVCCTX, (size_t) 0, (dvoid **) 0);

  OCIAttrSet ((dvoid *) Hp->svchp, (ub4) OCI_HTYPE_SVCCTX,
	      (dvoid *) Hp->srvhp, (ub4) 0,
	      (ub4) OCI_ATTR_SERVER, (OCIError *) Hp->errhp);

  OCIHandleAlloc ((dvoid *) Hp->envhp, (dvoid **) & Hp->usrhp,
		  (ub4) OCI_HTYPE_SESSION, (size_t) 0, (dvoid **) 0);

  OCIAttrSet ((dvoid *) Hp->usrhp, (ub4) OCI_HTYPE_SESSION,
	      (dvoid *) username,
	      (ub4) strlen ((text *) username),
	      (ub4) OCI_ATTR_USERNAME, (OCIError *) Hp->errhp);

  OCIAttrSet ((dvoid *) Hp->usrhp, (ub4) OCI_HTYPE_SESSION,
	      (dvoid *) password,
	      (ub4) strlen ((text *) password),
	      (ub4) OCI_ATTR_PASSWORD, (OCIError *) Hp->errhp);

  checkerr3 (Hp->errhp, OCISessionBegin ((OCISvcCtx *) Hp->svchp,
					 (OCIError *) Hp->errhp,
					 (OCISession *) Hp->usrhp,
					 OCI_CRED_RDBMS,
					 OCI_DEFAULT | OCI_SYSDBA),
	     Hp->errmsg);

  OCIAttrSet ((dvoid *) Hp->svchp, (ub4) OCI_HTYPE_SVCCTX,
	      (dvoid *) Hp->usrhp, (ub4) 0,
	      (ub4) OCI_ATTR_SESSION, (OCIError *) Hp->errhp);


  SetApplicationInfo ((OCIHandlePool *) Hp);


  checkerr (Hp->errhp,
	    OCIServerVersion (Hp->svchp, Hp->errhp, (text *) ServerVersion,
			      (ub4) sizeof (ServerVersion),
			      (ub1) OCI_HTYPE_SVCCTX));
  checkerr (Hp->errhp,
	    OCIServerRelease2 (Hp->svchp, Hp->errhp, (text *) buffver,
			       sizeof (buffver), OCI_HTYPE_SVCCTX, &version,
			       OCI_DEFAULT));

  fprintf (logfile, "OCIServerVersion: %s\n", ServerVersion);
  fprintf (logfile, "OCIServerRelease2:\t %i.%i.%i.%i.%i\n",
	   OCI_SERVER_RELEASE_REL (version),
	   OCI_SERVER_RELEASE_REL_UPD (version),
	   OCI_SERVER_RELEASE_REL_UPD_REV (version),
	   OCI_SERVER_RELEASE_REL_UPD_INC (version),
	   OCI_SERVER_RELEASE_EXT (version));


  sword major, minor, update, patch, port_update;
  OCIClientVersion (&major, &minor, &update, &patch, &port_update);
  fprintf (stdout, "OCIClientVersion:\t %i.%i.%i.%i.%i\n", major, minor,
	   update, patch, port_update);

  fprintf (stdout, "Opening Cursors\n");
  HandleCursor (Hp, sqltext[CHKWTR], CHKWTR);
  return ((OCIHandlePool *) Hp);
}

void
HandleCursor (Hp, sqltext, OpenIndex)
     OCIHandlePool *Hp;
     OraText *sqltext;
     ub1 OpenIndex;
{
#ifdef DEBUG
  fprintf (stdout, "====Openursor===\n");
  fprintf (stdout, "sqltext=%s\n", (char *) sqltext);
  fprintf (stdout, " Hp->stmthpidx[OpenIndex]=%p\n",
	   Hp->stmthpidx[OpenIndex]);
#endif
  checkerr (Hp->errhp,
	    OCIHandleAlloc ((dvoid *) Hp->envhp,
			    (dvoid **) & Hp->stmthpidx[OpenIndex],
			    (ub4) OCI_HTYPE_STMT, (size_t) 0, (dvoid **) 0));

  checkerr (Hp->errhp,
	    OCIStmtPrepare ((OCIStmt *) Hp->stmthpidx[OpenIndex],
			    (OCIError *) Hp->errhp,
			    (CONST OraText *) sqltext,
			    (ub4) (strlen ((text *) sqltext)),
			    (ub4) OCI_NTV_SYNTAX, (ub4) OCI_DEFAULT));
}




void
FreeOCIcred (ptr)
     OCICred *ptr;
{

  memset (ptr->uname, 0x00, sizeof (text) * 255);
  memset (ptr->passw, 0x00, sizeof (text) * 255);
  memset (ptr->tnsal, 0x00, sizeof (text) * 1024);
  free (ptr->uname);
  free (ptr->passw);
  free (ptr->tnsal);
  memset (ptr, 0x00, sizeof (OCICred));
  free (ptr);
}


OCICred *
MallocOCIcred ()
{
  OCICred *cred = malloc (sizeof (OCICred));
  cred->uname = malloc (sizeof (text) * 255);
  cred->passw = malloc (sizeof (text) * 255);
  cred->tnsal = malloc (sizeof (text) * 1024);
  return ((OCICred *) cred);
}

static void
checkerr (errhp, status)
     OCIError *errhp;
     sword status;
{
  text errbuf[512];
  ub4 errcode;

  switch (status)
    {
    case OCI_SUCCESS:
      break;
    case OCI_SUCCESS_WITH_INFO:
      printf ("Error - OCI_SUCCESS_WITH_INFO\n");
      break;
    case OCI_NEED_DATA:
      printf ("Error - OCI_NEED_DATA\n");
      break;
    case OCI_NO_DATA:
      break;
    case OCI_ERROR:
      OCIErrorGet ((dvoid *) errhp, (ub4) 1, (text *) NULL, (sb4 *) & errcode,
		   (text *) errbuf, (ub4) sizeof (errbuf),
		   (ub4) OCI_HTYPE_ERROR);
      printf ("ErrorBuffer - %s", errbuf);
      printf ("ErrorCode   - %d\n", errcode);
      //    exit(0);
      break;
    case OCI_INVALID_HANDLE:
      printf ("Error - OCI_INVALID_HANDLE\n");
      break;
    case OCI_STILL_EXECUTING:
      printf ("Error - OCI_STILL_EXECUTING\n");
      break;
    case OCI_CONTINUE:
      printf ("Error - OCI_CONTINUE\n");
      break;
    default:
      break;
    }
}

static void
checkerr2 (errhp, status, logfile)
     OCIError *errhp;
     sword status;
     FILE *logfile;
{
  text errbuf[512];
  ub4 errcode;

  switch (status)
    {
    case OCI_SUCCESS:
      break;
    case OCI_SUCCESS_WITH_INFO:
      printf ("Error - OCI_SUCCESS_WITH_INFO\n");
      break;
    case OCI_NEED_DATA:
      printf ("Error - OCI_NEED_DATA\n");
      break;
    case OCI_NO_DATA:
      break;
    case OCI_ERROR:
      OCIErrorGet ((dvoid *) errhp, (ub4) 1, (text *) NULL, (sb4 *) & errcode,
		   (text *) errbuf, (ub4) sizeof (errbuf),
		   (ub4) OCI_HTYPE_ERROR);
      fprintf (logfile, "ErrorBuffer - %s", errbuf);
      fprintf (logfile, "ErrorCode   - %d\n", errcode);
      break;
    case OCI_INVALID_HANDLE:
      printf ("Error - OCI_INVALID_HANDLE\n");
      break;
    case OCI_STILL_EXECUTING:
      printf ("Error - OCI_STILL_EXECUTING\n");
      break;
    case OCI_CONTINUE:
      printf ("Error - OCI_CONTINUE\n");
      break;
    default:
      break;
    }
}

void
SetApplicationInfo (Hp)
     OCIHandlePool *Hp;
{
  text *ApplicationInfo = "CONTROLLERTEST";
  text *sqltext = "begin DBMS_APPLICATION_INFO.SET_MODULE(:b1,NULL); end;";
  text *traceid = "alter session set tracefile_identifier='CONTROLLERTEST'";
  checkerr (Hp->errhp, OCIHandleAlloc ((dvoid *) Hp->envhp,
				       (dvoid **) & Hp->stmthp,
				       (ub4) OCI_HTYPE_STMT,
				       (size_t) 0, (dvoid **) 0));
  checkerr (Hp->errhp,
	    OCIStmtPrepare ((OCIStmt *) Hp->stmthp,
			    (OCIError *) Hp->errhp,
			    (CONST OraText *) sqltext,
			    (ub4) (strlen ((text *) sqltext)),
			    (ub4) OCI_NTV_SYNTAX, (ub4) OCI_DEFAULT));
  checkerr (Hp->errhp,
	    OCIBindByName ((OCIStmt *) Hp->stmthp,
			   (OCIBind **) & Hp->bnd,
			   (OCIError *) Hp->errhp,
			   (OraText *) ":b1", strlen (":b1"),
			   ApplicationInfo,
			   strlen ((char *) ApplicationInfo),
			   SQLT_CHR, (dvoid *) 0, (ub2 *) 0,
			   (ub2 *) 0, (ub4) 0, (ub4 *) 0, OCI_DEFAULT));
  checkerr (Hp->errhp,
	    OCIStmtExecute ((OCISvcCtx *) Hp->svchp,
			    (OCIStmt *) Hp->stmthp,
			    (OCIError *) Hp->errhp, (ub4) 1,
			    (ub4) 0, (CONST OCISnapshot *) NULL,
			    (OCISnapshot *) NULL, (ub4) OCI_DEFAULT));
  checkerr (Hp->errhp,
	    OCIStmtPrepare ((OCIStmt *) Hp->stmthp,
			    (OCIError *) Hp->errhp,
			    (CONST OraText *) traceid,
			    (ub4) (strlen ((text *) traceid)),
			    (ub4) OCI_NTV_SYNTAX, (ub4) OCI_DEFAULT));
  checkerr (Hp->errhp,
	    OCIStmtExecute ((OCISvcCtx *) Hp->svchp,
			    (OCIStmt *) Hp->stmthp,
			    (OCIError *) Hp->errhp, (ub4) 1,
			    (ub4) 0, (CONST OCISnapshot *) NULL,
			    (OCISnapshot *) NULL, (ub4) OCI_DEFAULT));
  checkerr (Hp->errhp,
	    OCIHandleFree ((dvoid *) Hp->stmthp, (ub4) OCI_HTYPE_STMT));
}


void
OCICheckConnection (Hp, ret)
     OCIHandlePool *Hp;
     ub1 *ret;
{
  ub1 Retry;
  text errbuf[512];
  ub4 errcode;
  ub1 rc = 0;

  rc = OCIPing ((dvoid *) Hp->svchp, Hp->errhp, OCI_DEFAULT);
  if (rc != 0)
    {
      text errbuf[512];
      ub4 errcode;
      OCIErrorGet ((dvoid *) Hp->svchp, (ub4) 1, (text *) NULL,
		   (sb4 *) & errcode, (text *) errbuf, (ub4) sizeof (errbuf),
		   (ub4) OCI_HTYPE_ERROR);
      if (errcode != 0)
	{
	  printf ("ErrorBuffer - %s", errbuf);
	  printf ("ErrorCode   - %d\n", errcode);
	  fprintf (stdout, "OCIPing error\n");
	}
    }

  memcpy (ret, &rc, sizeof (rc));

}



static void
checkerr3 (errhp, status, errmsg)
     OCIError *errhp;
     sword status;
     OCIErrmsg *errmsg;
{
  text errbuf[512];
  ub4 errcode;

  /** Reset from previous execution */
  memset ((text *) & (errmsg->errbuf), 0x00, 512);
  errmsg->errcode = 0;

  /*
   * reset output variable 
   */
  switch (status)
    {
    case OCI_ERROR:
      memset (errmsg, 0x00, sizeof (OCIErrmsg));
      OCIErrorGet ((dvoid *) errhp, (ub4) 1, (text *) NULL, (sb4 *) & errcode,
		   (text *) errbuf, (ub4) sizeof (errbuf),
		   (ub4) OCI_HTYPE_ERROR);
      errbuf[strlen ((text *) errbuf)] = '\0';
      memcpy ((text *) & (errmsg->errbuf), (text *) & errbuf,
	      strlen ((text *) errbuf));
      memcpy ((ub4 *) & (errmsg->errcode), (ub4 *) & errcode, sizeof (ub4));
      errmsg->lenmsg = (ub2) strlen ((text *) errbuf);

    case OCI_SUCCESS:
      break;
    case OCI_SUCCESS_WITH_INFO:
      /*
       * printf ("Error - OCI_SUCCESS_WITH_INFO\n");
       * break;
       */
      memset (errmsg, 0x00, sizeof (OCIErrmsg));
      OCIErrorGet ((dvoid *) errhp, (ub4) 1, (text *) NULL, (sb4 *) & errcode,
		   (text *) errbuf, (ub4) sizeof (errbuf),
		   (ub4) OCI_HTYPE_ERROR);
      errbuf[strlen ((text *) errbuf)] = '\0';
      memcpy ((text *) & (errmsg->errbuf), (text *) & errbuf,
	      strlen ((text *) errbuf));
      memcpy ((ub4 *) & (errmsg->errcode), (ub4 *) & errcode, sizeof (ub4));
      errmsg->lenmsg = (ub2) strlen ((text *) errbuf);

    case OCI_NEED_DATA:
      printf ("Error - OCI_NEED_DATA\n");
      memset (errmsg, 0x00, sizeof (OCIErrmsg));
      OCIErrorGet ((dvoid *) errhp, (ub4) 1, (text *) NULL, (sb4 *) & errcode,
		   (text *) errbuf, (ub4) sizeof (errbuf),
		   (ub4) OCI_HTYPE_ERROR);
      errbuf[strlen ((text *) errbuf)] = '\0';
      memcpy ((text *) & (errmsg->errbuf), (text *) & errbuf,
	      strlen ((text *) errbuf));
      memcpy ((ub4 *) & (errmsg->errcode), (ub4 *) & errcode, sizeof (ub4));
      errmsg->lenmsg = (ub2) strlen ((text *) errbuf);
      break;
    case OCI_INVALID_HANDLE:
      printf ("Error - OCI_INVALID_HANDLE\n");
      break;
    case OCI_STILL_EXECUTING:
      printf ("Error - OCI_STILL_EXECUTING\n");
      break;
    case OCI_CONTINUE:
      printf ("Error - OCI_CONTINUE\n");
      break;
    default:
      break;
    }
}




ub1
OCIClosePdb (db, crsidx, ptr)
     OCIHandlePool *db;
     ub1 crsidx;
     pdbinfo *ptr;
{
  ub1 rc;
  text sql00[SQLDDLLEN];
  text sql01[SQLDDLLEN] =
    "alter pluggable database \"%s\" close instances=all";

  memset (&sql00, 0x00, SQLDDLLEN);
  sprintf ((text *) & sql00, (char *) sql01, (text *) ptr->pdbname);

  sql00[strlen ((text *) sql00)] = '\0';

  HandleCursor (db, sql00, crsidx);
  ExecuteCrsidx (db, crsidx);
  CloseCursor (db, crsidx);

  fprintf (stdout, "---> %s\n", sql00);

  return ((ub1) rc);
}




ub1
OCIOpenPdb (db, crsidx, ptr)
     OCIHandlePool *db;
     ub1 crsidx;
     pdbinfo *ptr;
{
  ub1 rc;

  text sql00[SQLDDLLEN];
  text sql01[SQLDDLLEN] = "alter pluggable database %s open instances=all";

  memset (&sql00, 0x00, SQLDDLLEN);
  sprintf ((text *) & sql00, (char *) sql01, (text *) ptr->pdbname);

  sql00[strlen ((text *) sql00)] = '\0';

  HandleCursor (db, sql00, crsidx);
  ExecuteCrsidx (db, crsidx);
  CloseCursor (db, crsidx);

  fprintf (stdout, "---> %s\n", sql00);

  return ((ub1) rc);
}

ub1
OCIDropPdb (db, crsidx, ptr)
     OCIHandlePool *db;
     ub1 crsidx;
     pdbinfo *ptr;
{

  ub1 rc;
  text sql00[SQLDDLLEN];
  text sql01[SQLDDLLEN] =
    "alter pluggable database \"%s\" close instances=all";
  text sql02[SQLDDLLEN] =
    "drop pluggable database \"%s\" including datafiles";
  fprintf(stdout,"call::OCIDropPdb");

  if (ptr->pdbname != NULL)
    {
      memset (&sql00, 0x00, SQLDDLLEN);

      sprintf ((text *) & sql00, (char *) sql01, (text *) ptr->pdbname);
      sql00[strlen ((text *) sql00)] = '\0';
      HandleCursor (db, sql00, crsidx);
      ExecuteCrsidx (db, crsidx);
      CloseCursor (db, crsidx);

      memset (&sql00, 0x00, SQLDDLLEN);
      sprintf ((text *) & sql00, (char *) sql02, (text *) ptr->pdbname);
      sql00[strlen ((text *) sql00)] = '\0';
      HandleCursor (db, sql00, crsidx);
      ExecuteCrsidx (db, crsidx);
      CloseCursor (db, crsidx);
    }

  return ((ub1) rc);
}


ub1
OCICreatePdb (db, crsidx, ptr)
     OCIHandlePool *db;
     ub1 crsidx;
     pdbinfo *ptr;
{
  ub1 rc;
  text sql00[SQLDDLLEN];
  text sql01[SQLDDLLEN] = "create pluggable database \"%s\" \
admin user \"%s\" identified by \"%s\" roles = (DBA) \
storage ( maxsize %s max_shared_temp_size %s ) ";


  if (STRNCMP (ptr->reusetmp, "TRUE") || STRNCMP (ptr->reusetmp, "true"))
    {
      STRCAT (sql01, " tempfile reuse ");
    }

  if (STRNCMP (ptr->unlimitedstg, "TRUE")
      || STRNCMP (ptr->unlimitedstg, "true"))
    {
      fprintf (stdout, "unlimitedstg not yet implemented\n");
    }

  STRCAT (sql01, " file_name_convert=%s ");


  memset (&sql00, 0x00, SQLDDLLEN);
  sprintf ((text *) & sql00, (char *) sql01, (text *) ptr->pdbname,
	   (text *) ptr->adminName, (text *) ptr->adminPwd,
	   (text *) ptr->totalSize, (text *) ptr->tempSize,
	   (text *) ptr->filenamecon);

  sql00[strlen ((text *) sql00)] = '\0';

  HandleCursor (db, sql00, crsidx);
  ExecuteCrsidx (db, crsidx);
  CloseCursor (db, crsidx);

  fprintf (stdout, "---> %s\n", sql00);
  return (rc);
}

int
ExecuteCrsidx (Hp, crsidx)
     OCIHandlePool *Hp;
     ub1 crsidx;
{


  checkerr3 (Hp->errhp, OCIStmtExecute ((OCISvcCtx *) Hp->svchp,
					(OCIStmt *)
					Hp->stmthpidx[crsidx],
					(OCIError *) Hp->errhp,
					(ub4) 1, (ub4) 0,
					(CONST OCISnapshot *)
					NULL,
					(OCISnapshot *) NULL,
					(ub4) OCI_DEFAULT), Hp->errmsg);
  return ((int) 0);
}



void
CloseCursor (Hp, OpenIndex)
     OCIHandlePool *Hp;
     ub1 OpenIndex;
{

  checkerr (Hp->errhp,
	    OCIHandleFree ((dvoid *) Hp->stmthpidx[OpenIndex],
			   (ub4) OCI_HTYPE_STMT));
}



OCIHandlePool *
CastPtr (addr)
     ub8 addr;
{
  OCIHandlePool *Hp;
#ifdef DEBUG
  fprintf (stdout, "DEBUG sizeof(OCIHandlePool)=%i\n",
	   sizeof (OCIHandlePool));
  fprintf (stdout, "DEBUG sizeof(Hp)=%i\n", sizeof (Hp));
  fprintf (stdout, "DEBUG sizeof(ub4)=%i\n", sizeof (ub4));
  fprintf (stdout, "DEBUG sizeof(ub8)=%i\n", sizeof (ub8));
  fprintf (stdout, "DEBUG %lu\n", addr);
#endif
  Hp = (OCIHandlePool *) addr;
  return ((OCIHandlePool *) Hp);
}


void
OCIFreeHandle (Hp)
     OCIHandlePool *Hp;
{

  checkerr3 (Hp->errhp,
	     OCISessionEnd ((OCISvcCtx *) Hp->svchp, (OCIError *) Hp->errhp,
			    (OCISession *) Hp->usrhp, (ub4) OCI_DEFAULT),
	     Hp->errhp);



  checkerr3 (Hp->errhp, OCIServerDetach ((OCIServer *) Hp->srvhp,
					 (OCIError *) Hp->errhp,
					 (ub4) OCI_DEFAULT), Hp->errhp);
  /*
   * Cursor closure is managed separatly 
   if (stmthp) {
   OCIHandleFree((dvoid *)stmthp, (ub4) OCI_HTYPE_STMT);
   }
   */

  free (Hp->errmsg);


  if (Hp->usrhp)
    {
      OCIHandleFree ((dvoid *) Hp->usrhp, (ub4) OCI_HTYPE_SESSION);
    }
  if (Hp->svchp)
    {
      OCIHandleFree ((dvoid *) Hp->svchp, (ub4) OCI_HTYPE_SVCCTX);
    }
  if (Hp->srvhp)
    {
      OCIHandleFree ((dvoid *) Hp->srvhp, (ub4) OCI_HTYPE_SERVER);
    }
  if (Hp->errhp)
    {
      OCIHandleFree ((dvoid *) Hp->errhp, (ub4) OCI_HTYPE_ERROR);
    }
  if (Hp->envhp)
    {
      OCIHandleFree ((dvoid *) Hp->envhp, (ub4) OCI_HTYPE_ENV);
    }

  free (Hp);
}
