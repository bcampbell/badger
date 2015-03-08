package badger

type DocSet map[DocID]struct{}

func Union(a, b DocSet) DocSet {
	out := DocSet{}
	var id DocID
	for id, _ = range a {
		out[id] = struct{}{}
	}
	for id, _ = range b {
		out[id] = struct{}{}
	}
	return out
}

func Intersect(a, b DocSet) DocSet {
	out := DocSet{}
	var id DocID
	for id, _ = range a {
		if _, got := b[id]; got {
			out[id] = struct{}{}
		}
	}
	return out
}

// Subtract removes all members of b from a
func (a DocSet) Subtract(b DocSet) {
	var id DocID
	for id, _ = range b {
		delete(a, id)
	}
}
