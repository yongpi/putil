package psql

import (
	"fmt"
	"sort"
	"strings"
)

type InsertStatement struct {
	HolderType PlaceHolderType
	TableName  string
	Columns    []string
	Values     [][]interface{}
}

func NewInsert(holderType PlaceHolderType) *InsertStatement {
	return &InsertStatement{HolderType: holderType}
}

func (it *InsertStatement) Table(table string) *InsertStatement {
	it.TableName = table
	return it
}

func (it *InsertStatement) Column(columns ...string) *InsertStatement {
	it.Columns = append(it.Columns, columns...)
	return it
}

func (it *InsertStatement) Value(values ...interface{}) *InsertStatement {
	it.Values = append(it.Values, values)
	return it
}

func (it *InsertStatement) SetMap(data map[string]interface{}) *InsertStatement {
	var columns []string
	for key := range data {
		columns = append(columns, key)
	}

	sort.Strings(columns)
	it.Columns = nil
	it.Values = nil

	var values []interface{}
	for _, column := range columns {
		it.Columns = append(it.Columns, column)
		values = append(values, data[column])
	}
	it.Values = append(it.Values, values)

	return it
}

func (it *InsertStatement) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	_, err = sql.WriteString(fmt.Sprintf("INSERT INTO %s ", it.TableName))
	if err != nil {
		return
	}

	if len(it.Columns) > 0 {
		_, err = sql.WriteString(fmt.Sprintf("(%s)", strings.Join(it.Columns, ",")))
		if err != nil {
			return
		}
	}

	sql.WriteString(" VALUES ")
	if len(it.Values) > 0 {
		for li, list := range it.Values {
			if li > 0 {
				_, err = sql.WriteString(",")
				if err != nil {
					return
				}
			}
			var markList []string
			for _, value := range list {
				markList = append(markList, it.HolderType.Mark())
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
