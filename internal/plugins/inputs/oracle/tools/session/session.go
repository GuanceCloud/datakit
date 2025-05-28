// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/sijms/go-ora/v2" // 或你使用的其他 Oracle 驱动
)

var (
	flagDBConnectionStrings = flag.String("conn", "", "Oracle database connection string, example: oracle://system:1234@localhost:1521/XEPDB1")
	flagNumSessions         = flag.Int("session-count", 20, "")                     // 要创建的会话数量
	flagLoopDelay           = flag.Duration("loop-delay", 500*time.Millisecond, "") // 每个会话循环执行SQL的间隔
)

// sessionWorker 代表一个独立的数据库会话执行单元.
func sessionWorker(id int, wg *sync.WaitGroup, db *sqlx.DB, shutdown chan struct{}) {
	defer wg.Done()
	log.Printf("Session Worker %d: Starting\n", id)

	// 为这个 worker 获取一个连接，并保持它直到 worker 结束.
	// 注意：长时间持有连接可能不是连接池的最佳实践，但对于模拟独立长会话是合适的.
	conn, err := db.Connx(context.Background())
	if err != nil {
		log.Printf("Session Worker %d: Failed to get connection: %v\n", id, err)
		return
	}
	defer conn.Close() // nolint:errcheck
	log.Printf("Session Worker %d: Connection acquired.\n", id)

	// 模拟 CPU 消耗的 SQL.
	cpuSQL := `
DECLARE
  v_dummy NUMBER;
  v_iterations NUMBER := DBMS_RANDOM.VALUE(40000, 60000); -- 随机迭代次数
BEGIN
  FOR i IN 1..v_iterations LOOP
    v_dummy := SIN(i) * COS(i) + POWER(i, 2);
  END LOOP;
END;`

	// 模拟 PGA 和一些逻辑读的 SQL.
	// 确保 session_test_data 表存在并有少量数据.
	// 或者替换为查询数据字典等操作.
	pgaSQL := `
SELECT id, val, object_name, object_type
FROM (
    SELECT d.id, d.val, o.object_name, o.object_type, ROW_NUMBER() OVER (ORDER BY DBMS_RANDOM.VALUE) as rn
    FROM (
        SELECT LEVEL AS id, DBMS_RANDOM.STRING('X', 30) AS val
        FROM DUAL CONNECT BY LEVEL <= 200 -- 生成少量动态数据
    ) d
    CROSS JOIN (
        SELECT object_name, object_type FROM all_objects WHERE ROWNUM <= 10 -- 与少量字典对象交叉
    ) o
    WHERE ROWNUM <= 200 -- 限制最终结果集大小
)
ORDER BY val DESC, id ASC`

	// 简单的查询.
	dictSQL := `SELECT COUNT(*) FROM all_tab_columns WHERE ROWNUM <= DBMS_RANDOM.VALUE(500,1500)`

	ticker := time.NewTicker(*flagLoopDelay)
	defer ticker.Stop()

	for {
		select {
		case <-shutdown:
			log.Printf("Session Worker %d: Shutting down\n", id)
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 为每个SQL操作设置超时.

			log.Printf("Session Worker %d: Executing CPU-intensive SQL...\n", id)
			_, err := conn.ExecContext(ctx, cpuSQL)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.Printf("Session Worker %d: CPU SQL timed out: %v\n", id, err)
				} else {
					log.Printf("Session Worker %d: Error executing CPU SQL: %v\n", id, err)
				}
				cancel() // 确保在出错时也调用 cancel.
				// 如果连接因错误失效，可能需要重新获取连接或终止 worker.
				// 对于简单模拟，我们继续尝试.
				// time.Sleep(5 * time.Second) // 发生错误后稍作等待
				// continue
				return // 或者直接退出worker.
			}

			log.Printf("Session Worker %d: Executing PGA/IO-intensive SQL...\n", id)
			var dummyResults []struct { // 定义一个结构体来接收结果，即使我们不直接使用它们.
				ID         int            `db:"ID"`
				Val        string         `db:"VAL"`
				ObjectName sql.NullString `db:"OBJECT_NAME"`
				ObjectType sql.NullString `db:"OBJECT_TYPE"`
			}
			err = conn.SelectContext(ctx, &dummyResults, pgaSQL)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.Printf("Session Worker %d: PGA SQL timed out: %v\n", id, err)
				} else {
					log.Printf("Session Worker %d: Error executing PGA SQL: %v\n", id, err)
				}
				cancel()
				return
			}

			log.Printf("Session Worker %d: Executing Dictionary SQL...\n", id)
			var count int
			err = conn.GetContext(ctx, &count, dictSQL)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.Printf("Session Worker %d: Dictionary SQL timed out: %v\n", id, err)
				} else {
					log.Printf("Session Worker %d: Error executing Dictionary SQL: %v\n", id, err)
				}
				cancel()
				return
			}
			log.Printf("Session Worker %d: Dictionary query returned count: %d. Loop finished.\n", id, count)

			cancel() // 确保在每次循环结束时调用 cancel.
		}
	}
}

func main() { // nolint: typecheck
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	log.Println("Starting Oracle session load generator...")

	conns := strings.Split(*flagDBConnectionStrings, ",")
	var wg sync.WaitGroup
	shutdown := make(chan struct{}) // 用于通知 goroutines 关闭

	for _, conn := range conns {
		db, err := sqlx.Connect("oracle", conn)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			return
		}
		defer db.Close() // nolint:errcheck

		// 设置连接池参数 (可选，但对于多goroutine可能重要)
		db.SetMaxOpenConns(*flagNumSessions + 5) // 允许比 worker 数量稍多的连接
		db.SetMaxIdleConns(*flagNumSessions)
		db.SetConnMaxLifetime(time.Hour)

		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping database: %v", err)
			return
		}
		log.Println("Successfully connected and pinged database.")

		for i := 0; i < *flagNumSessions; i++ {
			wg.Add(1)
			go sessionWorker(i+1, &wg, db, shutdown)
			time.Sleep(100 * time.Millisecond) // 稍微错开启动，避免瞬间冲击
		}
	}

	log.Printf("%d session workers started. Press Ctrl+C to exit.\n", *flagNumSessions)

	// 捕获中断信号以实现优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // 等待信号

	log.Println("Shutdown signal received. Closing workers...")
	close(shutdown) // 通知所有 worker 关闭
	wg.Wait()       // 等待所有 worker 完成
	log.Println("All session workers finished.")
	log.Println("Program exiting.")
}
