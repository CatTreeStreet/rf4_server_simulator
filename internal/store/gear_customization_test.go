package store

import (
	"context"
	"testing"
)

func TestAdminGearDefinitionOverrideIsUsedByRuntimeReads(t *testing.T) {
	db := openShopCatalogTestDB(t)
	ctx := context.Background()
	if _, err := db.SQL.Exec(`INSERT INTO item_def(numeric_id,item_id,category,kind,brand,attrs_json)
		VALUES(990101,'custom_rod_test',13,'rod','test','{"original":1}')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL.Exec(`INSERT INTO item_catalog(item_id,name,name_zh,kind,numeric_id)
		VALUES('custom_rod_test','Test Rod','测试鱼竿','rod',990101)`); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveAdminGearAttrs(ctx, 990101, `{"custom":2,"kngaaefkpka":1148846080}`, "test"); err != nil {
		t.Fatal(err)
	}
	def, err := db.GetItemDef(ctx, 990101)
	if err != nil {
		t.Fatal(err)
	}
	if def.AttrsJSON != `{"custom":2,"kngaaefkpka":1148846080}` {
		t.Fatalf("runtime definition did not use override: %s", def.AttrsJSON)
	}
	if err := db.ResetAdminGearAttrs(ctx, 990101); err != nil {
		t.Fatal(err)
	}
	def, err = db.GetItemDefByItemID(ctx, "custom_rod_test")
	if err != nil {
		t.Fatal(err)
	}
	if def.AttrsJSON != `{"original":1}` {
		t.Fatalf("reset did not restore base attrs: %s", def.AttrsJSON)
	}
}

func TestAdminCanAddArbitraryCatalogItemToShopShelf(t *testing.T) {
	db := openShopCatalogTestDB(t)
	ctx := context.Background()
	if _, err := db.SQL.Exec(`INSERT INTO item_def(numeric_id,item_id,category,kind,brand,attrs_json)
		VALUES(990102,'custom_rod_test_shop',13,'rod','test','{}')`); err != nil {
		t.Fatal(err)
	}
	if err := db.AddAdminCustomShopItem(ctx, "fishingstore_006", "custom_rod_test_shop", 12345, 7, 88, "test"); err != nil {
		t.Fatal(err)
	}
	p, err := db.ShopItemGet(ctx, "fishingstore_006", "custom_rod_test_shop")
	if err != nil {
		t.Fatal(err)
	}
	if p.Silver != 12345 || p.Gold != 7 || p.Stock != 88 {
		t.Fatalf("custom shelf row mismatch: %+v", p)
	}
	if err := db.RemoveAdminCustomShopItem(ctx, "fishingstore_006", "custom_rod_test_shop"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ShopItemGet(ctx, "fishingstore_006", "custom_rod_test_shop"); err == nil {
		t.Fatal("removed custom item still present in shop")
	}
}
