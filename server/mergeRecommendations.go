/*
Implementation of strategies to merge two lists of recommendations, where the one take precendence of the other.
*/
package server

import (
	"log"
	"sort"
)

type WithID interface {
	getID() interface{}
}

/*
Assumptions:
1. lowerPriority has all properties
2. arguments will not be modified
*/
func SimpleMerge[T WithID](higherPriority, lowerPriority []T) []T {
	scratch_copy := make([]T, 0, len(lowerPriority))
	scratch_copy = append(scratch_copy, lowerPriority...)

	for _, rh := range higherPriority {
		for idl, rl := range scratch_copy {
			if rh.getID() == rl.getID() {
				copy(scratch_copy[idl:], scratch_copy[idl+1:])
				scratch_copy = scratch_copy[:len(scratch_copy)-1]
				//break out of the inner range loop, which is essential because the slice has changed
				break
			}
		}
	}
	//now, we can move the lower priority recommendations to the back of its array, and put the higher priority ones to the front
	// First claim back the full capacity
	scratch_copy = scratch_copy[:cap(scratch_copy)]
	copy(scratch_copy[len(higherPriority):], scratch_copy) // This ignores what goes beyond scratch_copy[len(higherPriority)], which is the intention
	copy(scratch_copy, higherPriority)
	return scratch_copy
}

/*
Assumptions (in addition to SimpleMerge)
3. higherPriorty and lowerPriority are sorted more or less in the same order. This allows faster matching.
The output must be exactly the same as SimpleMerge
*/
func FasterMerge[T WithID](higherPriority, lowerPriority []T) []T {
	if len(higherPriority) == 0 {
		return lowerPriority
	}
	hitlist := make([]int, 0, len(higherPriority))
	// Here we make use of the fact that things are usually ordered the same way. This startegy should reduce the number of copies significantly
	// We iterate over higherpriority and try to find the entries one by one in the other list.
	// we go over copy1, until we find the current entry in higherPriority.
	// We record the index at which a hit was found.
	// Then we move the current_entry_idx, but when we go over copy1, we continue from the recorded index + 1, because the entry is likely after this.
	// If we reach the end of copy1, it means the current entry must have been above, so we restart from the top.
	//
	// Once all are found, we sort the hit list, and copy the parts 0-first hit, between the hits, and last-end, directly to the end of the output.
	// Finally we copy the higherpriority entries to the beginning.

	low_priority_index := 0
HighPriorityLoop:
	for _, current_entry := range higherPriority {
		search_start := low_priority_index
		// last part of copy1
		for ; low_priority_index < len(lowerPriority); low_priority_index++ {
			if current_entry.getID() == lowerPriority[low_priority_index].getID() {
				// We  found a hit
				hitlist = append(hitlist, low_priority_index)
				low_priority_index++
				continue HighPriorityLoop
			}
		}
		// still to continue with the beginnign of the list
		for low_priority_index = 0; low_priority_index < search_start; low_priority_index++ {
			if current_entry.getID() == lowerPriority[low_priority_index].getID() {
				// We  found a hit
				hitlist = append(hitlist, low_priority_index)
				low_priority_index++
				continue HighPriorityLoop
			}
		}
		log.Panic("An entry from the high priority list could not be found in the low priority list, this must never happen")
	}
	// sort the hits
	sort.Slice(hitlist, func(i, j int) bool {
		return hitlist[i] < hitlist[j]
	})
	// copy time
	result := make([]T, 0, len(lowerPriority))
	// high priority at the front
	result = append(result, higherPriority...)
	// now all the chunck around the hits
	// start
	result = append(result, lowerPriority[0:hitlist[0]]...)
	// in between
	for chunck_start_exclusive_idx, chunck_start_exclusive := range hitlist[:len(hitlist)-1] {
		chunk_end_exlusive := hitlist[chunck_start_exclusive_idx+1]
		result = append(result, lowerPriority[chunck_start_exclusive+1:chunk_end_exlusive]...)
	}
	//end
	result = append(result, lowerPriority[hitlist[len(hitlist)-1]+1:]...)
	return result
}
