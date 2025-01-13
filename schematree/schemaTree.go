package schematree

import (
	"RecommenderServer/schematree/serialization"
	"compress/gzip"
	"encoding/gob"
	"io"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"fortio.org/safecast"

	"google.golang.org/protobuf/proto"
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

	pMap := NewPropMap()
	tree = &SchemaTree{
		PropMap: pMap,
		Root:    newRootNode(),
		MinSup:  minSup,
		Typed:   typed,
	}
	tree.Root.ID.traversalPointer = &tree.Root
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

// getOrCreateChild returns the child of a node associated to a IItem. If such child does not exist, a new child is created.
func (node *SchemaNode) getOrCreateChild(term *IItem) *SchemaNode {

	// binary search for the child
	globalNodeLocks[lock_for_node(node)].RLock()
	children := node.Children
	i := sort.Search(
		len(children),
		func(i int) bool {
			// nosemgrep: go.lang.security.audit.unsafe.use-of-unsafe-block
			return uintptr(unsafe.Pointer(children[i].ID)) >= uintptr(unsafe.Pointer(term)) // #nosec G103 # The unsafe pointers are converted to uintptr and only used to create an ordering. They are never converted back to Pointers.
		})

	if i < len(children) {
		if child := children[i]; child.ID == term {
			globalNodeLocks[lock_for_node(node)].RUnlock()
			return child
		}
	}
	globalNodeLocks[lock_for_node(node)].RUnlock()

	// We have to add the child, aquire a lock for this term
	globalNodeLocks[lock_for_node(node)].Lock()

	// search again, since child might meanwhile have been added by other thread or previous search might have missed
	children = node.Children
	i = sort.Search(
		len(children),
		func(i int) bool {
			// nosemgrep: go.lang.security.audit.unsafe.use-of-unsafe-block
			return uintptr(unsafe.Pointer(children[i].ID)) >= uintptr(unsafe.Pointer(term)) // #nosec G103 # The unsafe pointers are converted to uintptr and only used to create an ordering. They are never converted back to Pointers.
		})
	if i < len(node.Children) {
		if child := children[i]; child.ID == term {
			globalNodeLocks[lock_for_node(node)].Unlock()
			return child
		}
	}

	// child not found, but i is the index where it would be inserted.
	// create a new one...
	globalItemLocks[lock_for_term(term)].Lock()
	newChild := &SchemaNode{term, node, []*SchemaNode{}, term.traversalPointer, 0}
	term.traversalPointer = newChild
	globalItemLocks[lock_for_term(term)].Unlock()

	// ...and insert it at position i
	node.Children = append(node.Children, nil)
	copy(node.Children[i+1:], node.Children[i:])
	node.Children[i] = newChild

	globalNodeLocks[lock_for_node(node)].Unlock()

	return newChild
}

func lock_for_term(item *IItem) uint64 {
	// nosemgrep: go.lang.security.audit.unsafe.use-of-unsafe-block
	val := uint64(uintptr(unsafe.Pointer(item)) % lockPrime) // #nosec G103 # the unsafe pointer is immediately converted into a uintptr and never back to a pointer type.
	return val
}

func lock_for_node(node *SchemaNode) uint64 {
	// nosemgrep: go.lang.security.audit.unsafe.use-of-unsafe-block
	val := uint64(uintptr(unsafe.Pointer(node)) % lockPrime) // #nosec G103 # the unsafe pointer is immediately converted into a uintptr and never back to a pointer type.
	return val
}

// Insert inserts all properties of a new subject into the schematree
// The subject is given by
// thread-safe
func (tree *SchemaTree) Insert(propertySet []*IItem) {

	// transform into iList of properties
	properties := make(IList, len(propertySet))
	copy(properties, propertySet)

	// sort the properties descending by support
	properties.Sort()

	// insert sorted item list into the schemaTree
	node := &tree.Root
	node.incrementSupport()
	for _, prop := range properties {
		node = node.getOrCreateChild(prop) // recurse, i.e., node.getOrCreateChild(prop).insert(properties[1:], types)
		node.incrementSupport()
	}

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

func (tree *SchemaTree) AllProperties() []string {
	all_properties := make([]string, 0)
	for prop := range tree.PropMap.prop {
		all_properties = append(all_properties, prop)
	}
	return all_properties
}

func (tree *SchemaTree) SaveProtocolBuffer(writer io.Writer) error {
	t1 := time.Now()
	//	log.Printf("Writing schema to protocol buffer file %v... ", filePath)

	pb_tree := &serialization.SchemaTree{}

	// encode propMap
	pb_propmap := &serialization.PropMap{}
	// first get them in order
	props := make([]*IItem, tree.PropMap.Len()+1) // one extra for the root node
	for _, p := range tree.PropMap.list_properties() {
		props[int(p.SortOrder)] = p
	}
	// add the root node at the end
	if props[len(props)-1] != nil {
		log.Panic("something is taking the space reserved for the root node!")
	}
	props[len(props)-1] = tree.Root.ID

	//then store all in order
	for _, p := range props {
		pb_propmap_item := &serialization.PropMapItem{
			Str:        *p.Str,
			TotalCount: p.TotalCount,
			SortOrder:  p.SortOrder,
		}
		pb_propmap.Items = append(pb_propmap.Items, pb_propmap_item)
	}

	pb_tree.PropMap = pb_propmap
	// encode MinSup

	pb_tree.MinSup = tree.MinSup

	// encode from root
	var root *serialization.SchemaNode = tree.Root.AsProtoSchemaNode()

	pb_tree.Root = root

	// encode Typed
	if tree.Typed {
		pb_tree.Options = []serialization.Options{serialization.Options_TYPED}
	}
	//else {
	//no action needed. The default is an empty option list, wich is fine
	//}

	out, err := proto.Marshal(pb_tree)
	if err != nil {
		return err
	}
	// TODO check whether gzip compression helps

	nn, err := writer.Write(out)
	if err != nil || nn != len(out) {
		log.Panicf("Could not write all output to the file. Error %s , written %d", err, nn)
	}
	log.Printf("done (%v)\n", time.Since(t1))

	return nil
}

func LoadProtocolBufferFromReader(input io.Reader) (*SchemaTree, error) {
	log.Println("Start loading schema (protocol buffer format)")
	in, err := io.ReadAll(input)
	if err != nil {
		log.Fatalln("Error reading input:", err)
	}
	return loadProtocolBuffer(in)
}

/*
Load a schematree from the protocol buffer representation.
*/
func loadProtocolBuffer(in []byte) (*SchemaTree, error) {
	t1 := time.Now()
	pb_tree := &serialization.SchemaTree{}
	if err := proto.Unmarshal(in, pb_tree); err != nil {
		return nil, err
	}

	tree := New(false, 1)

	// decode propMap
	var props []*IItem

	for _, pb_item := range pb_tree.PropMap.Items[:len(pb_tree.PropMap.Items)-1] { //we do not do the root node
		asiitem := tree.PropMap.Get_or_create(pb_item.Str)
		asiitem.TotalCount = pb_item.TotalCount
		// This sortorder is overwritten just like in the gob implementation, but that seems unnecesary.
		// sortOrder was the index in the items array, but that is already set in the item from Get_or_create anyway
		// item.SortOrder = uint32(sortOrder)
		if asiitem.SortOrder != pb_item.SortOrder {
			log.Panic("The sort order does not seem to be consistent.")
		}
		props = append(props, asiitem)
	}
	log.Printf("%v properties... \n", len(props))

	// decode IItem for root
	pb_root_item := pb_tree.PropMap.Items[len(pb_tree.PropMap.Items)-1]
	rootItem := &IItem{&pb_root_item.Str, pb_root_item.TotalCount, pb_root_item.SortOrder, nil}
	props = append(props, rootItem)

	// decode Root, cheating the root to take its value from the last item in the props
	pb_tree.Root.SortOrder = safecast.MustConvert[uint32](len(pb_tree.PropMap.Items) - 1)
	log.Printf("decoding tree...")
	tree.Root = *FromProtoSchemaNode(pb_tree.Root, props)

	// decode MinSup
	tree.MinSup = pb_tree.MinSup

	//decode Typed
	for _, option := range pb_tree.Options {
		switch option {
		case serialization.Options_TYPED:
			tree.Typed = true
		default:
			log.Fatal("Unknown option in protocol buffer tree")
		}
	}

	log.Println("Time for decoding ", time.Since(t1), " seconds")
	return tree, nil

}

// Load loads a binarized SchemaTree from disk
func Load(f io.Reader, stripURI bool) (*SchemaTree, error) {

	// func Load(filePath string, stripURI bool) (*SchemaTree, error) {
	// Alternatively via GobDecoder(...): https://stackoverflow.com/a/12854659

	t1 := time.Now()

	r, err := gzip.NewReader(f)
	if err != nil {
		log.Printf("Encountered error while trying to decompress the file: %v\n", err)
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
		item.SortOrder = safecast.MustConvert[uint32](sortOrder)
		tree.PropMap.prop[*item.Str] = item
	}
	log.Printf("%v properties... ", len(props))

	// decode MinSup
	err = d.Decode(&tree.MinSup)
	if err != nil {
		return nil, err
	}

	// decode Root
	log.Printf("start decoding tree...")
	err = tree.Root.decodeGob(d, props)

	if err != nil {
		return nil, err
	}

	// legacy import is no longer supported.
	if !strings.HasPrefix(*tree.Root.ID.Str, "root") {
		log.Panicln("The loaded model does not contain a root node. Such models are no longer supported.")
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
		log.Printf("Encountered error while decoding the file: %v\n", err)
		return nil, err
	}

	log.Println("finished decoding tree in ", time.Since(t1))
	return tree, err
}
