package schematree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"fmt"
	"os"
	"log"
)

var treePath = "../testdata/10M.nt.gz.schemaTree.bin"
var treePathTyped = "../testdata/10M.nt.gz.schemaTree.typed.bin"

func TestSchemaTree(t *testing.T) {
	tree := New(false, 0)
	t.Run("Root is a proper empty root node", func(t *testing.T) { emptyRootNodeTest(t, tree.Root) })

}

func TestLoad(t *testing.T) {

	t.Run("TypedSchemaTree", func(t *testing.T) {
		f, err := os.Open(treePathTyped)
		if err != nil {
			fmt.Printf("Encountered error while trying to open the file: %v\n", err)
			log.Panic(err)
		}
		tree, _ := Load(f, false)
		assert.EqualValues(t, 1497, len(tree.PropMap))
		assert.EqualValues(t, 1, tree.MinSup)
		assert.True(t, tree.Typed)

	})
	t.Run("UnTypedSchemaTree", func(t *testing.T) {
		f, err := os.Open(treePath)
		if err != nil {
			fmt.Printf("Encountered error while trying to open the file: %v\n", err)
			log.Panic(err)
		}
		tree, _ := Load(f, false)
		assert.EqualValues(t, 1242, len(tree.PropMap))
		assert.EqualValues(t, 1, tree.MinSup)
		assert.False(t, tree.Typed)
	})

}