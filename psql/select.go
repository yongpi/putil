package psql

import (
	"fmt"
	"strings"
)

type selectTransform struct {
	holderType  PlaceHolderType
	table       string
	columns     []SqlTransform
	wheres      []SqlTransform
	orderBys    []SqlTransform
	limitValue  *int64
	offsetValue *int64
	joins       []SqlTransform
	groupBys    []SqlTransform
}

func NewSelect(holderType PlaceHolderType) *selectTransform {
	return &selectTransform{holderType: holderType}
}

func (st *selectTransform) Column(columns ...string) *selectTransform {
	for _, column := range columns {
		st.columns = append(st.columns, SqlParam{query: column})
	}
	return st
}

func (st *selectTransform) From(table string) *selectTransform {
	st.table = table
	return st
}

func (st *selectTransform) OrderBy(orderBys ...string) *selectTransform {
	for _, orderBy := range orderBys {
		st.orderBys = append(st.orderBys, SqlParam{query: orderBy})
	}
	return st
}

func (st *selectTransform) GroupBy(groupBys ...string) *selectTransform {
	for _, groupBy := range groupBys {
		st.groupBys = append(st.groupBys, SqlParam{query: groupBy})
	}
	return st
}

func (st *selectTransform) Limit(limit int64) *selectTransform {
	st.limitValue = &limit
	return st
}

func (st *selectTransform) Offset(offset int64) *selectTransform {
	st.offsetValue = &offset
	return st
}

func (st *selectTransform) Where(query interface{}, args ...interface{}) *selectTransform {
	st.wheres = append(st.wheres, SqlParam{query: query, args: args})
	return st
}

func (st *selectTransform) Join(query string, args ...interface{}) *selectTransform {
	st.joins = append(st.joins, SqlParam{query: fmt.Sprintf("JOIN %s", query), args: args})
	return st
}

func (st *selectTransform) LeftJoin(query string, args ...interface{}) *selectTransform {
	st.joins = append(st.joins, SqlParam{query: fmt.Sprintf("LEFT JOIN %s", query), args: args})
	return st
}

func (st *selectTransform) RightJoin(query string, args ...interface{}) *selectTransform {
	st.joins = append(st.joins, SqlParam{query: fmt.Sprintf("RIGHT JOIN %s", query), args: args})
	return st
}

func (st *selectTransform) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	holdType := st.holderType
	sql.WriteString("SELECT ")

	if len(st.columns) == 0 {
		return "", nil, fmt.Errorf("select sql lack of column")
	}
	args, err = appendToSql(st.columns, ", ", &sql, args, holdType)
	if err != nil {
		return
	}

	if st.table == "" {
		return "", nil, fmt.Errorf("select sql lack of table")
	}
	sql.WriteString(fmt.Sprintf(" FROM %s ", st.table))

	if len(st.joins) > 0 {
		args, err = appendToSql(st.joins, " ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if len(st.wheres) > 0 {
		sql.WriteString(" Where ")
		args, err = appendToSql(st.wheres, " AND ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if len(st.groupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		args, err = appendToSql(st.groupBys, ", ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if len(st.orderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		args, err = appendToSql(st.orderBys, ", ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if st.limitValue != nil {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", *st.limitValue))
	}
	if st.offsetValue != nil {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", *st.offsetValue))
	}

	return sql.String(), args, nil
}
