package psql

import (
	"fmt"
	"sort"
	"strings"
)

type insertTransform struct {
	holderType PlaceHolderType
	table      string
	columns    []string
	values     [][]interface{}
}

func NewInsert(holderType PlaceHolderType) *insertTransform {
	return &insertTransform{holderType: holderType}
}

func (it *insertTransform) Table(table string) *insertTransform {
	it.table = table
	return it
}

func (it *insertTransform) Columns(columns ...string) *insertTransform {
	it.columns = append(it.columns, columns...)
	return it
}

func (it *insertTransform) Values(values ...interface{}) *insertTransform {
	it.values = append(it.values, values)
	return it
}

func (it *insertTransform) SetMap(data map[string]interface{}) *insertTransform {
	var columns []string
	for key := range data {
		columns = append(columns, key)
	}

	sort.Strings(columns)
	it.columns = nil
	it.values = nil

	var values []interface{}
	for _, column := range columns {
		it.columns = append(it.columns, column)
		values = append(values, data[column])
	}
	it.values = append(it.values, values)

	return it
}

func (it *insertTransform) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	_, err = sql.WriteString(fmt.Sprintf("INSERT INTO %s ", it.table))
	if err != nil {
		return
	}

	if len(it.columns) > 0 {
		_, err = sql.WriteString(fmt.Sprintf("(%s)", strings.Join(it.columns, ",")))
		if err != nil {
			return
		}
	}

	sql.WriteString(" VALUES ")
	if len(it.values) > 0 {
		for li, list := range it.values {
			if li > 0 {
				_, err = sql.WriteString(",")
				if err != nil {
					return
				}
			}
			var markList []string
			for _, value := range list {
				markList = append(markList, it.holderType.Mark())
				args = append(args, value)
			}
			if len(markList) > 0 {
				_, err = sql.WriteString(fmt.Sprintf("(%s)", strings.Join(markList, ",")))
				if err != nil {
					return
				}
			}
		}
	}

	return sql.String(), args, nil
}
