package driver

import (
	"fmt"
	"maps"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/charmbracelet/log"
	"github.com/jbrusegaard/graph-struct-manager/comparator"
	"github.com/jbrusegaard/graph-struct-manager/gsmtypes"
)

var cardinality = gremlingo.Cardinality

// Query represents a chainable query builder
type Query[T gsmtypes.VertexType] struct {
	db            *GremlinDriver
	conditions    []*QueryCondition
	ids           []any
	label         string
	limit         *int
	offset        *int
	orderBy       *OrderCondition
	dedup         bool
	subTraversals map[string]*gremlingo.GraphTraversal
	debugString   *strings.Builder
}

type QueryCondition struct {
	field     string
	operator  comparator.Comparator
	value     any
	traversal *gremlingo.GraphTraversal
}

func (qc *QueryCondition) String() string {
	if qc.traversal != nil {
		return ""
	}

	if qc.field == "id" {
		return fmt.Sprintf(".HasId(%v)", qc.value)
	}
	var sb strings.Builder
	sb.WriteString(".Has(")
	sb.WriteString(qc.field)
	sb.WriteString(", ")

	switch qc.operator {
	case comparator.EQ, "eq":
		sb.WriteString("P.Eq(")
	case comparator.NEQ, "neq":
		sb.WriteString("P.Neq(")
	case comparator.GT, "gt":
		sb.WriteString("P.Gt(")
	case comparator.GTE, "gte":
		sb.WriteString("P.Gte(")
	case comparator.LT, "lt":
		sb.WriteString("P.Lt(")
	case comparator.LTE, "lte":
		sb.WriteString("P.Lte(")
	case comparator.IN:
		sb.WriteString("P.Within(")
	case comparator.CONTAINS:
		sb.WriteString("TextP.Containing(")
	case comparator.WITHOUT:
		sb.WriteString("P.Without(")
	}

	sb.WriteString(fmt.Sprintf("%v))", qc.value))
	return sb.String()
}

type OrderCondition struct {
	field string
	desc  bool
}

func getLabel[T gsmtypes.VertexType]() (string, error) {
	var v T
	// Use getLabelFromValue to support both pointer and value receivers
	label := getLabelFromVertex(v)
	return label, nil
}

// NewQuery creates a new query builder for type T
func NewQuery[T gsmtypes.VertexType](db *GremlinDriver) *Query[T] {
	label, _ := getLabel[T]()
	queryAsString := strings.Builder{}
	queryAsString.WriteString("V()")
	if label != "" {
		queryAsString.WriteString(".HasLabel(")
		queryAsString.WriteString(label)
		queryAsString.WriteString(")")
	}
	ids := make([]any, 0)
	return &Query[T]{
		db:            db,
		debugString:   &queryAsString,
		ids:           ids,
		conditions:    make([]*QueryCondition, 0),
		label:         label,
		subTraversals: make(map[string]*gremlingo.GraphTraversal),
		orderBy:       nil,
	}
}

// Where adds a condition to the query
func (q *Query[T]) Where(field string, operator comparator.Comparator, value any) *Query[T] {
	queryCondition := QueryCondition{
		field:    field,
		operator: operator,
		value:    value,
	}
	q.writeDebugString(queryCondition.String())

	q.conditions = append(
		q.conditions, &queryCondition,
	)
	return q
}

// WhereTraversal adds a custom Gremlin traversal condition
func (q *Query[T]) WhereTraversal(traversal *gremlingo.GraphTraversal) *Query[T] {
	queryCondition := QueryCondition{
		traversal: traversal,
	}
	q.writeDebugString(queryCondition.String())
	q.conditions = append(
		q.conditions, &queryCondition,
	)
	return q
}

// Dedup removes duplicate results from the query
func (q *Query[T]) Dedup() *Query[T] {
	q.writeDebugString(".Dedup()")
	q.dedup = true
	return q
}

// IDs adds the ids to the query
// You can use this to speed up the query by using the graph index
func (q *Query[T]) IDs(id ...any) *Query[T] {
	if os.Getenv("GSM_DEBUG") == "true" {
		q.writeDebugString(".V(")
		for _, id := range id {
			q.writeDebugString(fmt.Sprintf("%v, ", id))
		}
		q.writeDebugString(")")
	}
	q.ids = append(q.ids, id...)
	return q
}

