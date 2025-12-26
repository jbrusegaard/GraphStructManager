package gsmtypes

import (
	"testing"
	"time"
)

func TestTypes(t *testing.T) {
	t.Parallel()
	vertex := Vertex{
		ID: "1",
	}
	if vertex.ID != "1" {
		t.Errorf("Vertex ID should be 1, got %s", vertex.ID)
	}

	if vertex.CreatedAt != (time.Time{}) {
		t.Errorf("Vertex CreatedAt should be 0, got %v", vertex.CreatedAt)
	}
	if vertex.GetVertexCreatedAt() != (time.Time{}) {
		t.Errorf("Vertex CreatedAt should be 0, got %v", vertex.GetVertexCreatedAt())
	}
	if vertex.GetVertexID() != "1" {
		t.Errorf("Vertex ID should be 1, got %s", vertex.GetVertexID())
	}
	if vertex.GetVertexLastModified() != (time.Time{}) {
		t.Errorf("Vertex LastModified should be 0, got %v", vertex.GetVertexLastModified())
	}
	if vertex.LastModified != (time.Time{}) {
		t.Errorf("Vertex LastModified should be 0, got %v", vertex.LastModified)
	}
}
