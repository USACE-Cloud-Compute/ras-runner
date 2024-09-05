#!/usr/bin/sh

# usage ./run.sh /sim/model/ Muncie 01 01

MODELDIR=$1
MODEL=$2

RAS_LIB_PATH=/ras/libs:/ras/libs/mkl:/ras/libs/rhel_8
export LD_LIBRARY_PATH=$RAS_LIB_PATH:$LD_LIBRARY_PATH

RAS_EXE_PATH=/ras:/ras/bin
export PATH=$RAS_EXE_PATH:$PATH

cd $MODELDIR
RasUnsteady $2.p$3.tmp.hdf x$4