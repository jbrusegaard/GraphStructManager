package gsmtypes

import "time"

const (
	LastModified = "lastModified"
	CreatedAt    = "createdAt"
)

type Vertex struct {
	ID           any       `json:"id"           gremlin:"id"`
	LastModified time.Time `json:"lastModified" gremlin:"lastModified"`
	CreatedAt    time.Time `json:"createdAt"    gremlin:"createdAt"`
}

func (v Vertex) GetVertexID() any                 { return v.ID }
func (v Vertex) GetVertexLastModified() time.Time { return v.LastModified }
func (v Vertex) GetVertexCreatedAt() time.Time    { return v.CreatedAt }

type Edge struct {
	ID           any    `json:"id"           gremlin:"id"`
	LastModified string `json:"lastModified" gremlin:"lastModified"`
	CreatedAt    int64  `json:"createdAt"    gremlin:"createdAt"`
}

func (e Edge) GetEdgeID() any              { return e.ID }
func (e Edge) GetEdgeLastModified() string { return e.LastModified }
func (e Edge) GetEdgeCreatedAt() int64     { return e.CreatedAt }
