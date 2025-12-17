package driver

import (
	"testing"

	"github.com/jbrusegaard/graph-struct-manager/comparator"
	"github.com/jbrusegaard/graph-struct-manager/types"
)

type testVertex struct {
	types.Vertex
	Name string `json:"name" gremlin:"name"`
}

const DbUrl = "ws://localhost:8182"

func TestDriverConnections(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "ValidConnection",
			url:     DbUrl,
			wantErr: false,
		},
		{
			name:    "InvalidURL",
			url:     "invalid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()
				db, err := Open(tt.url)
				if (err != nil) != tt.wantErr {
					t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if db != nil {
					defer db.Close()
				}
			},
		)
	}
}

func TestDriverTable(t *testing.T) {
	db, err := Open(DbUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	table := db.Label("TestVertex")
	if table == nil {
		t.Fatal("Table should not be nil")
	}
	if table.label != "TestVertex" {
		t.Errorf("Table label should be TestVertex, got %s", table.label)
	}
}

func TestDriverModel(t *testing.T) {
	db, err := Open(DbUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := Model[testVertex](db)
	if model == nil {
		t.Fatal("Model should not be nil")
	}
}

func TestDriverWhere(t *testing.T) {
	db, err := Open(DbUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := Where[testVertex](db, "name", comparator.EQ, "test")
	if model == nil {
		t.Fatal("Model should not be nil")
	}
}
