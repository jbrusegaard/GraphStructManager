package driver

import (
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// Query represents a chainable query builder
type Query[T VertexType] struct {
	db         *GremlinDriver
	conditions []QueryCondition
	label      string
	limit      *int
	offset     *int
	orderBy    []OrderCondition
}

type QueryCondition struct {
	field    string
	operator string
	value    any
}

type OrderCondition struct {
	field string
	desc  bool
}

// NewQuery creates a new query builder for type T
func NewQuery[T VertexType](db *GremlinDriver) *Query[T] {
	label, _ := getStructName[T]()
	return &Query[T]{
		db:         db,
		conditions: make([]QueryCondition, 0),
		label:      label,
		orderBy:    make([]OrderCondition, 0),
	}
}

// Where adds a condition to the query
func (q *Query[T]) Where(field string, operator string, value any) *Query[T] {
	q.conditions = append(q.conditions, QueryCondition{
		field:    field,
		operator: operator,
		value:    value,
	})
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
func (q *Query[T]) OrderBy(field string) *Query[T] {
	q.orderBy = append(q.orderBy, OrderCondition{field: field, desc: false})
	return q
}

// OrderByDesc adds descending ordering to the query
func (q *Query[T]) OrderByDesc(field string) *Query[T] {
	q.orderBy = append(q.orderBy, OrderCondition{field: field, desc: true})
	return q
}

// Find executes the query and returns all matching results
func (q *Query[T]) Find() ([]T, error) {
	query := q.buildQuery()
	queryResults, err := query.ElementMap().ToList()
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

// First executes the query and returns the first result
func (q *Query[T]) First() (T, error) {
	var v T
	query := q.buildQuery()
	result, err := query.ElementMap().Next()
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
	_, err := query.Drop().Next()
	return err
}

// buildQuery constructs the Gremlin traversal from the query conditions
func (q *Query[T]) buildQuery() *gremlingo.GraphTraversal {
	query := q.db.g.V().HasLabel(q.label)

	// Apply conditions
	for _, condition := range q.conditions {
		switch condition.operator {
		case "=", "eq":
			query = query.Has(condition.field, condition.value)
		case "!=", "neq":
			query = query.Has(condition.field, gremlingo.P.Neq(condition.value))
		case ">", "gt":
			query = query.Has(condition.field, gremlingo.P.Gt(condition.value))
		case ">=", "gte":
			query = query.Has(condition.field, gremlingo.P.Gte(condition.value))
		case "<", "lt":
			query = query.Has(condition.field, gremlingo.P.Lt(condition.value))
		case "<=", "lte":
			query = query.Has(condition.field, gremlingo.P.Lte(condition.value))
		case "in":
			if slice, ok := condition.value.([]any); ok {
				query = query.Has(condition.field, gremlingo.P.Within(slice...))
			}
		case "contains":
			query = query.Has(condition.field, gremlingo.TextP.Containing(condition.value))
		}
	}

	// Apply ordering
	for _, order := range q.orderBy {
		if order.desc {
			query = query.Order().By(order.field, gremlingo.Order.Desc)
		} else {
			query = query.Order().By(order.field, gremlingo.Order.Asc)
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
