package driver

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/gobeam/stringy"
	"github.com/jbrusegaard/graph-struct-manager/comparator"
	"github.com/jbrusegaard/graph-struct-manager/gsmtypes"
)

var cardinality = gremlingo.Cardinality

// Query represents a chainable query builder
type Query[T VertexType] struct {
	db            *GremlinDriver
	conditions    []*QueryCondition
	label         string
	limit         *int
	offset        *int
	orderBy       *OrderCondition
	dedup         bool
	queryAsString *strings.Builder
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

func getLabel[T VertexType]() (string, error) {
	structName, err := getStructName[T]()
	if err != nil {
		return "", err
	}
	return stringy.New(structName).SnakeCase().ToLower(), nil
}

// NewQuery creates a new query builder for type T
func NewQuery[T VertexType](db *GremlinDriver) *Query[T] {
	label, _ := getLabel[T]()
	queryAsString := strings.Builder{}
	queryAsString.WriteString("V()")
	if label != "" {
		queryAsString.WriteString(".HasLabel(")
		queryAsString.WriteString(label)
		queryAsString.WriteString(")")
	}
	return &Query[T]{
		db:            db,
		queryAsString: &queryAsString,
		conditions:    make([]*QueryCondition, 0),
		label:         label,
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
	q.queryAsString.WriteString(queryCondition.String())

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
	q.queryAsString.WriteString(queryCondition.String())
	q.conditions = append(
		q.conditions, &queryCondition,
	)
	return q
}

// Dedup removes duplicate results from the query
func (q *Query[T]) Dedup() *Query[T] {
	q.queryAsString.WriteString(".Dedup()")
	q.dedup = true
	return q
}

// Limit sets the maximum number of results
func (q *Query[T]) Limit(limit int) *Query[T] {
	q.queryAsString.WriteString(".Limit(")
	q.queryAsString.WriteString(strconv.Itoa(limit))
	q.queryAsString.WriteString(")")
	q.limit = &limit
	return q
}

// Offset sets the number of results to skip
func (q *Query[T]) Offset(offset int) *Query[T] {
	q.queryAsString.WriteString(".Skip(")
	q.queryAsString.WriteString(strconv.Itoa(offset))
	q.queryAsString.WriteString(")")
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
	q.queryAsString.WriteString(".OrderBy(")
	q.queryAsString.WriteString(field)
	q.queryAsString.WriteString(", ")
	if order == Desc {
		q.queryAsString.WriteString("Order.Desc")
	} else {
		q.queryAsString.WriteString("Order.Asc")
	}
	q.queryAsString.WriteString(")")
	desc := order != 0
	q.orderBy = &OrderCondition{field: field, desc: desc}
	return q
}

// Find executes the query and returns all matching results
func (q *Query[T]) Find() ([]T, error) {
	q.queryAsString.WriteString(".ToList()")
	query := q.buildQuery()
	queryResults, err := toMapTraversal(query, true).ToList()
	if err != nil {
		return nil, err
	}

	results := make([]T, 0, len(queryResults))
	for _, result := range queryResults {
		var v T
		err = unloadGremlinResultIntoStruct(&v, result)
		if err != nil {
			return nil, err
		}
		results = append(results, v)
	}
	return results, nil
}

// Take executes the query and returns the first result
func (q *Query[T]) Take() (T, error) {
	q.queryAsString.WriteString(".Next()")
	var v T
	query := q.buildQuery()
	result, err := toMapTraversal(query, true).Next()
	if err != nil {
		return v, err
	}

	err = unloadGremlinResultIntoStruct(&v, result)
	return v, err
}

// Count returns the number of matching results
func (q *Query[T]) Count() (int, error) {
	q.queryAsString.WriteString(".Count()")
	query := q.buildQuery()
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
	q.queryAsString.WriteString(".Drop().Iterate()")
	query := q.buildQuery()
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
	result, err := toMapTraversal(query, true).Next()
	if err != nil {
		return v, err
	}
	err = unloadGremlinResultIntoStruct(&v, result)
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
	query := q.buildQuery()
	query.Property(cardinality.Single, gsmtypes.LastModified, time.Now().UTC())
	switch fieldType.Kind() { //nolint: exhaustive // We are only handling slices and maps otherwise regular cardinality
	case reflect.Slice:
		sliceValue, _ := value.([]any)
		for _, v := range sliceValue {
			q.queryAsString.WriteString(".Property(Cardinality.Set, ")
			q.queryAsString.WriteString(propertyName)
			q.queryAsString.WriteString(", ")
			fmt.Fprintf(q.queryAsString, "%v", v)
			q.queryAsString.WriteString(")")
			query = query.Property(gremlingo.Cardinality.Set, propertyName, v)
		}
	case reflect.Map:
		mapValue, _ := value.(map[any]any)
		for k := range mapValue {
			q.queryAsString.WriteString(".Property(Cardinality.Set, ")
			q.queryAsString.WriteString(propertyName)
			q.queryAsString.WriteString(", ")
			fmt.Fprintf(q.queryAsString, "%v", k)
			q.queryAsString.WriteString(")")
			query = query.Property(gremlingo.Cardinality.Set, propertyName, k)
		}
	default:
		q.queryAsString.WriteString(".Property(Cardinality.Single, ")
		q.queryAsString.WriteString(propertyName)
		q.queryAsString.WriteString(", ")
		fmt.Fprintf(q.queryAsString, "%v", value)
		q.queryAsString.WriteString(")")
		query = query.Property(gremlingo.Cardinality.Single, propertyName, value)
	}
	errChan := query.Iterate()
	return <-errChan
}

// buildQuery constructs the Gremlin traversal from the query conditions
func (q *Query[T]) buildQuery() *gremlingo.GraphTraversal {
	if os.Getenv("GSM_DEBUG") == "true" {
		q.db.logger.Infof("Running Query: %s", q.queryAsString.String())
	}
	query := q.db.g.V()

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

func toMapTraversal(query *gremlingo.GraphTraversal, args ...any) *gremlingo.GraphTraversal {
	return query.ValueMap(args...).By(
		anonymousTraversal.Choose(
			anonymousTraversal.Count(Scope.Local).Is(P.Eq(1)),
			anonymousTraversal.Unfold(),
			anonymousTraversal.Identity(),
		),
	)
}
