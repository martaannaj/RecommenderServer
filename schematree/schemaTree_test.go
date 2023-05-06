package schematree

import (
	"testing"

	"log"
	"os"

	"github.com/stretchr/testify/assert"
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
			log.Printf("Encountered error while trying to open the file: %v\n", err)
			log.Panic(err)
		}
		tree, err := Load(f, false)
		assert.NoError(t, err, "An error occured restoring the schematree.")
		assert.EqualValues(t, 1497, tree.PropMap.Len())
		assert.EqualValues(t, 1, tree.MinSup)
		assert.True(t, tree.Typed)
		assert.Len(t, tree.AllProperties(), 1497)
	})
	t.Run("UnTypedSchemaTree", func(t *testing.T) {
		f, err := os.Open(treePath)
		if err != nil {
			log.Printf("Encountered error while trying to open the file: %v\n", err)
			log.Panic(err)
		}
		tree, err := Load(f, false)
		assert.NoError(t, err, "An error occured restoring the schematree.")
		assert.EqualValues(t, 1242, tree.PropMap.Len())
		assert.EqualValues(t, 1, tree.MinSup)
		assert.False(t, tree.Typed)
		assert.Len(t, tree.AllProperties(), 1242)
	})

}

func TestSaveLoadProtocolBuffer(t *testing.T) {

	t.Run("Save large schematree with protocol buffers and load back", func(t *testing.T) {
		original_input_file, err := os.Open(treePathTyped)
		if err != nil {
			log.Printf("Encountered error while trying to open the file: %v\n", err)
			log.Panic(err)
		}
		original_tree, _ := Load(original_input_file, false)
		// store
		proto_file, err := os.CreateTemp("", "schemaTree_test_protocol_buffer")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(proto_file.Name())

		err = original_tree.SaveProtocolBuffer(proto_file)
		if err != nil {
			log.Panicf("Encountered error while saving protocol buffer the file: %v\n", err)
		}
		// load back
		_, err = proto_file.Seek(0, 0)

		if err != nil {
			log.Printf("Encountered error while seeking in the file: %v\n", proto_file)
			log.Panic(err)
		}
		new_tree, _ := LoadProtocolBufferFromReader(proto_file)

		assert.EqualValues(t, original_tree.PropMap.Len(), new_tree.PropMap.Len())
		for k, v := range original_tree.PropMap.prop {
			other_v := new_tree.PropMap.prop[k]
			assert.EqualValues(t, v.Str, other_v.Str, "Propmap items do not contain the same strings")
			assert.EqualValues(t, v.TotalCount, other_v.TotalCount, "Propmap items do not preserve Totalcount")
			assert.EqualValues(t, v.SortOrder, other_v.SortOrder, "Propmap items do not preserve the SortOrder")
		}
		// we already know the maps have the same lengths, so no need to check that there are no extra items in the PropMap

		assert.EqualValues(t, original_tree.MinSup, new_tree.MinSup)
		assert.EqualValues(t, original_tree.Typed, new_tree.Typed)

		// Now compare the complete tree
		depthFirstCompare(t, &original_tree.Root, &new_tree.Root)

	})

}

// SchemaNode is a nodes of the Schema FP-Tree
//
//	type SchemaNode struct {
//		ID            *IItem
//		parent        *SchemaNode
//		FirstChildren [firstChildren]*SchemaNode
//		Children      []*SchemaNode
//		nextSameID    *SchemaNode // node traversal pointer
//		Support       uint32      // total frequency of the node in the path
//	}
func depthFirstCompare(t *testing.T, left *SchemaNode, right *SchemaNode) {
	assert.EqualValues(t, left.ID.Str, right.ID.Str)
	if left.nextSameID == nil {
		assert.True(t, right.nextSameID == nil)
	} else {
		assert.EqualValues(t, left.nextSameID.ID.Str, right.nextSameID.ID.Str)
	}
	assert.EqualValues(t, left.Support, right.Support)

	if left.parent == nil {
		assert.True(t, right.parent == nil)
	} else {
		assert.EqualValues(t, left.parent.ID.Str, right.parent.ID.Str)
	}
	leftChildren := left.Children
	rightChildren := right.Children
	assert.EqualValues(t, len(leftChildren), len(rightChildren), "Unequal number of children, got %d and %d", len(leftChildren), len(rightChildren))
	for i, leftChild := range leftChildren {
		rightChild := rightChildren[i]
		depthFirstCompare(t, leftChild, rightChild)
	}
}
