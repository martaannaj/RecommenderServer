# !/usr/bin/sh bash

echo "Starting RecommenderServer..."

DIR=`pwd`

export GOROOT=${DIR}/go
export GOPATH=${DIR}/goProjects
export PATH=${GOPATH}/bin:${GOROOT}/bin:${PATH}

cd ${DIR}/recommender

./RecommenderServer serve ./testdata/latest-truthy-item-filtered-sorted.nt.gz.schemaTree.typed.bin -p 8771
PID=$!
sleep 10
if ! kill -0 ${PID}; then
  echo "ERROR: Service process has prematurely ended!"
  exit 1
fi

exit ${EXIT_CODE}