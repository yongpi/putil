package psql

import (
	"fmt"
	"strings"
)

type deleteTransform struct {
	holderType PlaceHolderType
	table      string
	wheres     []SqlTransform
}

func NewDelete(holderType PlaceHolderType) *deleteTransform {
	return &deleteTransform{holderType: holderType}
}

func (t *deleteTransform) Table(table string) *deleteTransform {
	t.table = table
	return t
}

func (t *deleteTransform) Where(query interface{}, args ...interface{}) *deleteTransform {
	t.wheres = append(t.wheres, SqlParam{query: query, args: args})
	return t
}

func (t *deleteTransform) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	_, err = sql.WriteString(fmt.Sprintf("DELETE FROM %s ", t.table))
	if err != nil {
		return
	}
	if len(t.wheres) > 0 {
		_, err = sql.WriteString("WHERE ")
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
