package domain

import "errors"

var (
	ErrSystemTagModification = errors.New("system tags cannot be modified or deleted")
	ErrTagInvalid            = errors.New("invalid arguments")
)

var (
	SystemTags = map[string]struct{}{
		"отчетность": {},
		"операции":   {},
		"звонок":     {},
	}
)

type Tag struct {
	Name     string
	IsSystem bool
}

func NewTag(name string) (Tag, error) {
	_, isSys := SystemTags[name]

	return Tag{
		Name:     name,
		IsSystem: isSys,
	}, nil
}

func (t Tag) CanDelete() error {
	if t.IsSystem {
		return ErrSystemTagModification
	}
	return nil
}
