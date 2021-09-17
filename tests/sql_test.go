package tests

import (
	"testing"

	"github.com/putil/psql"
)

func TestSelect(t *testing.T) {
	query, args, err := psql.NewSelect(psql.Question).
		Column([]string{"id", "name"}...).
		From("test").
		LeftJoin("sku on sku.id=test.id").
		Where(psql.Eq{"name": "sss", "type": []int64{1, 2, 3}}).
		Where(psql.Or{psql.Eq{"name": "ccc"}, psql.And{psql.Eq{"id": 111}, psql.Eq{"desc": "sssss"}}}).
		Where(psql.Eq{"name": "dddd"}).
		Where(psql.And{psql.Eq{"type": 5}}).
		GroupBy("id").
		Limit(1).
		Offset(10).
		ToSql()
	if err != nil {
		t.Fatal(err)
	}
	exSql := "SELECT id, name FROM test " +
		"LEFT JOIN sku on sku.id=test.id " +
		"Where name = ? And type IN (?,?,?) AND (name = ? OR (id = ? AND desc = ?)) AND name = ? AND type = ? " +
		"GROUP BY id " +
		"LIMIT 1 " +
		"OFFSET 10"
	if query != exSql {
		t.Errorf("query not expected sql, query = %s", query)
	}
	exValue := []interface{}{"sss", int64(1), int64(2), int64(3), "ccc", 111, "sssss", "dddd", 5}
	for index, value := range args {
		if exValue[index] != value {
			t.Errorf("args not expected value, args = %#v, value = %v", args, value)
		}
	}
}
