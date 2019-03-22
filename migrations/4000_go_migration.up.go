package migrations

import (
	"fmt"
)

func (m *TestGoMigrationsRunner) Up_4000() error {
	type info struct {
		id int
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() { _ = tx.Commit() }()

	_, err = tx.Exec("ALTER TABLE some_table ADD COLUMN name VARCHAR, ADD COLUMN metadata VARCHAR")
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM some_table")
	if err != nil {
		return err
	}

	infos := []info{}

	for rows.Next() {
		info := info{}
		err = rows.Scan(&info.id)
		if err != nil {
			return err
		}
		infos = append(infos, info)
	}

	for _, info := range infos {
		name := fmt.Sprintf("name_%v", info.id)
		_, err = tx.Exec(`UPDATE some_table SET name=$1, metadata=$2 WHERE id=$3`, name, m.metadata, info.id)
		if err != nil {
			return err
		}
	}

	return nil
}
