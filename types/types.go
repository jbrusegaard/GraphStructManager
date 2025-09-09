package types

type Vertex struct {
	Id           any
	LastModified int64 `json:"lastModified" gremlin:"lastModified"`
}

type Edge struct {
	Id           any
	LastModified string `json:"lastModified" gremlin:"lastModified"`
}
