package driver

type VertexType interface {
	GetVertexId() any
	GetVertexLastModified() int64
}

// validateIsVertexStruct validates that the parameter v has Vertex as an anonymous struct
// func validateIsVertexStruct(v any) error {
// 	rv := reflect.ValueOf(v)
//
// 	// Handle pointer types by getting the underlying value
// 	if rv.Kind() == reflect.Ptr {
// 		if rv.IsNil() {
// 			return fmt.Errorf("value cannot be nil")
// 		}
// 		rv = rv.Elem()
// 	}
//
// 	// Check if it's a struct
// 	if rv.Kind() != reflect.Struct {
// 		return fmt.Errorf("value must be a struct, got %s", rv.Kind())
// 	}
//
// 	// Get the struct type
// 	rt := rv.Type()
//
// 	// Check for anonymous Vertex field
// 	for i := 0; i < rt.NumField(); i++ {
// 		field := rt.Field(i)
//
// 		// Check if this field is anonymous and is of type types.Vertex
// 		if field.Anonymous && field.Type == reflect.TypeOf(types.Vertex{}) {
// 			return nil
// 		}
// 	}
//
// 	return fmt.Errorf("struct must contain anonymous types.Vertex field")
// }
