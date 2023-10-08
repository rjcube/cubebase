package cubebase

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/rjcube/cubebase/pg"
)

type CubeContext struct {
	Db  *sqlx.DB
	Rdb *redis.Client
	Ctx context.Context
}

type UserInfo struct {
	ID              int64          `json:"id"`
	AccountName     string         `json:"account_name"`
	AccountNickname *string        `json:"account_nickname"`
	EmployeeName    string         `json:"employee_name"`
	EmployeeCode    string         `json:"employee_code"`
	RoleIds         pg.Int64Array  `json:"role_ids"`
	Remark          sql.NullString `json:"remark"`
	IsDelete        sql.NullBool   `json:"is_delete"`
	IsCreate        sql.NullBool   `json:"is_create"`
}
