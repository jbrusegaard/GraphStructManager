package types

type Vertex struct {
	Id           any   `json:"id"           gremlin:"id"`
	LastModified int64 `json:"lastModified" gremlin:"lastModified"`
}

func (v Vertex) GetVertexId() any             { return v.Id }
func (v Vertex) GetVertexLastModified() int64 { return v.LastModified }

type Edge struct {
	Id           any    `json:"id"           gremlin:"id"`
	LastModified string `json:"lastModified" gremlin:"lastModified"`
}
