package domain

import "errors"

var (
	ErrSystemTagModification = errors.New("system tags cannot be modified or deleted")
	ErrTagEmpty              = errors.New("tag name cannot be empty")
)

var SystemTags = map[string]bool{
	"отчетность": true,
	"операции":   true,
	"звонок":     true,
}

type Tag struct {
	Name     string
	IsSystem bool
}

func NewTag(name string) (Tag, error) {
	if name == "" {
		return Tag{}, ErrTagEmpty
	}
	return Tag{
		Name:     name,
		IsSystem: SystemTags[name],
	}, nil
}

func (t Tag) CanDelete() error {
	if t.IsSystem {
		return ErrSystemTagModification
	}
	return nil
}
