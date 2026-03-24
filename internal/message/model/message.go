package model

type Message struct {
	ID         int64  `gorm:"primaryKey;autoIncrement:false;comment:'Snowflake ID'"`
	ToUserID   int64  `gorm:"index;not null;comment:'Receiver ID'"`
	FromUserID int64  `gorm:"index;not null;comment:'Sender ID'"`
	Content    string `gorm:"type:text;not null;comment:'Message content'"`
	CreateTime int64  `gorm:"autoCreateTime:milli;comment:'Creation time in ms'"`
}

func (m *Message) TableName() string {
	return "messages"
}
