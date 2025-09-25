package types

type Vertex struct {
	Id           any   `json:"id"           gremlin:"id"`
	LastModified int64 `json:"lastModified" gremlin:"lastModified"`
	CreatedAt    int64 `json:"createdAt"    gremlin:"createdAt"`
}

func (v Vertex) GetVertexId() any             { return v.Id }
func (v Vertex) GetVertexLastModified() int64 { return v.LastModified }
func (v Vertex) GetVertexCreatedAt() int64    { return v.CreatedAt }

type Edge struct {
	Id           any    `json:"id"           gremlin:"id"`
	LastModified string `json:"lastModified" gremlin:"lastModified"`
	CreatedAt    int64  `json:"createdAt"    gremlin:"createdAt"`
}

func (e Edge) GetEdgeId() any              { return e.Id }
func (e Edge) GetEdgeLastModified() string { return e.LastModified }
func (e Edge) GetEdgeCreatedAt() int64     { return e.CreatedAt }
