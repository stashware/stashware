package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db        *sql.DB  = nil
	dbx       *sqlx.DB = nil
	ErrNoRows          = sql.ErrNoRows
)

func DbExec(sqlstr string) (err error) {
	_, err = db.Exec(sqlstr)

	return
}

func DbGet(sqlstr string, parse interface{}) (err error) {
	err = dbx.Get(parse, sqlstr)

	return
}

func DbGetInt64(sqlstr string) (num int64, err error) {
	err = dbx.Get(&num, sqlstr)

	return
}

func DbSelect(sqlstr string, parse interface{}) (err error) {
	err = dbx.Select(parse, sqlstr)

	return
}

func initSqlx(db_path string) (err error) {
	db, err = sql.Open("sqlite3", db_path)
	if nil != err {
		return
	}
	dbx = sqlx.NewDb(db, "sqlite3")

	return
}

func initTables() (err error) {
	slqstr := `

--
-- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
-- 
CREATE TABLE IF NOT EXISTS tb_blocks (
    indep_hash          VARCHAR(256) PRIMARY KEY,
    hash                VARCHAR(256),
    height              INTEGER,
    previous_block      VARCHAR(256),
    nonce               VARCHAR(256),
    timestamp           INTEGER,
    last_retarget       INTEGER,            -- 
    diff                VARCHAR(256),
    cumulative_diff     INTEGER,
    reward_addr         VARCHAR(256),
    reward_fee          INTEGER,
    reward_pool         INTEGER,
    weave_size          INTEGER,            -- 
    block_size          INTEGER,            -- 
    txs                 TEXT,               -- 
    tx_root             VARCHAR(256),
    wallet_list         TEXT,               --  
    wallet_list_hash    VARCHAR(256),
    network             INTEGER             --  
);
CREATE INDEX idx_blocks_height ON tb_blocks (height);

--
-- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
-- 
CREATE TABLE IF NOT EXISTS tb_transactions (
    row                 INTEGER PRIMARY KEY AUTOINCREMENT,
    ctime               TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id                  VARCHAR(256) UNIQUE,
    block_indep_hash    VARCHAR(256) DEFAULT "",
    last_tx             VARCHAR(256) DEFAULT "",
    owner               VARCHAR(2000) DEFAULT "",      -- sender pub_key
    from_address        VARCHAR(256) DEFAULT "",       -- sender address
    target              VARCHAR(256) DEFAULT "",       -- receiver address
    quantity            INTEGER DEFAULT 0,             -- send amount
    signature           VARCHAR(2000) DEFAULT "",
    reward              INTEGER DEFAULT 0,
    tags                VARCHAR(2000) DEFAULT "",
    data_hash           VARCHAR(256) DEFAULT ""
);
CREATE INDEX idx_tx_block_indep_hash ON tb_transactions (block_indep_hash);
CREATE INDEX idx_tx_from_address ON tb_transactions (from_address);

--
-- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
-- 
CREATE TABLE IF NOT EXISTS tb_wallet (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    address             VARCHAR(256)  NOT NULL UNIQUE,
    last_tx             VARCHAR(256),                       -- tx id 
    balance             INTEGER
);


--
-- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
-- 
CREATE TABLE IF NOT EXISTS tb_chain (
    indep_hash          VARCHAR(256) PRIMARY KEY,
    hash_list           TEXT,               -- encode
    height              INTEGER,
    is_active           INTEGER
);


--
-- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
-- 
CREATE TABLE IF NOT EXISTS tb_pool (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    tx_id               VARCHAR(256) UNIQUE,
    from_address        VARCHAR(256) DEFAULT "",
    target              VARCHAR(256) DEFAULT "",
    last_tx             VARCHAR(256) DEFAULT "",
    ctime               TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
	`
	return DbExec(slqstr)
}
