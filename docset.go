package badger

type docSet map[uintptr]struct{}

func Union(a, b docSet) docSet {
	out := docSet{}
	var id uintptr
	for id, _ = range a {
		out[id] = struct{}{}
	}
	for id, _ = range b {
		out[id] = struct{}{}
	}
	return out
}

func Intersect(a, b docSet) docSet {
	out := docSet{}
	var id uintptr
	for id, _ = range a {
		if _, got := b[id]; got {
			out[id] = struct{}{}
		}
	}
	return out
}
