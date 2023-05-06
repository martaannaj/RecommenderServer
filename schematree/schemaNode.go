package schematree

import (
	"RecommenderServer/schematree/serialization"
	"encoding/gob"
	"log"
	"math"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// SchemaNode is a nodes of the Schema FP-Tree
type SchemaNode struct {
	ID         *IItem
	parent     *SchemaNode
	Children   []*SchemaNode
	nextSameID *SchemaNode // node traversal pointer
	Support    uint32      // total frequency of the node in the path
}

// newRootNode creates a new root node for a given propMap
// The node has its own IItem, but the traversalPointer for the IItem is not yet set
func newRootNode() SchemaNode {
	uuid, err := uuid.NewRandom()
	if err != nil {
		log.Panic("failed to create a random UUID", err)
	}
	rootID := "root" + uuid.String()
	rootItem := &IItem{&rootID, 0, math.MaxUint32, nil}

	rootNode := SchemaNode{rootItem, nil, []*SchemaNode{}, nil, 0}
	return rootNode
}

const lockPrime = 97 // arbitrary prime number
var globalItemLocks [lockPrime]*sync.Mutex
var globalNodeLocks [lockPrime]*sync.RWMutex

// incrementSupport increments the support of the schema node by one
func (node *SchemaNode) incrementSupport() {
	atomic.AddUint32(&node.Support, 1)
}

// convert the SchemaNode into a Protocolbuffer schemanode
func (node *SchemaNode) AsProtoSchemaNode() *serialization.SchemaNode {

	pb_node := serialization.SchemaNode{
		SortOrder: node.ID.SortOrder,
		Support:   node.Support,
		Children:  make([]*serialization.SchemaNode, 0, len(node.Children)),
	}

	// Children
	for _, child := range node.Children {
		pb_child := child.AsProtoSchemaNode()
		pb_node.Children = append(pb_node.Children, pb_child)
	}

	return &pb_node
}

func FromProtoSchemaNode(pb_node *serialization.SchemaNode, props []*IItem) *SchemaNode {
	// function scoping to allow for garbage collection
	// err := func() error {
	// ID

	node := &SchemaNode{}
	node.ID = props[pb_node.SortOrder]

	// traversal pointer repopulation
	node.nextSameID = node.ID.traversalPointer
	node.ID.traversalPointer = node

	// Support
	node.Support = pb_node.Support

	// Children
	for _, pb_child := range pb_node.Children {
		child := FromProtoSchemaNode(pb_child, props)
		child.parent = node
		node.Children = append(node.Children, child)
	}

	return node
}

// decodeGob decodes the schema node from its binary representation
func (node *SchemaNode) decodeGob(d *gob.Decoder, props []*IItem) error {
	// function scoping to allow for garbage collection
	// err := func() error {
	// ID
	var id uint32
	err := d.Decode(&id)
	if err != nil {
		return err
	}
	node.ID = props[int(id)]

	// traversal pointer repopulation
	node.nextSameID = node.ID.traversalPointer
	node.ID.traversalPointer = node

	// Support
	err = d.Decode(&node.Support)
	if err != nil {
		return err
	}

	// Children
	var length int
	err = d.Decode(&length)
	if err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		child := &SchemaNode{nil, node, nil, nil, 0}
		err = child.decodeGob(d, props)
		if err != nil {
			return err
		}
		node.Children = append(node.Children, child)
	}

	return nil
}

// prefixContains checks if all properties of a given list are ancestors of a node
// internal! propertyPath *MUST* be sorted in sortOrder (i.e. descending support)
// thread-safe!
func (node *SchemaNode) prefixContains(propertyPath IList) bool {
	nextP := len(propertyPath) - 1                         // index of property expected to be seen next
	for cur := node; cur.parent != nil; cur = cur.parent { // walk from leaf towards root

		if cur.ID.SortOrder < propertyPath[nextP].SortOrder { // we already walked past the next expected property
			return false
		}
		if cur.ID == propertyPath[nextP] {
			nextP--
			if nextP < 0 { // we encountered all expected properties!
				return true
			}
		}
	}
	return false
}
