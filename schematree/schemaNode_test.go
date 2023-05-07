package schematree

import (
	"strings"
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
	assert.True(t, strings.HasPrefix(*root.ID.Str, "root"), "ID of root node does not start with \"root\"")
	assert.Nil(t, root.parent, "parent of root not nil")
	assert.Equal(t, 0, len(root.Children), "root node should have a constant number of first children")

}

func TestNewRootNode(t *testing.T) {
	root := newRootNode()
	emptyRootNodeTest(t, root)
	// When not in a tree yet, the traversalPinter must be nil
	assert.Nil(t, root.ID.traversalPointer)
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
