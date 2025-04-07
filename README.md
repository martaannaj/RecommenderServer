[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

This is the implementation of a Property Recommender Server intended for wikibase. 
It is a newer implementation and can also be used as a general maximum likelihood recommeder as described in  [Gleim L.C. et al. (2020) SchemaTree: Maximum-Likelihood Property Recommendation for Wikidata. 
In: Harth A. et al. (eds) The Semantic Web. ESWC 2020. Lecture Notes in Computer Science, vol 12123. Springer, Cham](https://doi.org/10.1007/978-3-030-49461-2_11).

If you find any security vulnerabilities in this project, please read [SECURITY.md](SECURITY.md) for instructions.
Other requests, such as feature requests, or suggestions can be filed on github. We are also happy receiving pull requests. If your request is regarding the paper, or something related to the old index format, please file your issue at [https://github.com/lgleim/SchemaTreeRecommender](https://github.com/lgleim/SchemaTreeRecommender). 


## Installation

1. Install the go runtime (and VS Code + Golang tools), curently we only maintain for go v1.23
1. Run `go get .` in this folder to install all dependencies
1. Run `go build .` in this folder to build the executable
1. Optionally run `go install .` to install the executable in the $PATH

## Overview

This systems works as follows

1. An index, the schematree, is created from a datasource. This source can be provided in two forms, either a wikidata dump (json.bz2) or a tsv file (.tsv) with on each line a transaction.
2. With this index, recommendations can be created. These can be served using the included HTTP server.
3. When the recommendations are not good enough (as determined using the configurable workflow), several backoff strategies can be applied to improve them. The default is the one which was deemed best after extensive experiemantation in the abovementioned paper.

Since creating the index from a file can take a significant amount of time, it can be serialized to disk in protocol buffer format. 

## Examples
### Index creation ###

The index is created in the same folder where the input data resides.

Create an index from the test data example

```bash
cp testdata/P97.tsv .
./RecommenderServer build-tree from-tsv P97.tsv
```

Create an index using a wikidata dump (download from https://dumps.wikimedia.org/wikidatawiki/entities/)

```bash
./RecommenderServer build-tree from-dump latest-all.json.bz2 
```

### Starting the server ###

Start the server (for testing purposes)
```bash
./RecommenderServer serve P97.tsv.schemaTree.typed.pb
```
Start the server with certificates, on port 8000
```bash
./RecommenderServer serve --cert location_of_TLS_cert --key location_of_private_TLS_key --port 8000 latest-all.json.bz2.schemaTree.typed.pb
```
Start the server with a [custom workflow](configuration/README.md)
```bash
./RecommenderServer serve --workflow workflow_config_file P97.tsv.schemaTree.typed.pb
```

Test with a request. This request means that the entity already has the properties P1319 and P155. The response includes the recommended properties.
```bash
curl -d '{"properties":["P1319", "P155"],"types":[]}' http://localhost:8080/recommender
```

Response:
```json
{"recommendations":[
{"property":"P156","probability":0.4444444444444444},{"property":"P155","probability":0.4444444444444444},{"property":"P582","probability":0.4444444444444444},{"property":"P642","probability":0.2222222222222222},{"property":"P1326","probability":0.1111111111111111},{"property":"P1534","probability":0.1111111111111111},{"property":"P580","probability":0.1111111111111111},{"property":"P276","probability":0.1111111111111111},{"property":"P8555","probability":0.1111111111111111}
]}
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


## using TLS

In an untrusted environment, one shall use TLS to protect the network traffic to and from this recommender server. To use TLS, use the cert and key options to specify the location of the certificate file and the location of the private key file.

semgrep raises a warning for the fact that this server has the option to run without TLS. This warning has been suppressed.

## Compiling the protocol buffer files

The serialization uses protocol buffers. The files needed are precompiled in the repository.
If changes are made, you can recompile. As follows:

1. get the protoc binary from https://github.com/protocolbuffers/protobuf
2. install the go specific generator ```go install google.golang.org/protobuf/cmd/protoc-gen-go@latest```, make sure it is on your path.
3. Compile using 

```bash
protoc  --go_out=.  schematree/serialization/schematree.proto
```
