package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"rf4server/internal/store"
)

func (s *Server) handleGearDefinitions(w http.ResponseWriter,r *http.Request) {
	q:=strings.TrimSpace(r.URL.Query().Get("q"))
	limit,_:=strconv.Atoi(r.URL.Query().Get("limit"))
	rows,err:=s.db.SearchAdminGear(r.Context(),q,limit)
	if err!=nil { httpError(w,500,err.Error());return }
	writeJSON(w,rows)
}

func (s *Server) handleGearDefinitionSave(w http.ResponseWriter,r *http.Request) {
	var req struct { NumericID int64 `json:"numeric_id"`; AttrsJSON string `json:"attrs_json"` }
	if err:=readJSON(r,&req);err!=nil||req.NumericID<=0||strings.TrimSpace(req.AttrsJSON)=="" { httpError(w,400,"参数错误");return }
	if err:=s.db.SaveAdminGearAttrs(r.Context(),req.NumericID,req.AttrsJSON,adminOf(r));err!=nil { httpError(w,400,err.Error());return }
	_ = s.db.AuditLogWrite(r.Context(),adminOf(r),"gear_definition",strconv.FormatInt(req.NumericID,10),fmt.Sprintf("attrs_json bytes=%d",len(req.AttrsJSON)))
	writeJSON(w,map[string]any{"ok":true,"note":"已保存到运行时覆盖层；游戏实际读取物品定义时会应用该属性 JSON"})
}

func (s *Server) handleGearDefinitionReset(w http.ResponseWriter,r *http.Request) {
	var req struct { NumericID int64 `json:"numeric_id"` }
	if err:=readJSON(r,&req);err!=nil||req.NumericID<=0 { httpError(w,400,"参数错误");return }
	if err:=s.db.ResetAdminGearAttrs(r.Context(),req.NumericID);err!=nil { httpError(w,500,err.Error());return }
	_ = s.db.AuditLogWrite(r.Context(),adminOf(r),"gear_definition_reset",strconv.FormatInt(req.NumericID,10),"")
	writeJSON(w,map[string]any{"ok":true})
}

func (s *Server) handleCustomShopItemAdd(w http.ResponseWriter,r *http.Request) {
	var req struct { StoreID string `json:"store_id"`; ItemID string `json:"item_id"`; Silver int64 `json:"silver"`; Gold int64 `json:"gold"`; Stock int64 `json:"stock"` }
	if err:=readJSON(r,&req);err!=nil||strings.TrimSpace(req.StoreID)==""||strings.TrimSpace(req.ItemID)=="" { httpError(w,400,"参数错误");return }
	if err:=s.db.AddAdminCustomShopItem(r.Context(),req.StoreID,req.ItemID,req.Silver,req.Gold,req.Stock,adminOf(r));err!=nil { httpError(w,400,err.Error());return }
	_ = s.db.AuditLogWrite(r.Context(),adminOf(r),"shop_custom_add",req.StoreID+"/"+req.ItemID,fmt.Sprintf("silver=%d gold=%d stock=%d",req.Silver,req.Gold,req.Stock))
	writeJSON(w,map[string]any{"ok":true,"note":"商品已加入该店最终货架"})
}

func (s *Server) handleCustomShopItemRemove(w http.ResponseWriter,r *http.Request) {
	var req struct { StoreID string `json:"store_id"`; ItemID string `json:"item_id"` }
	if err:=readJSON(r,&req);err!=nil||strings.TrimSpace(req.StoreID)==""||strings.TrimSpace(req.ItemID)=="" { httpError(w,400,"参数错误");return }
	if err:=s.db.RemoveAdminCustomShopItem(r.Context(),req.StoreID,req.ItemID);err!=nil { httpError(w,500,err.Error());return }
	_ = s.db.AuditLogWrite(r.Context(),adminOf(r),"shop_custom_remove",req.StoreID+"/"+req.ItemID,"")
	writeJSON(w,map[string]any{"ok":true})
}

var _ = store.ShopPrice{}
