package tui

import "slices"

// filterBy returns the items that match filter, or all of them when it is empty.
func filterBy[T any](items []T, filter string, match func(T, string) bool) []T {
	if filter == "" {
		return items
	}
	var filtered []T
	for _, item := range items {
		if match(item, filter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// follow returns the index of the first item that is reports true for, or cursor when there is none.
func follow[T any](items []T, cursor int, is func(T) bool) int {
	if i := slices.IndexFunc(items, is); i >= 0 {
		return i
	}
	return cursor
}

// prune drops the selected keys that are no longer listed, so a selection only ever counts what
// an action on it would reach.
func prune[K comparable, T any](selected map[K]bool, items []T, keyOf func(T) K) {
	if len(selected) == 0 {
		return
	}
	listed := make(map[K]bool, len(items))
	for _, item := range items {
		listed[keyOf(item)] = true
	}
	for k := range selected {
		if !listed[k] {
			delete(selected, k)
		}
	}
}

// selectedOr returns the keys of the selected items in list order, or the key of the item under
// the cursor when nothing is selected.
func selectedOr[K comparable, T any](selected map[K]bool, items []T, keyOf func(T) K, current func() (T, bool)) []K {
	if len(selected) > 0 {
		var keys []K
		for _, item := range items {
			if k := keyOf(item); selected[k] {
				keys = append(keys, k)
			}
		}
		return keys
	}
	if item, ok := current(); ok {
		return []K{keyOf(item)}
	}
	return nil
}

// toggle adds key to the selection, or removes it when it is already there.
func toggle[K comparable](selected map[K]bool, key K) {
	if selected[key] {
		delete(selected, key)
	} else {
		selected[key] = true
	}
}

// deleteNames returns items without the ones whose key is in keys.
func deleteNames[K comparable, T any](items []T, keys []K, keyOf func(T) K) []T {
	if len(keys) == 0 {
		return items
	}
	return slices.DeleteFunc(slices.Clone(items), func(item T) bool { return slices.Contains(keys, keyOf(item)) })
}
