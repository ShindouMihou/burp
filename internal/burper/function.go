package burper

type Function struct {
	Name        string
	Transformer Transformer
}

type Transformer = func(call *Call, tree *Tree) ([]byte, error)
