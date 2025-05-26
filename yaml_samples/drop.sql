alter pluggable database pdb1_TNT close instances=all;
drop pluggable database pdb1_TNT  including datafiles;
alter pluggable database pdb2_TNT close instances=all;
drop pluggable database pdb2_TNT  including datafiles;
exit;
