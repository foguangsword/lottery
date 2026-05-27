-- =============================================
-- Lottery 项目数据库初始化脚本
-- =============================================
-- 本脚本兼容 GORM（使用 gorm.Model 的默认字段）
-- 首次启动 MySQL 容器时自动执行
-- =============================================

CREATE DATABASE IF NOT EXISTS lottery_db
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE lottery_db;

-- ---------------------------------------------
-- 用户表
-- ---------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at      DATETIME(3) NULL,
    updated_at      DATETIME(3) NULL,
    deleted_at      DATETIME(3) NULL,

    username        VARCHAR(64)  NOT NULL UNIQUE COMMENT '用户名',
    email           VARCHAR(128) NOT NULL COMMENT '邮箱',
    password        VARCHAR(256) NOT NULL COMMENT 'bcrypt 哈希后的密码',
    eth_address     VARCHAR(42)  NOT NULL UNIQUE COMMENT '后端托管的以太坊地址',

    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ---------------------------------------------
-- 活动表（缓存链上数据，便于前端列表查询）
-- ---------------------------------------------
CREATE TABLE IF NOT EXISTS activities (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at      DATETIME(3) NULL,
    updated_at      DATETIME(3) NULL,
    deleted_at      DATETIME(3) NULL,

    contract_name   VARCHAR(128) NOT NULL UNIQUE COMMENT '对应合约中的活动名',
    display_name    VARCHAR(256) NOT NULL COMMENT '展示名称',
    description     TEXT COMMENT '活动描述',
    start_time      DATETIME     NOT NULL COMMENT '报名开始时间',
    end_time        DATETIME     NOT NULL COMMENT '报名截止时间',
    lucky_count     INT          NOT NULL COMMENT '中奖人数',
    status          ENUM('active','drawn','canceled') NOT NULL DEFAULT 'active' COMMENT '活动状态',
    create_tx       VARCHAR(66) COMMENT 'createActivity 交易 hash',

    INDEX idx_activities_status (status),
    INDEX idx_activities_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='活动表';

-- ---------------------------------------------
-- 报名记录表
-- ---------------------------------------------
CREATE TABLE IF NOT EXISTS registrations (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at      DATETIME(3) NULL,
    updated_at      DATETIME(3) NULL,
    deleted_at      DATETIME(3) NULL,

    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    activity_id     BIGINT UNSIGNED NOT NULL COMMENT '活动ID',
    chain_index     INT             NOT NULL COMMENT '合约返回的参与者序号',
    tx_hash         VARCHAR(66) COMMENT 'participate 交易 hash',

    UNIQUE KEY uk_user_activity (user_id, activity_id),
    INDEX idx_registrations_deleted_at (deleted_at),

    FOREIGN KEY (user_id)     REFERENCES users(id),
    FOREIGN KEY (activity_id) REFERENCES activities(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报名记录表';

-- ---------------------------------------------
-- 开奖结果表
-- ---------------------------------------------
CREATE TABLE IF NOT EXISTS draw_results (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at      DATETIME(3) NULL,
    updated_at      DATETIME(3) NULL,
    deleted_at      DATETIME(3) NULL,

    activity_id     BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '活动ID',
    seed            VARCHAR(32)   NOT NULL COMMENT '开奖 seed',
    shanghai_index  DECIMAL(10,2) NOT NULL COMMENT '上证收盘价',
    shenzhen_index  DECIMAL(10,2) NOT NULL COMMENT '深证收盘价',
    block_number    BIGINT        NOT NULL COMMENT 'drawWinner 交易所在区块',
    block_ts        BIGINT        NOT NULL COMMENT '区块时间戳',
    tx_hash         VARCHAR(66) COMMENT 'drawWinner 交易 hash',

    INDEX idx_draw_results_deleted_at (deleted_at),
    FOREIGN KEY (activity_id) REFERENCES activities(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='开奖结果表';

-- ---------------------------------------------
-- 中奖者表
-- ---------------------------------------------
CREATE TABLE IF NOT EXISTS winners (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at      DATETIME(3) NULL,
    updated_at      DATETIME(3) NULL,
    deleted_at      DATETIME(3) NULL,

    activity_id     BIGINT UNSIGNED NOT NULL COMMENT '活动ID',
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    win_order       INT             NOT NULL COMMENT '中奖顺序 (0, 1, 2...)',

    INDEX idx_winners_deleted_at (deleted_at),
    FOREIGN KEY (activity_id) REFERENCES activities(id),
    FOREIGN KEY (user_id)     REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='中奖者表';
