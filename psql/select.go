package psql

import (
	"fmt"
	"strings"
)

type SelectStatement struct {
	HolderType  PlaceHolderType
	TableName   string
	Columns     []string
	Wheres      []SqlCond
	OrderBys    []SqlCond
	LimitValue  *int64
	OffsetValue *int64
	Joins       []SqlCond
	GroupBys    []SqlCond
}

func NewSelect(holderType PlaceHolderType) *SelectStatement {
	return &SelectStatement{HolderType: holderType}
}

func (st *SelectStatement) Column(columns ...string) *SelectStatement {
	st.Columns = append(st.Columns, columns...)
	return st
}

func (st *SelectStatement) From(table string) *SelectStatement {
	st.TableName = table
	return st
}

func (st *SelectStatement) OrderBy(orderBys ...string) *SelectStatement {
	for _, orderBy := range orderBys {
		st.OrderBys = append(st.OrderBys, SqlParam{query: orderBy})
	}
	return st
}

func (st *SelectStatement) GroupBy(groupBys ...string) *SelectStatement {
	for _, groupBy := range groupBys {
		st.GroupBys = append(st.GroupBys, SqlParam{query: groupBy})
	}
	return st
}

func (st *SelectStatement) Limit(limit int64) *SelectStatement {
	st.LimitValue = &limit
	return st
}

func (st *SelectStatement) Offset(offset int64) *SelectStatement {
	st.OffsetValue = &offset
	return st
}

func (st *SelectStatement) Where(query interface{}, args ...interface{}) *SelectStatement {
	st.Wheres = append(st.Wheres, SqlParam{query: query, args: args})
	return st
}

func (st *SelectStatement) Join(query string, args ...interface{}) *SelectStatement {
	st.Joins = append(st.Joins, SqlParam{query: fmt.Sprintf("JOIN %s", query), args: args})
	return st
}

func (st *SelectStatement) LeftJoin(query string, args ...interface{}) *SelectStatement {
	st.Joins = append(st.Joins, SqlParam{query: fmt.Sprintf("LEFT JOIN %s", query), args: args})
	return st
}

func (st *SelectStatement) RightJoin(query string, args ...interface{}) *SelectStatement {
	st.Joins = append(st.Joins, SqlParam{query: fmt.Sprintf("RIGHT JOIN %s", query), args: args})
	return st
}

func (st *SelectStatement) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	holdType := st.HolderType
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
