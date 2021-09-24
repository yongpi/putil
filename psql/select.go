package psql

import (
	"fmt"
	"strings"
)

type SelectTransform struct {
	holderType  PlaceHolderType
	TableName   string
	Columns     []string
	Wheres      []SqlTransform
	OrderBys    []SqlTransform
	LimitValue  *int64
	OffsetValue *int64
	Joins       []SqlTransform
	GroupBys    []SqlTransform
}

func NewSelect(holderType PlaceHolderType) *SelectTransform {
	return &SelectTransform{holderType: holderType}
}

func (st *SelectTransform) Column(columns ...string) *SelectTransform {
	st.Columns = append(st.Columns, columns...)
	return st
}

func (st *SelectTransform) From(table string) *SelectTransform {
	st.TableName = table
	return st
}

func (st *SelectTransform) OrderBy(orderBys ...string) *SelectTransform {
	for _, orderBy := range orderBys {
		st.OrderBys = append(st.OrderBys, SqlParam{query: orderBy})
	}
	return st
}

func (st *SelectTransform) GroupBy(groupBys ...string) *SelectTransform {
	for _, groupBy := range groupBys {
		st.GroupBys = append(st.GroupBys, SqlParam{query: groupBy})
	}
	return st
}

func (st *SelectTransform) Limit(limit int64) *SelectTransform {
	st.LimitValue = &limit
	return st
}

func (st *SelectTransform) Offset(offset int64) *SelectTransform {
	st.OffsetValue = &offset
	return st
}

func (st *SelectTransform) Where(query interface{}, args ...interface{}) *SelectTransform {
	st.Wheres = append(st.Wheres, SqlParam{query: query, args: args})
	return st
}

func (st *SelectTransform) Join(query string, args ...interface{}) *SelectTransform {
	st.Joins = append(st.Joins, SqlParam{query: fmt.Sprintf("JOIN %s", query), args: args})
	return st
}

func (st *SelectTransform) LeftJoin(query string, args ...interface{}) *SelectTransform {
	st.Joins = append(st.Joins, SqlParam{query: fmt.Sprintf("LEFT JOIN %s", query), args: args})
	return st
}

func (st *SelectTransform) RightJoin(query string, args ...interface{}) *SelectTransform {
	st.Joins = append(st.Joins, SqlParam{query: fmt.Sprintf("RIGHT JOIN %s", query), args: args})
	return st
}

func (st *SelectTransform) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	holdType := st.holderType
	sql.WriteString("SELECT ")

	if len(st.Columns) == 0 {
		return "", nil, fmt.Errorf("select sql lack of column")
	}
	sql.WriteString(strings.Join(st.Columns, ","))

	if st.TableName == "" {
		return "", nil, fmt.Errorf("select sql lack of TableName")
	}
	sql.WriteString(fmt.Sprintf(" FROM %s ", st.TableName))

	if len(st.Joins) > 0 {
		args, err = appendToSql(st.Joins, " ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if len(st.Wheres) > 0 {
		sql.WriteString(" Where ")
		args, err = appendToSql(st.Wheres, " AND ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if len(st.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		args, err = appendToSql(st.GroupBys, ", ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if len(st.OrderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		args, err = appendToSql(st.OrderBys, ", ", &sql, args, holdType)
		if err != nil {
			return
		}
	}

	if st.LimitValue != nil {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", *st.LimitValue))
	}
	if st.OffsetValue != nil {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", *st.OffsetValue))
	}

	return sql.String(), args, nil
}
