package types

import (
	"testing"
)

func TestTypes(t *testing.T) {
	vertex := Vertex{
		Id: "1",
	}
	if vertex.Id != "1" {
		t.Errorf("Vertex ID should be 1, got %s", vertex.Id)
	}
	if vertex.LastModified != 0 {
		t.Errorf("Vertex LastModified should be 0, got %d", vertex.LastModified)
	}
	if vertex.GetVertexId() != "1" {
		t.Errorf("Vertex ID should be 1, got %s", vertex.GetVertexId())
	}
	if vertex.GetVertexLastModified() != 0 {
		t.Errorf("Vertex LastModified should be 0, got %d", vertex.GetVertexLastModified())
	}
}
