/*
 * 首頁內容模型。
 * 1. 保存首頁輪播圖片與三個主模組大圖配置。
 * 2. 綁定已完成上傳的媒體資產，供公開首頁讀取。
 */
package model

// 1. HomeContentPlacement stores one homepage media/content slot.
type HomeContentPlacement struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	PlacementType string `gorm:"type:varchar(32);not null;uniqueIndex:uk_home_content_position,priority:1;index" json:"placement_type"`
	ModuleCode    string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_home_content_position,priority:2;index" json:"module_code"`
	SlotIndex     int    `gorm:"not null;uniqueIndex:uk_home_content_position,priority:3" json:"slot_index"`
	MediaAssetID  int64  `gorm:"not null;index" json:"media_asset_id"`
	Title         string `gorm:"type:varchar(160)" json:"title"`
	Subtitle      string `gorm:"type:varchar(240)" json:"subtitle"`
	Body          string `gorm:"type:text" json:"body"`
	CreatedBy     *int64 `gorm:"index" json:"created_by"`
	UpdatedBy     *int64 `gorm:"index" json:"updated_by"`
	TimestampModel
}