// Limit sets the maximum number of results
func (q *Query[T]) Limit(limit int) *Query[T] {
	q.writeDebugString(".Limit(")
	q.writeDebugString(strconv.Itoa(limit))
	q.writeDebugString(")")
	q.limit = &limit
	return q
}

// Offset sets the number of results to skip
func (q *Query[T]) Offset(offset int) *Query[T] {
	q.writeDebugString(".Skip(")
	q.writeDebugString(strconv.Itoa(offset))
	q.writeDebugString(")")
	q.offset = &offset
	return q
}

// OrderBy adds ordering to the query
func (q *Query[T]) OrderBy(field string, order GremlinOrder) *Query[T] {
	if q.orderBy != nil {
		q.db.logger.Warn(
			"Order by was already defined secondary order by will override original order",
		)
	}
	q.writeDebugString(".OrderBy(")
	q.writeDebugString(field)
	q.writeDebugString(", ")
	if order == Desc {
		q.writeDebugString("Order.Desc")
	} else {
		q.writeDebugString("Order.Asc")
	}
	q.writeDebugString(")")
	desc := order != 0
	q.orderBy = &OrderCondition{field: field, desc: desc}
	return q
}

// Find executes the query and returns all matching results
func (q *Query[T]) Find() ([]T, error) {
	var val T
	q.writeDebugString(".ToList()")
	query := q.BuildQuery()
	queryMap := structToQueryMap(val)
	maps.Copy(queryMap, q.subTraversals)
	query = projectTraversal(query, queryMap)
	// query = ValueMapTraversal(query, true)
	queryResults, err := query.ToList()
	if err != nil {
		return nil, err
	}

	results := make([]T, 0, len(queryResults))
	for _, result := range queryResults {
		log.Infof("Result: %v", result)
		var v T
		err = UnloadGremlinResultIntoStruct(&v, result)
		if err != nil {
			return nil, err
		}
		results = append(results, v)
	}
	return results, nil
}

// Take executes the query and returns the first result
func (q *Query[T]) Take() (T, error) {
	q.writeDebugString(".Next()")
	var v T
	query := q.BuildQuery()
	queryMap := structToQueryMap(v)
	maps.Copy(queryMap, q.subTraversals)
	query = projectTraversal(query, queryMap)
	result, err := query.Next()
	if err != nil {
		return v, err
	}

	err = UnloadGremlinResultIntoStruct(&v, result)
	return v, err
}

// Count returns the number of matching results
func (q *Query[T]) Count() (int, error) {
	q.writeDebugString(".Count()")
	query := q.BuildQuery()
	result, err := query.Count().Next()
	if err != nil {
		return 0, err
	}
	num, err := result.GetInt()
	if err != nil {
		return 0, err
	}
	return num, nil
}

// Delete deletes all matching results
func (q *Query[T]) Delete() error {
	q.writeDebugString(".Drop().Iterate()")
	query := q.BuildQuery()
	err := query.Drop().Iterate()
	return <-err
}

// ID finds vertex by id in a more optimized way than using where
func (q *Query[T]) ID(id any) (T, error) {
	var v T
	query := q.db.g.V(id)
	label, err := getLabel[T]()
	if err != nil {
		return v, err
	}
	query = query.HasLabel(label)
	queryMap := structToQueryMap(v)
	maps.Copy(queryMap, q.subTraversals)
	query = projectTraversal(query, queryMap)
	result, err := query.Next()
	if err != nil {
		return v, err
	}
	err = UnloadGremlinResultIntoStruct(&v, result)
	return v, err
}

