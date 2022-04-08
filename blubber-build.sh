#!/usr/bin/env bash

# clean up previous builds
docker rm recommenderserver-test
docker rmi --force recommenderserver-test

docker rm recommenderserver
docker rmi --force recommenderserver

blubber() {
  if [ $# -lt 2 ]; then
    echo 'Usage: blubber config.yaml variant'
    return 1
  fi
  curl -s -H 'content-type: application/yaml' --data-binary @"$1" https://blubberoid.wikimedia.org/v1/"$2"
}

# build docker
blubber .pipeline/blubber.yaml test | docker build --tag recommenderserver-test --file - .
blubber .pipeline/blubber.yaml production | docker build --tag recommenderserver --file - .