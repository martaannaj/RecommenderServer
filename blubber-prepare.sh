# !/usr/bin/sh bash

mkdir recommender
ls | grep -v recommender | xargs mv -t recommender

m_error() {
    echo $1
    exit 2
}

install_go() {
  echo "Installing Go"
  cd /srv
  if [ ! -f /tmp/go1.13.linux-amd64.tar.gz ]; then
   if ! wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz -O /tmp/go1.13.linux-amd64.tar.gz; then
     m_error "Unable to download Go lang 1.13 from Google!"
   fi
  fi
  tar xvfz /tmp/go1.13.linux-amd64.tar.gz
  echo "Go installed"
}

build_recommenderserver() {
  cd /srv/recommender
  if ! go build .; then
    m_error "Failed to build RecommenderServer!"
  fi
}


install_go

export GOROOT=/srv/go
export GOPATH=/srv/goProjects
export PATH=${GOPATH}/bin:${GOROOT}/bin:${PATH}

build_recommenderserver


echo "Successfully prepared RecommenderServer!"