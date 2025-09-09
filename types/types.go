package types

type Vertex struct {
	Id           any   `json:"id"`
	LastModified int64 `json:"lastModified"`
}

type Edge struct {
	Id           any    `json:"id"`
	LastModified string `json:"lastModified"`
}
