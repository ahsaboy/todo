// migrate_time_format.go
//
// 将数据库中旧格式时间（YYYY-MM-DD HH:MM:SS，视为 UTC）迁移为 RFC3339 UTC。
//
// 用法：
//
//	go run scripts/migrate_time_format.go -db data/tasks.db [-dry-run]
//
// 执行前务必备份数据库文件。

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var timeColumns = []struct {
	table string
	col   string
}{
	{"tasks", "due_at"},
	{"tasks", "remind_at"},
	{"tasks", "repeat_end_date"},
	{"tasks", "reminder_sent_at"},
	{"tasks", "created_at"},
	{"tasks", "updated_at"},
	{"users", "created_at"},
	{"users", "updated_at"},
	{"user_api_keys", "created_at"},
	{"user_api_keys", "last_used_at"},
	{"user_reminder_configs", "created_at"},
	{"user_reminder_configs", "updated_at"},
}

func main() {
	dbPath := flag.String("db", "data/tasks.db", "SQLite 数据库路径")
	dryRun := flag.Bool("dry-run", false, "仅打印将要执行的更新，不实际写入")
	flag.Parse()

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	totalMigrated := 0
	for _, tc := range timeColumns {
		n, err := migrateColumn(db, tc.table, tc.col, *dryRun)
		if err != nil {
			log.Printf("WARNING: %s.%s: %v", tc.table, tc.col, err)
			continue
		}
		totalMigrated += n
	}

	action := "would be migrated"
	if !*dryRun {
		action = "migrated"
	}
	fmt.Printf("Done. %d rows %s.\n", totalMigrated, action)
}

func migrateColumn(db *sql.DB, table, col string, dryRun bool) (int, error) {
	// 查询所有非空、非 RFC3339 格式的值
	rows, err := db.Query(fmt.Sprintf(
		`SELECT rowid, "%s" FROM "%s" WHERE "%s" IS NOT NULL AND "%s" != '' AND "%s" NOT LIKE '%%T%%Z'`,
		col, table, col, col, col,
	))
	if err != nil {
		return 0, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	type row struct {
		id  int64
		val string
	}
	var candidates []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.val); err != nil {
			return 0, fmt.Errorf("scan: %w", err)
		}
		// 跳过已经是 RFC3339 的值
		if t, err := time.Parse(time.RFC3339, r.val); err == nil && t.Location() != time.UTC {
			// 有偏移量但非 Z 结尾，也算 RFC3339
			continue
		}
		candidates = append(candidates, r)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("rows: %w", err)
	}

	count := 0
	for _, c := range candidates {
		t, err := time.Parse("2006-01-02 15:04:05", c.val)
		if err != nil {
			// 可能已经是 RFC3339，跳过
			continue
		}
		newVal := t.UTC().Format(time.RFC3339)
		if dryRun {
			fmt.Printf("[dry-run] %s.%s rowid=%d: %q -> %q\n", table, col, c.id, c.val, newVal)
		} else {
			stmt := fmt.Sprintf(`UPDATE "%s" SET "%s" = ? WHERE rowid = ?`, table, col)
			if _, err := db.Exec(stmt, newVal, c.id); err != nil {
				log.Printf("WARNING: update %s.%s rowid=%d: %v", table, col, c.id, err)
				continue
			}
			fmt.Printf("[migrated] %s.%s rowid=%d: %q -> %q\n", table, col, c.id, c.val, newVal)
		}
		count++
	}
	return count, nil
}
