package psql

import (
	"fmt"
	"strings"
)

type DeleteTransform struct {
	holderType PlaceHolderType
	TableName  string
	Wheres     []SqlTransform
}

func NewDelete(holderType PlaceHolderType) *DeleteTransform {
	return &DeleteTransform{holderType: holderType}
}

func (t *DeleteTransform) Table(table string) *DeleteTransform {
	t.TableName = table
	return t
}

func (t *DeleteTransform) Where(query interface{}, args ...interface{}) *DeleteTransform {
	t.Wheres = append(t.Wheres, SqlParam{query: query, args: args})
	return t
}

func (t *DeleteTransform) ToSql() (query string, args []interface{}, err error) {
	var sql strings.Builder
	_, err = sql.WriteString(fmt.Sprintf("DELETE FROM %s ", t.TableName))
	if err != nil {
		return
	}
	if len(t.Wheres) > 0 {
		_, err = sql.WriteString("WHERE ")
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
