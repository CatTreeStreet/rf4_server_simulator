package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AdminGearDefinition is an editable view of a static item definition.
type AdminGearDefinition struct {
	NumericID   int64  `json:"numeric_id"`
	ItemID      string `json:"item_id"`
	Name        string `json:"name"`
	Category    int64  `json:"category"`
	Kind        string `json:"kind"`
	Brand       string `json:"brand"`
	AttrsJSON   string `json:"attrs_json"`
	HasOverride bool   `json:"has_override"`
}

func (d *DB) SearchAdminGear(ctx context.Context, q string, limit int) ([]*AdminGearDefinition, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	like := "%" + strings.TrimSpace(q) + "%"
	rows, err := d.SQL.QueryContext(ctx, `
		SELECT x.numeric_id, x.item_id, COALESCE(c.name_zh,c.name,''), COALESCE(x.category,0),
		       COALESCE(x.kind,''), COALESCE(x.brand,''), COALESCE(o.attrs_json,x.attrs_json,''),
		       CASE WHEN o.numeric_id IS NULL THEN 0 ELSE 1 END
		FROM item_def x LEFT JOIN item_catalog c ON c.item_id=x.item_id
		LEFT JOIN gear_def_override o ON o.numeric_id=x.numeric_id
		WHERE x.item_id LIKE ? OR COALESCE(c.name_zh,c.name,'') LIKE ? OR CAST(x.numeric_id AS TEXT)=?
		ORDER BY CASE WHEN x.item_id=? THEN 0 ELSE 1 END, x.numeric_id LIMIT ?`,
		like, like, strings.TrimSpace(q), strings.TrimSpace(q), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*AdminGearDefinition, 0)
	for rows.Next() {
		v := &AdminGearDefinition{}
		var has int
		if err := rows.Scan(&v.NumericID, &v.ItemID, &v.Name, &v.Category, &v.Kind, &v.Brand, &v.AttrsJSON, &has); err != nil {
			return nil, err
		}
		v.HasOverride = has != 0
		out = append(out, v)
	}
	return out, rows.Err()
}

func (d *DB) SaveAdminGearAttrs(ctx context.Context, numericID int64, attrsJSON, by string) error {
	if numericID <= 0 {
		return fmt.Errorf("numeric_id 必须大于 0")
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(attrsJSON), &obj); err != nil || obj == nil {
		return fmt.Errorf("属性必须是有效的 JSON 对象")
	}
	var exists int
	if err := d.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM item_def WHERE numeric_id=?`, numericID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("物品定义不存在: %d", numericID)
	}
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO gear_def_override(numeric_id,attrs_json,updated_ts,updated_by) VALUES(?,?,?,?)
		ON CONFLICT(numeric_id) DO UPDATE SET attrs_json=excluded.attrs_json,updated_ts=excluded.updated_ts,updated_by=excluded.updated_by`,
		numericID, attrsJSON, float64(time.Now().UnixNano())/1e9, by)
	return err
}

func (d *DB) ResetAdminGearAttrs(ctx context.Context, numericID int64) error {
	_, err := d.SQL.ExecContext(ctx, `DELETE FROM gear_def_override WHERE numeric_id=?`, numericID)
	return err
}

func (d *DB) applyGearDefinitionOverride(ctx context.Context, def *ItemDef) error {
	if def == nil {
		return nil
	}
	var attrs string
	err := d.SQL.QueryRowContext(ctx, `SELECT attrs_json FROM gear_def_override WHERE numeric_id=?`, def.NumericID).Scan(&attrs)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	def.AttrsJSON = attrs
	return nil
}

func (d *DB) addCustomShopRows(ctx context.Context, storeID string, base []*ShopPrice) ([]*ShopPrice, error) {
	rows, err := d.SQL.QueryContext(ctx, `SELECT item_id,silver,gold,stock FROM shop_item_custom WHERE store_id=? ORDER BY item_id`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := copyShopPrices(base)
	seen := make(map[string]bool, len(out))
	for _, p := range out {
		if p != nil {
			seen[p.ItemID] = true
		}
	}
	for rows.Next() {
		p := &ShopPrice{}
		if err := rows.Scan(&p.ItemID, &p.Silver, &p.Gold, &p.Stock); err != nil {
			return nil, err
		}
		if !seen[p.ItemID] {
			out = append(out, p)
			seen[p.ItemID] = true
		}
	}
	return out, rows.Err()
}

func (d *DB) AddAdminCustomShopItem(ctx context.Context, storeID, itemID string, silver, gold, stock int64, by string) error {
	storeID = strings.TrimSpace(storeID)
	itemID = strings.TrimSpace(itemID)
	if storeID == "" || itemID == "" {
		return fmt.Errorf("商店和商品不能为空")
	}
	if silver < 0 || gold < 0 || stock < 0 {
		return fmt.Errorf("价格和库存不能为负数")
	}
	var exists int
	if err := d.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM item_def WHERE item_id=?`, itemID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("商品定义不存在: %s", itemID)
	}
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO shop_item_custom(store_id,item_id,silver,gold,stock,updated_ts,updated_by) VALUES(?,?,?,?,?,?,?)
		ON CONFLICT(store_id,item_id) DO UPDATE SET silver=excluded.silver,gold=excluded.gold,stock=excluded.stock,updated_ts=excluded.updated_ts,updated_by=excluded.updated_by`,
		storeID, itemID, silver, gold, stock, float64(time.Now().UnixNano())/1e9, by)
	return err
}

func (d *DB) RemoveAdminCustomShopItem(ctx context.Context, storeID, itemID string) error {
	_, err := d.SQL.ExecContext(ctx, `DELETE FROM shop_item_custom WHERE store_id=? AND item_id=?`, strings.TrimSpace(storeID), strings.TrimSpace(itemID))
	return err
}
