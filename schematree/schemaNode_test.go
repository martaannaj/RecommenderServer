package schematree

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestTimeConsuming(t *testing.T) {
//     if testing.Short() {
//         t.Skip("skipping test in short mode.")
//     }
//     ...
// }

func testPropertyMap() propMap {
	return NewPropMap()
}

func emptyRootNodeTest(t *testing.T, root SchemaNode) {

	assert.NotNil(t, root.ID, "schemaNode ID is nil")
	assert.Equal(t, "root", *root.ID.Str, "iri of root node is not \"root\"")
	assert.Nil(t, root.parent, "parent of root not nil")
	assert.Equal(t, 0, len(root.AllChildren), "root node should have a constant number of first children")
}

func TestNewRootNode(t *testing.T) {
	root := newRootNode(testPropertyMap())
	emptyRootNodeTest(t, root)
}

func TestIncrementSupport(t *testing.T) {
	p := testPropertyMap().Get_or_create("root")
	node := SchemaNode{p, nil, []*SchemaNode{}, nil, 0}
	assert.Equal(t, uint32(0), node.Support)
	atomic.AddUint32(&node.Support, 1)
	assert.Equal(t, uint32(1), node.Support)
	atomic.AddUint32(&node.Support, 3)
	assert.Equal(t, uint32(4), node.Support)
}
