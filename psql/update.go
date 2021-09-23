package psql

import (
	"fmt"
	"sort"
	"strings"
)

type setParam struct {
	column string
	value  interface{}
}
type UpdateTransform struct {
	holderType PlaceHolderType
	table      string
	sets       []setParam
	wheres     []SqlTransform
}

func NewUpdate(holderType PlaceHolderType) *UpdateTransform {
	return &UpdateTransform{holderType: Question}
}

func (t *UpdateTransform) Table(table string) *UpdateTransform {
	t.table = table
	return t
}

func (t *UpdateTransform) Set(column string, value interface{}) *UpdateTransform {
	t.sets = append(t.sets, setParam{column: column, value: value})
	return t
}

func (t *UpdateTransform) SetMap(data map[string]interface{}) *UpdateTransform {
	var columns []string
	for key := range data {
		columns = append(columns, key)
	}
	sort.Strings(columns)

	for _, column := range columns {
		t.sets = append(t.sets, setParam{column: column, value: data[column]})
	}
	return t
}

func (t *UpdateTransform) Where(query interface{}, args ...interface{}) *UpdateTransform {
	t.wheres = append(t.wheres, SqlParam{query: query, args: args})
	return t
}

func (t *UpdateTransform) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	_, err = sql.WriteString(fmt.Sprintf("UPDATE %s ", t.table))
	if err != nil {
		return
	}

	if len(t.sets) > 0 {
		_, err = sql.WriteString("SET ")
		if err != nil {
			return
		}
		for index, set := range t.sets {
			if index > 0 {
				_, err = sql.WriteString(",")
				if err != nil {
					return
				}
			}
			_, err = sql.WriteString(fmt.Sprintf("%s=%s", set.column, t.holderType.Mark()))
			if err != nil {
				return
			}
			args = append(args, set.value)
		}
	}

	if len(t.wheres) > 0 {
		_, err = sql.WriteString(" WHERE ")
		if err != nil {
			return
		}

		args, err = appendToSql(t.wheres, " AND ", &sql, args, t.holderType)
		if err != nil {
			return
		}
	}

	return sql.String(), args, nil
}
