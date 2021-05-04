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
./RecommenderServer serve ./testdata/handcrafted-item-filtered-sorted.schemaTree.typed.bin

# Test with a request 
curl -d '{"properties":["local://prop/Color"],"types":[]}' http://localhost:8080/lean-recommender

```

### Performance Evaluation Details

| Dataset | Results |
| ------ | ------ |
| Wikidata | [here](evaluation/visualization_single_evaluation_wiki.ipynb) |
| LOD-a-lot | [here](evaluation/visualization_single_evaluation-LOD.ipynb) |
| Backoff strategies | [here](evaluation/visualization_batch.ipynb) |
 

