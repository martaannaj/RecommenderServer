package schematree

// Instance - An assessment on properties

type Instance struct {
	Props                 IList
	tree                  *SchemaTree
	useOptimisticCache    bool // using cache will make an optimistic assumption that `props` are not altered
	cachedRecommendations PropertyRecommendations
}

// NewInstance : constructor method
func NewInstance(argProps IList, argTree *SchemaTree, argUseCache bool) *Instance {
	return &Instance{
		Props:                 argProps,
		tree:                  argTree,
		useOptimisticCache:    argUseCache,
		cachedRecommendations: nil,
	}
}

// NewInstanceFromInput : constructor method to receive strings and convert them into the current
// assessment format that uses IList.
func NewInstanceFromInput(argProps []string, argTypes []string, argTree *SchemaTree, argUseCache bool) *Instance {
	propList := argTree.BuildPropertyList(argProps, argTypes)

	return &Instance{
		Props:                 propList,
		tree:                  argTree,
		useOptimisticCache:    argUseCache,
		cachedRecommendations: nil,
	}
}

// CalcRecommendations : Will execute the core schematree recommender on the properties and return
// the list of recommendations. Cache-enabled operation.
func (inst *Instance) CalcRecommendations() PropertyRecommendations {
	if inst.useOptimisticCache {
		if inst.cachedRecommendations == nil {
			inst.cachedRecommendations = inst.tree.RecommendProperty(inst.Props)
		}
		return inst.cachedRecommendations
	}
	return inst.tree.RecommendProperty(inst.Props)
}
