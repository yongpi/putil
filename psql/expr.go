package psql

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type expr map[string]interface{}
type symbol int

const (
	eq symbol = iota
	notEq
	like
	notLike
	lt
	lte
	gt
	gte
)

func (s symbol) string(isList, isNull bool) string {
	switch s {
	case eq:
		if isList {
			return "IN"
		}
		if isNull {
			return "IS NULL"
		}
		return "="
	case notEq:
		if isList {
			return "NOT IN"
		}
		if isNull {
			return "IS NOT NULL"
		}
		return "<>"
	case like:
		return "LIKE"
	case notLike:
		return "NOT LIKE"
	case lt:
		return "<"
	case lte:
		return "<="
	case gt:
		return ">"
	case gte:
		return ">="
	}
	return ""
}

func exprToSql(data expr, sl symbol, pt PlaceHolderType) (query string, args []interface{}, err error) {
	var sql strings.Builder
	var index int
	var keys []string

	for key := range data {
		keys = append(keys, key)
	}

	// 排序，避免每次都不一样
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]
		if index > 0 {
			_, err = sql.WriteString(" And ")
			if err != nil {
				return
			}
		}
		isNull := value == nil
		isList := isListType(value)

		sls := sl.string(isList, isNull)
		if isList && sl != eq && sl != notEq {
			return "", nil, fmt.Errorf("expression %s value can not be list, value = %#v", sls, value)
		}
		var exprSql string
		if isNull {
			exprSql = fmt.Sprintf("%s %s", key, sls)
		} else if isList {
			vv := reflect.ValueOf(value)
			var phs []string
			for i := 0; i < vv.Len(); i++ {
				args = append(args, vv.Index(i).Interface())
				phs = append(phs, pt.Mark())
			}
			exprSql = fmt.Sprintf("%s %s (%s)", key, sls, strings.Join(phs, ","))
		} else {
			exprSql = fmt.Sprintf("%s %s %s", key, sls, pt.Mark())
			args = append(args, value)
		}

		_, err = sql.WriteString(exprSql)
		if err != nil {
			return
		}
		index++
	}

	return sql.String(), args, nil
}

func isListType(value interface{}) bool {
	vt := reflect.TypeOf(value)
	return vt.Kind() == reflect.Array || vt.Kind() == reflect.Slice
}

type Eq expr

func (e Eq) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), eq, pt)
}

type NotEq expr

func (e NotEq) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), notEq, pt)
}

type Like expr

func (e Like) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), like, pt)
}

type NotLike expr

func (e NotLike) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), notLike, pt)
}

type Lt expr

func (e Lt) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), lt, pt)
}

type Lte expr

func (e Lte) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), lte, pt)
}

type Gt expr

func (e Gt) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), gt, pt)
}

type Gte expr

func (e Gte) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return exprToSql(expr(e), gte, pt)
}

type cond []SqlTransform
type condType int

const (
	and condType = iota
	or
)

func (cd condType) string() string {
	switch cd {
	case and:
		return "AND"
	case or:
		return "OR"
	}
	return ""
}

func condToSql(conditions cond, ct condType, pt PlaceHolderType) (query string, args []interface{}, err error) {
	var sql strings.Builder
	cts := ct.string()

	for index, condition := range conditions {
		if index > 0 {
			_, err = sql.WriteString(fmt.Sprintf(" %s ", cts))
			if err != nil {
				return "", nil, err
			}
		}
		cq, cs, err := condition.ToSql(pt)
		if err != nil {
			return "", nil, err
		}
		_, err = sql.WriteString(cq)
		if err != nil {
			return "", nil, err
		}
		args = append(args, cs...)
	}
	if len(conditions) > 1 {
		return fmt.Sprintf("(%s)", sql.String()), args, nil
	}
	return sql.String(), args, nil
}

type And cond

func (a And) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return condToSql(cond(a), and, pt)
}

type Or cond

func (o Or) ToSql(pt PlaceHolderType) (query string, args []interface{}, err error) {
	return condToSql(cond(o), or, pt)
}
