package badger

type docSet map[string]struct{}

func Union(a, b docSet) docSet {
	out := docSet{}
	var id string
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
	var id string
	for id, _ = range a {
		if _, got := b[id]; got {
			out[id] = struct{}{}
		}
	}
	return out
}
