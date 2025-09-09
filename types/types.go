package types

type Vertex struct {
	Id           any   `json:"id"           gremlin:"id"`
	LastModified int64 `json:"lastModified" gremlin:"lastModified"`
}

type Edge struct {
	Id           any    `json:"id"           gremlin:"id"`
	LastModified string `json:"lastModified" gremlin:"lastModified"`
}
