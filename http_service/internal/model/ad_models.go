/*
 * 廣告展示位模型。
 * 1. 保存公開列表右側廣告位與廣告素材的多選綁定。
 * 2. 支援每個頻道固定 10 個 slot，每個 slot 可配置多個展示廣告。
 * 3. 供公開接口按 slot 隨機挑選一個廣告輸出。
 */
package model

// 1. DisplayAdSlotAssignment stores one ad bound to one listing-side slot.
type DisplayAdSlotAssignment struct {
	ID               int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	DisplayChannel   string `gorm:"type:varchar(64);not null;uniqueIndex:uk_display_ad_slot_assignment,priority:1;index" json:"display_channel"`
	DisplayPlacement string `gorm:"type:varchar(64);not null;uniqueIndex:uk_display_ad_slot_assignment,priority:2;index" json:"display_placement"`
	SlotIndex        int    `gorm:"not null;uniqueIndex:uk_display_ad_slot_assignment,priority:3;index" json:"slot_index"`
	RewardAdID       int64  `gorm:"not null;uniqueIndex:uk_display_ad_slot_assignment,priority:4;index" json:"reward_ad_id"`
	SortOrder        int    `gorm:"not null;default:0" json:"sort_order"`
	DisplayTitle     string `gorm:"type:varchar(160);not null;default:''" json:"display_title"`
	DisplayText      string `gorm:"type:varchar(255);not null;default:''" json:"display_text"`
	TargetURL        string `gorm:"type:varchar(1024);not null;default:''" json:"target_url"`
	CreatedBy        *int64 `gorm:"index" json:"created_by"`
	UpdatedBy        *int64 `gorm:"index" json:"updated_by"`
	TimestampModel
}
