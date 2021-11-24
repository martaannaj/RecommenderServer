[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0) 

Individual descriptions in subfolders.


## Installation

1. Install the go runtime (and VS Code + Golang tools)
1. Run `go get .` in this folder to install all dependencies
1. Run `go build .` in this folder to build the executable
1. Run `go install .` to install the executable in the $PATH

## Example 

```bash

# Start the server 
# (TODO: add information about workflow strategies)
./RecommenderServer serve ./testdata/latest-truthy-item-filtered-sorted.nt.gz.schemaTree.typed.bin

# Test with a request 
curl -d '{"properties":["local://prop/Color"],"types":[]}' http://localhost:8080/lean-recommender

```

## Setting up on Cloud VPS 

#### Prerequisites

1. Docker
2. Blubber

#### Cloning the RecommenderServer

```
sudo su
mkdir /etc/docker-compose
cd /etc/docker-compose
git clone https://github.com/martaannaj/RecommenderServer.git recommenderserver
```

#### Building the images

This will build the two images: recommenderserver and recommenderserver-test

```
./blubber-build.sh
```

Or pulling the latest recommenderserver image from Docker Hub

```
pull martaannaj/recommenderserver
```

#### Setup as systemd service

The ```docker-compose.yml``` within this repository assumes that the image used is the one pulled from Docker hub.

```
sudo su
cat << EOF > /etc/systemd/system/docker-compose\@.service
[Unit]
Description=docker-compose %i service
Requires=docker.service network-online.target
After=docker.service network-online.target

[Service]
WorkingDirectory=/etc/docker-compose/%i
Type=simple
TimeoutStartSec=900
Restart=always
ExecStart=/usr/bin/docker-compose up
ExecStop=/usr/bin/docker-compose down

[Install]
WantedBy=multi-user.target
EOF

systemctl enable docker-compose@recommenderserver
systemctl start docker-compose@recommenderserver
```

With the ```docker-compose.yml``` in this repository this should have the service running on port 8771.

 

