package testdata

type Item struct {
	Value int
}

func ptrEscape() *Item {
	item := Item{Value: 42}
	return &item
}
