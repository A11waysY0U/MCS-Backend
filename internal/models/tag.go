package models

type Tag struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	TagName   string `gorm:"unique" json:"tag_name"`
	CreaterID uint   `gorm:"not null" json:"creater_id"`
	Color     string `gorm:"default:'#409EFF';size:7" json:"color"`
}

type FileTag struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	TagID  uint `gorm:"index;not null" json:"tag_id"`
	FileID uint `gorm:"index;not null" json:"file_id"`
}
