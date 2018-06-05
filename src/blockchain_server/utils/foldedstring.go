package utils

import (
	"sort"
	"strings"
)

type FoldedStrings []string

func SortFoldedStrings(slice []string) {
	sort.Sort(FoldedStrings(slice))
}

func (slice FoldedStrings) Len() int { return len(slice) }

func (slice FoldedStrings) Less(i, j int) bool {
	return strings.ToLower(slice[i]) < strings.ToLower(slice[j])
}

func (slice FoldedStrings) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice *FoldedStrings) Sort() {
	sort.Strings(*slice)
}

func (slice *FoldedStrings) Insert(s string) {
	if !sort.StringsAreSorted(*slice) {
		sort.Strings(*slice)
	}
	index := sort.SearchStrings(*slice, s)

	if index>=len(*slice) || (*slice)[index]!=s {
		*slice = append((*slice)[:index], append([]string{s}, (*slice)[index:]...)...)
	}
}

func (slice *FoldedStrings) Contains(s string) bool {
	index := sort.SearchStrings(*slice, s)
	return index>=0 && index<len(*slice) &&
		strings.ToLower((*slice)[index])==strings.ToLower(s)
}
