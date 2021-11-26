package schematree

import (
	"encoding/gob"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	gzip "github.com/klauspost/pgzip"
)

// TypedSchemaTree is a schematree that includes type information as property nodes
type SchemaTree struct {
	PropMap propMap    // PropMap maps the string representations of properties to the corresponding IItem
	Root    SchemaNode // Root is the root node of the schematree. All further nodes are descendants of this node.
	MinSup  uint32     // TODO (not used)
	Typed   bool       // Typed indicates if this schematree includes type information as properties
}

// New returns a newly allocated and initialized schema tree
func New(typed bool, minSup uint32) (tree *SchemaTree) {
	if minSup < 1 {
		minSup = 1
	}

	pMap := make(propMap)
	tree = &SchemaTree{
		PropMap: pMap,
		Root:    newRootNode(pMap),
		MinSup:  minSup,
		Typed:   typed,
	}
	tree.init()
	return
}

// Init initializes the datastructure for usage
func (tree *SchemaTree) init() {
	for i := range globalItemLocks {
		globalItemLocks[i] = &sync.Mutex{}
		globalNodeLocks[i] = &sync.RWMutex{}
	}
}

// Load loads a binarized SchemaTree from disk
func Load(f io.Reader, stripURI bool) (*SchemaTree, error) {

	// func Load(filePath string, stripURI bool) (*SchemaTree, error) {
	// Alternatively via GobDecoder(...): https://stackoverflow.com/a/12854659

	t1 := time.Now()

	r, err := gzip.NewReader(f)
	if err != nil {
		fmt.Printf("Encountered error while trying to decompress the file: %v\n", err)
		return nil, err
	}
	defer r.Close()

	/// decoding
	tree := New(false, 1)
	d := gob.NewDecoder(r)

	// decode propMap
	var props []*IItem
	err = d.Decode(&props)
	if err != nil {
		return nil, err
	}

	if stripURI {
		for i, prop := range props {
			uri := strings.Split(*prop.Str, "/")
			if strings.HasPrefix(*prop.Str, "t#") {
				typeValue := "t#" + uri[len(uri)-1]
				props[i].Str = &typeValue
			} else {
				props[i].Str = &uri[len(uri)-1]
			}
		}
	}

	for sortOrder, item := range props {
		item.SortOrder = uint32(sortOrder)
		tree.PropMap[*item.Str] = item
	}
	fmt.Printf("%v properties... ", len(props))

	// decode MinSup
	err = d.Decode(&tree.MinSup)
	if err != nil {
		return nil, err
	}

	// decode Root
	fmt.Printf("decoding tree...")
	err = tree.Root.decodeGob(d, props)

	if err != nil {
		return nil, err
	}

	// legacy import bug workaround
	if *tree.Root.ID.Str != "root" {
		fmt.Println("WARNING!!! Encountered legacy root node import bug - root node counts will be incorrect!")
		tree.Root.ID = tree.PropMap.get("root")
	}

	//decode Typed
	var i int
	err = d.Decode(&i)
	if i == 1 {
		tree.Typed = true
	}
	if err != nil {
		return nil, err
	}

	if err != nil {
		fmt.Printf("Encountered error while decoding the file: %v\n", err)
		return nil, err
	}

	fmt.Println(time.Since(t1))
	return tree, err
}

// Support returns the total cooccurrence-frequency of the given property list
func (tree *SchemaTree) Support(properties IList) uint32 {
	var support uint32

	if len(properties) == 0 {
		return tree.Root.Support // empty set occured in all transactions
	}

	properties.Sort() // descending by support

	// check all branches that include least frequent term
	for term := properties[len(properties)-1].traversalPointer; term != nil; term = term.nextSameID {
		if term.prefixContains(properties) {
			support += term.Support
		}
	}

	return support
}
