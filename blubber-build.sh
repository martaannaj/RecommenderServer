#!/usr/bin/env bash

# clean up previous builds
docker rm recommenderserver-test
docker rmi --force recommenderserver-test

docker rm recommenderserver
docker rmi --force recommenderserver

# build docker
blubber .pipeline/blubber.yaml test | docker build --tag recommenderserver-test --file - .
blubber .pipeline/blubber.yaml production | docker build --tag recommenderserver --file - .