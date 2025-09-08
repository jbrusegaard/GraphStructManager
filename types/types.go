package types

type Vertex struct {
	Id           any
	LastModified int64 `gremlin:"lastModified"`
}

type Edge struct {
	Id           any    `gremlin:"id"`
	LastModified string `gremlin:"lastModified"`
}