// Update updates a property of the struct
// NOTE: Slices will be updated as Cardinality.Set
// NOTE: Maps will be updated as Cardinality.Set with keys as the value of the property
func (q *Query[T]) Update(propertyName string, value any) error {
	// figure out if propertyName is in the struct
	_, fieldType, err := getStructFieldNameAndType[T](propertyName)
	if err != nil {
		return fmt.Errorf("propertyName not found in gremlin struct tags: %s", propertyName)
	}
	query := q.BuildQuery()
	query.Property(cardinality.Single, gsmtypes.LastModified, time.Now().UTC())
	switch fieldType.Kind() { //nolint: exhaustive // We are only handling slices and maps otherwise regular cardinality
	case reflect.Slice:
		cardinality := gremlingo.Cardinality.List
		cardinalityString := "Cardinality.List"
		if q.db.dbDriver == Neptune {
			cardinalityString = "Cardinality.Set"
			cardinality = gremlingo.Cardinality.Set
		}
		sliceValue, _ := value.([]any)
		for _, v := range sliceValue {
			q.writeDebugString(".Property(")
			q.writeDebugString(cardinalityString)
			q.writeDebugString(", ")
			q.writeDebugString(propertyName)
			q.writeDebugString(", ")
			q.writeDebugString(fmt.Sprintf("%v", v))
			q.writeDebugString(")")
			query = query.Property(cardinality, propertyName, v)
		}
	case reflect.Map:
		mapValue, _ := value.(map[any]any)
		for k := range mapValue {
			q.writeDebugString(".Property(Cardinality.Set, ")
			q.writeDebugString(propertyName)
			q.writeDebugString(", ")
			q.writeDebugString(fmt.Sprintf("%v", k))
			q.writeDebugString(")")
			query = query.Property(gremlingo.Cardinality.Set, propertyName, k)
		}
	default:
		q.writeDebugString(".Property(Cardinality.Single, ")
		q.writeDebugString(propertyName)
		q.writeDebugString(", ")
		q.writeDebugString(fmt.Sprintf("%v", value))
		q.writeDebugString(")")
		query = query.Property(gremlingo.Cardinality.Single, propertyName, value)
	}
	errChan := query.Iterate()
	return <-errChan
}

func (q *Query[T]) SubTraversal(structField string, value *gremlingo.GraphTraversal) *Query[T] {
	var v T
	rt := reflect.TypeOf(v)
	field, ok := rt.FieldByName(structField)
	if !ok {
		q.db.logger.Errorf("struct field: %s not found in struct", structField)
		return q
	}
	gremlinTag := field.Tag.Get("gremlin")
	if gremlinTag == "" || gremlinTag == "-" {
		q.db.logger.Errorf(
			"struct field: %s does not have a gremlin tag or is not exported",
			structField,
		)
		return q
	}
	q.subTraversals[gremlinTag] = value
	return q
}

// writeDebugString writes a string to the debug string if GSM_DEBUG is set to true
func (q *Query[T]) writeDebugString(s string) {
	if os.Getenv("GSM_DEBUG") == "true" {
		q.debugString.WriteString(s)
	}
}

// BuildQuery constructs the Gremlin traversal from the query conditions
func (q *Query[T]) BuildQuery() *gremlingo.GraphTraversal {
	if os.Getenv("GSM_DEBUG") == "true" {
		q.db.logger.Infof("Running Query: %s", q.debugString.String())
		q.debugString.Reset()
	}
	var query *gremlingo.GraphTraversal
	if len(q.ids) > 0 {
		query = q.db.g.V(q.ids...)
	} else {
		query = q.db.g.V()
	}

	if q.label != "" {
		query = query.HasLabel(q.label)
	}

	q.addQueryConditions(query)

	if q.dedup {
		query = query.Dedup()
	}

	if q.orderBy != nil {
		if q.orderBy.desc {
			query.Order().By(q.orderBy.field, Order.Desc)
		} else {
			query.Order().By(q.orderBy.field, Order.Asc)
		}
	}

	// Apply offset
	if q.offset != nil {
		query = query.Skip(*q.offset)
	}

	// Apply limit
	if q.limit != nil {
		query = query.Limit(*q.limit)
	}

	return query
}

func (q *Query[T]) addQueryConditions(query *gremlingo.GraphTraversal) {
	// Apply conditions
	for _, condition := range q.conditions {
		if condition.traversal != nil {
			query = query.Where(condition.traversal)
			continue
		}
		switch condition.operator {
		case comparator.EQ, "eq":
			if condition.field == "id" {
				query = query.HasId(condition.value)
			} else {
				query = query.Has(condition.field, condition.value)
			}
		case comparator.NEQ, "neq":
			query = query.Has(condition.field, gremlingo.P.Neq(condition.value))
		case comparator.GT, "gt":
			query = query.Has(condition.field, gremlingo.P.Gt(condition.value))
		case comparator.GTE, "gte":
			query = query.Has(condition.field, gremlingo.P.Gte(condition.value))
		case comparator.LT, "lt":
			query = query.Has(condition.field, gremlingo.P.Lt(condition.value))
		case comparator.LTE, "lte":
			query = query.Has(condition.field, gremlingo.P.Lte(condition.value))
		case comparator.IN:
			if slice, ok := condition.value.([]any); ok {
				query = query.Has(condition.field, gremlingo.P.Within(slice...))
			}
		case comparator.CONTAINS:
			if strVal, ok := condition.value.(string); ok {
				query = query.Has(condition.field, gremlingo.TextP.Containing(strVal))
			}
		case comparator.WITHOUT:
			if slice, ok := condition.value.([]any); ok {
				query = query.Has(condition.field, gremlingo.P.Without(slice...))
			}
		}
	}
}

