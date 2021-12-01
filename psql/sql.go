package psql

import (
	"fmt"
	"io"
)

type PlaceHolderType int

const (
	Question PlaceHolderType = iota
)

func (pt PlaceHolderType) Mark() string {
	switch pt {
	case Question:
		return "?"
	}
	return ""
}

type SqlCond interface {
	ToWhere(pt PlaceHolderType) (query string, args []interface{}, err error)
}

type SqlBuilder struct {
	HolderType PlaceHolderType
}

type SqlStatement interface {
	ToSql() (query string, args []interface{}, err error)
}

func NewSqlBuilder(holderType PlaceHolderType) SqlBuilder {
	return SqlBuilder{HolderType: holderType}
}

func (s SqlBuilder) Select(columns ...string) *SelectStatement {
	return NewSelect(s.HolderType).Column(columns...)
}

func (s SqlBuilder) Insert(table string) *InsertStatement {
	return NewInsert(s.HolderType).Table(table)
}

func (s SqlBuilder) Delete(table string) *DeleteStatement {
	return NewDelete(s.HolderType).Table(table)
}

func (s SqlBuilder) Update(table string) *UpdateStatement {
	return NewUpdate(s.HolderType).Table(table)
}

func Select(columns ...string) *SelectStatement {
	return NewSelect(Question).Column(columns...)
}

func Insert(table string) *InsertStatement {
	return NewInsert(Question).Table(table)
}

func Delete(table string) *DeleteStatement {
	return NewDelete(Question).Table(table)
}

func Update(table string) *UpdateStatement {
	return NewUpdate(Question).Table(table)
}

type SqlParam struct {
	query interface{}
	args  []interface{}
}

func (sp SqlParam) ToWhere(pt PlaceHolderType) (query string, args []interface{}, err error) {
	st, ok := sp.query.(SqlCond)
	if ok {
		return st.ToWhere(pt)
	}

	switch qt := sp.query.(type) {
	case string:
		return qt, args, nil
	case map[string]interface{}:
		return Eq(qt).ToWhere(pt)
	default:
		return query, args, fmt.Errorf("query has wrong type. query = %#v", query)
	}
}

func appendToSql(transforms []SqlCond, connect string, writer io.Writer, args []interface{}, holderType PlaceHolderType) ([]interface{}, error) {
	for index, tran := range transforms {
		if index > 0 {
			_, err := io.WriteString(writer, connect)
			if err != nil {
				return nil, err
			}
		}
		tq, targs, err := tran.ToWhere(holderType)
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(writer, tq)
		if err != nil {
			return nil, err
		}
		args = append(args, targs...)
	}

	return args, nil
}
