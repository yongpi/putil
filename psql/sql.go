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

type SqlTransform interface {
	ToSql(pt PlaceHolderType) (query string, args []interface{}, err error)
}

type SqlParam struct {
	query interface{}
	args  []interface{}
}

func (sp SqlParam) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	st, ok := sp.query.(SqlTransform)
	if ok {
		return st.ToSql(pt)
	}

	switch qt := sp.query.(type) {
	case string:
		return qt, args, nil
	case map[string]interface{}:
		return Eq(qt).ToSql(pt)
	default:
		return query, args, fmt.Errorf("query has wrong type. query = %#v", query)
	}
}

func appendToSql(transforms []SqlTransform, connect string, writer io.Writer, args []interface{}, holderType PlaceHolderType) ([]interface{}, error) {
	for index, tran := range transforms {
		if index > 0 {
			_, err := io.WriteString(writer, connect)
			if err != nil {
				return nil, err
			}
		}
		tq, targs, err := tran.ToSql(holderType)
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
