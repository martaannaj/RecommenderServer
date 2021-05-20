#!/usr/bin/env bash

BIN="./evaluation"
EVALBASE="../../testdata/latest-truthy-item-filtered-sorted-1in10000" 
COMMON="-results -testSet $EVALBASE-test.nt.gz"

TYPEDBACKOFFTREE="-model $EVALBASE-train.nt.gz.threeNodes.schemaTree.typed.bin -typed -workflow Wiki_backoff.json" # the binary with three pointers in the node


TAKEITER="-handler takeMoreButCommon"
BYNONTYPES="-groupBy numNonTypes"
BYSETSIZE="-groupBy setSize"

cd ..
cd SchemaTreeBuilder
go build .
./SchemaTreeBuilder build-tree-typed ../testdata/latest-truthy-item-filtered-sorted-1in10000-train.nt.gz
cd ..

cd RecommenderServer
go build .
cd evaluation
go build .
$BIN $COMMON $TYPEDBACKOFFTREE $TAKEITER $BYSETSIZE -name two-nodes-typed-tooFewRecs-takeMoreButCommon-setSize

$BIN $COMMON $TYPEDBACKOFFTREE $TAKEITER $BYNONTYPES -name two-nodes-typed-tooFewRecs-takeMoreButCommon-byNonTypes