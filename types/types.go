package types

import "errors"

type IdType interface {
	string
	int
}

type ResultInterface interface {
	GetId() string
	GetLastModified() string
}

type Vertex struct {
	id           string
	lastModified int64 `gremlin:"lastModified"`
}

func (v *Vertex) GetId() string {
	return v.id
}

func (v *Vertex) GetLastModified() int64 {
	return v.lastModified
}

func (v *Vertex) SetId(id string) error {
	if v.id != "" {
		return errors.New("id already set")
	}
	v.id = id
	return nil
}

func (v *Vertex) SetLastModified(lastModified int64) {
	v.lastModified = lastModified
}

type Edge struct {
	Id           string `gremlin:"id"`
	LastModified string `gremlin:"lastModified"`
}

func (e *Edge) GetId() string {
	return e.Id
}

func (e *Edge) GetLastModified() string {
	return e.LastModified
}
