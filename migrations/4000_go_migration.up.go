package migrations

import (
	"fmt"
	"time"
)

func (m *TestGoMigrationsRunner) Up_4000(tableName, metadata string) error {
	type info struct {
		id     int
		tstamp time.Time
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()

	_, err = tx.Exec("ALTER TABLE " + tableName + " ADD COLUMN name VARCHAR, ADD COLUMN metadata VARCHAR")
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
		tx.Exec(`UPDATE some_table SET name=$1, metadata=$2 WHERE id=$3`, name, metadata, info.id)
	}

	return nil
}
