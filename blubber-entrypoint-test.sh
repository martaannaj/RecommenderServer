# !/usr/bin/sh bash

echo "Starting RecommenderServer..."

DIR=`pwd`

export GOROOT=/usr/lib/go-1.21
export PATH=${GOROOT}/bin:${PATH}
# export GOPATH=/srv/goProjects
# export PATH=${GOPATH}/bin:${PATH}

cd ${DIR}/recommender

go test ./...
