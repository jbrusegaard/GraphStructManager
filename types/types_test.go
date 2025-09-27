package types

import (
	"testing"
	"time"
)

func TestTypes(t *testing.T) {
	t.Parallel()
	vertex := Vertex{
		Id: "1",
	}
	if vertex.Id != "1" {
		t.Errorf("Vertex ID should be 1, got %s", vertex.Id)
	}

	if vertex.CreatedAt != (time.Time{}) {
		t.Errorf("Vertex CreatedAt should be 0, got %v", vertex.CreatedAt)
	}
	if vertex.GetVertexCreatedAt() != (time.Time{}) {
		t.Errorf("Vertex CreatedAt should be 0, got %v", vertex.GetVertexCreatedAt())
	}
	if vertex.GetVertexId() != "1" {
		t.Errorf("Vertex ID should be 1, got %s", vertex.GetVertexId())
	}
	if vertex.GetVertexLastModified() != (time.Time{}) {
		t.Errorf("Vertex LastModified should be 0, got %v", vertex.GetVertexLastModified())
	}
	if vertex.LastModified != (time.Time{}) {
		t.Errorf("Vertex LastModified should be 0, got %v", vertex.LastModified)
	}
}
