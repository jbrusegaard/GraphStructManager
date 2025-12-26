package driver

import (
	"fmt"
	"reflect"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/gobeam/stringy"
	"github.com/jbrusegaard/graph-struct-manager/comparator"
	"github.com/jbrusegaard/graph-struct-manager/gsmtypes"
)

var cardinality = gremlingo.Cardinality

// Query represents a chainable query builder
type Query[T VertexType] struct {
	db         *GremlinDriver
	conditions []QueryCondition
	label      string
	limit      *int
	offset     *int
	orderBy    *OrderCondition
}

type QueryCondition struct {
	field     string
	operator  comparator.Comparator
	value     any
	traversal *gremlingo.GraphTraversal
}

type OrderCondition struct {
	field string
	desc  bool
}

// NewQuery creates a new query builder for type T
func NewQuery[T VertexType](db *GremlinDriver) *Query[T] {
	structName, _ := getStructName[T]()
	label := stringy.New(structName)
	return &Query[T]{
		db:         db,
		conditions: make([]QueryCondition, 0),
		label:      label.SnakeCase().ToLower(),
		orderBy:    nil,
	}
}

// Where adds a condition to the query
func (q *Query[T]) Where(field string, operator comparator.Comparator, value any) *Query[T] {
	q.conditions = append(
		q.conditions, QueryCondition{
			field:    field,
			operator: operator,
			value:    value,
		},
	)
	return q
}

// WhereTraversal adds a custom Gremlin traversal condition
func (q *Query[T]) WhereTraversal(traversal *gremlingo.GraphTraversal) *Query[T] {
	q.conditions = append(
		q.conditions, QueryCondition{
			traversal: traversal,
		},
	)
	return q
}

// Dedup removes duplicate results from the query
func (q *Query[T]) Dedup() *Query[T] {
	q.conditions = append(
		q.conditions, QueryCondition{
			traversal: gremlingo.T__.Dedup(),
		},
	)
	return q
}

// Limit sets the maximum number of results
func (q *Query[T]) Limit(limit int) *Query[T] {
	q.limit = &limit
	return q
}

// Offset sets the number of results to skip
func (q *Query[T]) Offset(offset int) *Query[T] {
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
	desc := order != 0
	q.orderBy = &OrderCondition{field: field, desc: desc}
	return q
}

// Find executes the query and returns all matching results
func (q *Query[T]) Find() ([]T, error) {
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
	query := q.buildQuery()
	err := query.Drop().Iterate()
	return <-err
}

// ID finds vertex by id in a more optimized way than using where
func (q *Query[T]) ID(id any) (T, error) {
	var v T
	query := q.db.g.V(id)
	result, err := toMapTraversal(query, true).Next()
	if err != nil {
		return v, err
	}
	err = unloadGremlinResultIntoStruct(&v, result)
	return v, err
}

func (q *Query[T]) Update(propertyName string, value any) error {
	// figure out if propertyName is in the struct
	_, fieldType, err := getStructFieldNameAndType[T](propertyName)
	if err != nil {
		return fmt.Errorf("propertyName not found in gremlin struct tags: %s", propertyName)
	}
	query := q.buildQuery()
	query.Property(cardinality.Single, gsmtypes.LastModified, time.Now().UTC())
	if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map {
		query = query.Property(gremlingo.Cardinality.Set, propertyName, value)
	} else {
		query = query.Property(gremlingo.Cardinality.Single, propertyName, value)
	}
	errChan := query.Iterate()
	return <-errChan
}

// buildQuery constructs the Gremlin traversal from the query conditions
func (q *Query[T]) buildQuery() *gremlingo.GraphTraversal {
	query := q.db.g.V().HasLabel(q.label)

	q.addQueryConditions(query)

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
