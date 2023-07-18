#!/usr/bin/sh

# usage ./run-geom-preproc.sh /sim/model/ Muncie 01

MODELDIR=$1
MODEL=$2

RAS_LIB_PATH=/ras/libs:/ras/libs/mkl:/ras/libs/rhel_8
export LD_LIBRARY_PATH=$RAS_LIB_PATH:$LD_LIBRARY_PATH

RAS_EXE_PATH=/ras:/ras/bin
export PATH=$RAS_EXE_PATH:$PATH

cd $MODELDIR
cp $2.g$3.hdf $2.g$3.tmp.hdf
RasGeomPreprocess $2.x$3