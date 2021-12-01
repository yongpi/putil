package psql

import (
	"fmt"
	"strings"
)

type DeleteStatement struct {
	HolderType PlaceHolderType
	TableName  string
	Wheres     []SqlCond
}

func NewDelete(holderType PlaceHolderType) *DeleteStatement {
	return &DeleteStatement{HolderType: holderType}
}

func (t *DeleteStatement) Table(table string) *DeleteStatement {
	t.TableName = table
	return t
}

func (t *DeleteStatement) Where(query interface{}, args ...interface{}) *DeleteStatement {
	t.Wheres = append(t.Wheres, SqlParam{query: query, args: args})
	return t
}

func (t *DeleteStatement) ToSql() (query string, args []interface{}, err error) {
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

		args, err = appendToSql(t.Wheres, " AND ", &sql, args, t.HolderType)
		if err != nil {
			return
		}
	}

	return sql.String(), args, nil
}
