package domain

type LabelKind string

const (
	LabelKindContext LabelKind = "Context"
	LabelKindTag     LabelKind = "Tag"
)

type Label struct {
	Value      string
	Kind       LabelKind
	Filterable bool
}
