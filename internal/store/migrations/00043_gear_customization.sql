-- 通用装备定义覆盖与商店自定义上架。静态库继续只读，所有编辑保存在主库。
CREATE TABLE IF NOT EXISTS gear_def_override(
    numeric_id INTEGER PRIMARY KEY,
    attrs_json TEXT NOT NULL,
    updated_ts REAL NOT NULL DEFAULT 0,
    updated_by TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS shop_item_custom(
    store_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    silver INTEGER NOT NULL DEFAULT 0,
    gold INTEGER NOT NULL DEFAULT 0,
    stock INTEGER NOT NULL DEFAULT 1000,
    updated_ts REAL NOT NULL DEFAULT 0,
    updated_by TEXT NOT NULL DEFAULT '',
    PRIMARY KEY(store_id,item_id)
);
CREATE INDEX IF NOT EXISTS idx_shop_item_custom_store ON shop_item_custom(store_id);
