package psql

import (
	"fmt"
	"sort"
	"strings"
)

type SetParam struct {
	Column string
	Value  interface{}
}
type UpdateTransform struct {
	holderType PlaceHolderType
	TableName  string
	Sets       []SetParam
	Wheres     []SqlTransform
}

func NewUpdate(holderType PlaceHolderType) *UpdateTransform {
	return &UpdateTransform{holderType: Question}
}

func (t *UpdateTransform) Table(table string) *UpdateTransform {
	t.TableName = table
	return t
}

func (t *UpdateTransform) Set(column string, value interface{}) *UpdateTransform {
	t.Sets = append(t.Sets, SetParam{Column: column, Value: value})
	return t
}

func (t *UpdateTransform) SetMap(data map[string]interface{}) *UpdateTransform {
	var columns []string
	for key := range data {
		columns = append(columns, key)
	}
	sort.Strings(columns)

	for _, column := range columns {
		t.Sets = append(t.Sets, SetParam{Column: column, Value: data[column]})
	}
	return t
}

func (t *UpdateTransform) Where(query interface{}, args ...interface{}) *UpdateTransform {
	t.Wheres = append(t.Wheres, SqlParam{query: query, args: args})
	return t
}

func (t *UpdateTransform) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	_, err = sql.WriteString(fmt.Sprintf("UPDATE %s ", t.TableName))
	if err != nil {
		return
	}

	if len(t.Sets) > 0 {
		_, err = sql.WriteString("SET ")
		if err != nil {
			return
		}
		for index, set := range t.Sets {
			if index > 0 {
				_, err = sql.WriteString(",")
				if err != nil {
					return
				}
			}
			_, err = sql.WriteString(fmt.Sprintf("%s=%s", set.Column, t.holderType.Mark()))
			if err != nil {
				return
			}
			args = append(args, set.Value)
		}
	}

	if len(t.Wheres) > 0 {
		_, err = sql.WriteString(" WHERE ")
		if err != nil {
			return
		}

		args, err = appendToSql(t.Wheres, " AND ", &sql, args, t.holderType)
		if err != nil {
			return
		}
	}

	return sql.String(), args, nil
}