// ValueMapTraversal is a helper function to create a value map traversal
// if a property has a single value, it will be returned as a single value
// if a property has multiple values, it will be returned as a slice of values
func ValueMapTraversal(query *gremlingo.GraphTraversal, args ...any) *gremlingo.GraphTraversal {
	return query.ValueMap(args...).By(
		anonymousTraversal.Choose(
			anonymousTraversal.Count(Scope.Local).Is(P.Eq(1)),
			anonymousTraversal.Unfold(),
			anonymousTraversal.Identity(),
		),
	)
}

func structToQueryMap(v gsmtypes.VertexType) map[string]*gremlingo.GraphTraversal {
	queryMap := make(map[string]*gremlingo.GraphTraversal)
	rt := reflect.TypeOf(v)

	// Handle pointer types
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	for i := range rt.NumField() {
		field := rt.Field(i)

		// Handle embedded (anonymous) structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Create a zero value of the embedded struct type to process it
			embeddedValue := reflect.New(field.Type).Interface()
			// Check if it implements VertexType, if so use it, otherwise process as struct type
			if embeddedVertexType, ok := embeddedValue.(gsmtypes.VertexType); ok {
				// Recursively process the embedded struct
				embeddedMap := structToQueryMap(embeddedVertexType)
				// Merge the embedded map into the parent map
				maps.Copy(queryMap, embeddedMap)
			} else {
				// If it doesn't implement VertexType, process it directly using reflection
				embeddedMap := processStructType(field.Type)
				maps.Copy(queryMap, embeddedMap)
			}
			continue
		}

		gremlinTag := field.Tag.Get("gremlin")
		if gremlinTag == "" || gremlinTag == "-" {
			continue
		}

		var traversals []any
		if field.Type.Kind() == reflect.Array || field.Type.Kind() == reflect.Slice {
			traversals = append(traversals, gremlingo.T__.Values(gremlinTag).Fold())
		} else {
			traversals = append(traversals, gremlingo.T__.Values(gremlinTag))
		}
		traversals = append(traversals, gremlingo.T__.Constant(nil))
		queryMap[gremlinTag] = gremlingo.T__.Coalesce(traversals...)
	}
	return queryMap
}

// processStructType processes a struct type directly without requiring VertexType interface
func processStructType(rt reflect.Type) map[string]*gremlingo.GraphTraversal {
	queryMap := make(map[string]*gremlingo.GraphTraversal)

	// Handle pointer types
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return queryMap
	}

	for i := range rt.NumField() {
		field := rt.Field(i)

		// Handle embedded (anonymous) structs recursively
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embeddedMap := processStructType(field.Type)
			for k, v := range embeddedMap {
				queryMap[k] = v
			}
			continue
		}

		gremlinTag := field.Tag.Get("gremlin")
		if gremlinTag == "" || gremlinTag == "-" {
			continue
		}

		var traversals []any
		if field.Type.Kind() == reflect.Array || field.Type.Kind() == reflect.Slice {
			traversals = append(traversals, gremlingo.T__.Values(gremlinTag).Fold())
		} else {
			traversals = append(traversals, gremlingo.T__.Values(gremlinTag))
		}
		traversals = append(traversals, gremlingo.T__.Constant(nil))
		queryMap[gremlinTag] = gremlingo.T__.Coalesce(traversals...)
	}
	return queryMap
}

// projectTraversal projects the query map into the query
func projectTraversal(
	query *gremlingo.GraphTraversal,
	queryMap map[string]*gremlingo.GraphTraversal,
) *gremlingo.GraphTraversal {
	keys := make([]any, 0, len(queryMap))
	for k := range queryMap {
		keys = append(keys, k)
	}
	query = query.Project(keys...)
	for _, k := range keys {
		if k.(string) == "id" { //nolint:errcheck // we already know the key is a string
			query = query.By(
				gremlingo.T__.Id(),
			)
			continue
		}
		query = query.By(
			queryMap[k.(string)], //nolint:errcheck // we already know the key is a string
		)
	}
	return query
}
