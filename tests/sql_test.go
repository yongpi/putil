package tests

import (
	"testing"

	"github.com/yongpi/putil/psql"
)

func TestSelect(t *testing.T) {
	query, args, err := psql.Select([]string{"id", "name"}...).
		From("test").
		LeftJoin("sku on sku.id=test.id").
		Where(psql.Eq{"name": "sss", "type": []int{1, 2, 3}}).
		Where(psql.Or{psql.Eq{"name": "ccc"}, psql.And{psql.Eq{"id": 111}, psql.Eq{"desc": "sssss"}}}).
		Where(psql.Eq{"name": "dddd"}).
		Where(psql.And{psql.Eq{"type": 5}}).
		GroupBy("id").
		Limit(1).
		Offset(10).
		ToSql()
	if err != nil {
		t.Error(err)
	}
	exSql := "SELECT id,name FROM test " +
		"LEFT JOIN sku on sku.id=test.id " +
		"Where name = ? And type IN (?,?,?) AND (name = ? OR (id = ? AND desc = ?)) AND name = ? AND type = ? " +
		"GROUP BY id " +
		"LIMIT 1 " +
		"OFFSET 10"
	if query != exSql {
		t.Errorf("query not expected sql, query = %s", query)
	}
	exValue := []interface{}{"sss", 1, 2, 3, "ccc", 111, "sssss", "dddd", 5}
	for index, value := range args {
		if exValue[index] != value {
			t.Errorf("args not expected value, args = %#v, value = %v", args, value)
		}
	}
}

func TestInsert(t *testing.T) {
	query, args, err := psql.Insert("test").
		Column("id", "name").
		Value(1, "name1").
		Value(2, "name2").ToSql()

	if err != nil {
		t.Error(err)
	}

	exQuery := "INSERT INTO test (id,name) VALUES (?,?),(?,?)"
	if query != exQuery {
		t.Errorf("query not expected sql, query = %s", query)
	}
	exValue := []interface{}{1, "name1", 2, "name2"}
	for index, value := range args {
		if exValue[index] != value {
			t.Errorf("args not expected value, args = %#v, value = %v", args, value)
		}
	}

	query, args, err = psql.Insert("test").
		SetMap(map[string]interface{}{"id": 1, "name": "name1"}).ToSql()

	if err != nil {
		t.Error(err)
	}

	exQuery = "INSERT INTO test (id,name) VALUES (?,?)"
	if query != exQuery {
		t.Errorf("query not expected sql, query = %s", query)
	}
	exValue = []interface{}{1, "name1"}
	for index, value := range args {
		if exValue[index] != value {
			t.Errorf("args not expected value, args = %#v, value = %v", args, value)
		}
	}
}

func TestUpdate(t *testing.T) {
	query, args, err := psql.Update("test").
		Set("title", "ssss").
		Set("id", 1).
		Where(psql.Eq{"name": "sss", "type": []int{1, 2, 3}}).
		Where(psql.Or{psql.Eq{"name": "ccc"}, psql.And{psql.Eq{"id": 111}, psql.Eq{"desc": "sssss"}}}).
		Where(psql.Eq{"name": "dddd"}).
		Where(psql.And{psql.Eq{"type": 5}}).
		ToSql()
	if err != nil {
		t.Error(err)
	}

	exQuery := "UPDATE test " +
		"SET title=?,id=? " +
		"WHERE name = ? And type IN (?,?,?) AND (name = ? OR (id = ? AND desc = ?)) AND name = ? AND type = ?"
	if query != exQuery {
		t.Errorf("query not expected sql, query = %s", query)
	}
	exValue := []interface{}{"ssss", 1, "sss", 1, 2, 3, "ccc", 111, "sssss", "dddd", 5}
	for index, value := range args {
		if exValue[index] != value {
			t.Errorf("args not expected value, args = %#v, value = %v", args, value)
		}
	}

}

func TestDelete(t *testing.T) {
	query, args, err := psql.Delete("test").
		Where(psql.Eq{"name": "sss", "type": []int{1, 2, 3}}).
		Where(psql.Or{psql.Eq{"name": "ccc"}, psql.And{psql.Eq{"id": 111}, psql.Eq{"desc": "sssss"}}}).
		Where(psql.Eq{"name": "dddd"}).
		Where(psql.And{psql.Eq{"type": 5}}).
		ToSql()
	if err != nil {
		t.Error(err)
	}

	exQuery := "DELETE FROM test " +
		"WHERE name = ? And type IN (?,?,?) AND (name = ? OR (id = ? AND desc = ?)) AND name = ? AND type = ?"
	if query != exQuery {
		t.Errorf("query not expected sql, query = %s", query)
	}
	exValue := []interface{}{"sss", 1, 2, 3, "ccc", 111, "sssss", "dddd", 5}
	for index, value := range args {
		if exValue[index] != value {
			t.Errorf("args not expected value, args = %#v, value = %v", args, value)
		}
	}

}
