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
type updateTransform struct {
	holderType PlaceHolderType
	table      string
	sets       []setParam
	wheres     []SqlTransform
}

func NewUpdate(holderType PlaceHolderType) *updateTransform {
	return &updateTransform{holderType: Question}
}

func (t *updateTransform) Table(table string) *updateTransform {
	t.table = table
	return t
}

func (t *updateTransform) Set(column string, value interface{}) *updateTransform {
	t.sets = append(t.sets, setParam{column: column, value: value})
	return t
}

func (t *updateTransform) SetMap(data map[string]interface{}) *updateTransform {
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

func (t *updateTransform) Where(query interface{}, args ...interface{}) *updateTransform {
	t.wheres = append(t.wheres, SqlParam{query: query, args: args})
	return t
}

func (t *updateTransform) ToSql() (query string, args []interface{}, err error) {
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
