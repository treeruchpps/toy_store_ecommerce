package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	*sqlx.DB
	dsn string
}

func NewPostgresDB(dataSourceName string) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// กำหนดค่า connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// ทดสอบการเชื่อมต่อ
	if err = db.PingContext(ctx); err != nil {
		db.Close() // ปิดการเชื่อมต่อถ้าไม่สามารถ ping ได้
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{
		DB:  db,
		dsn: dataSourceName,
	}, nil
}

func (db *PostgresDB) Reconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newDB, err := sqlx.ConnectContext(ctx, "postgres", db.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// ตั้งค่า connection pool
	newDB.SetMaxOpenConns(25)
	newDB.SetMaxIdleConns(10)
	newDB.SetConnMaxLifetime(5 * time.Minute)

	// ทดสอบการเชื่อมต่อ
	if err = newDB.PingContext(ctx); err != nil {
		newDB.Close() // ปิดการเชื่อมต่อใหม่ถ้าไม่สามารถ ping ได้
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// ปิดการเชื่อมต่อเดิม (ถ้ามี) และกำหนดการเชื่อมต่อใหม่
	if db.DB != nil {
		db.DB.Close()
	}
	db.DB = newDB

	return nil
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}
