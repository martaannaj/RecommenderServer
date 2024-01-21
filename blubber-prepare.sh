# !/usr/bin/sh bash

mkdir recommender
ls | grep -v recommender | xargs mv -t recommender

m_error() {
    echo $1
    exit 2
}

build_recommenderserver() {
  cd /srv/recommender
  if ! go build .; then
    m_error "Failed to build RecommenderServer!"
  fi
}

export GOROOT=/usr/lib/go-1.21
export PATH=${GOROOT}/bin:${PATH}
# export GOPATH=/srv/goProjects
# export PATH=${GOPATH}/bin:${PATH}

build_recommenderserver

echo "Successfully prepared RecommenderServer!"