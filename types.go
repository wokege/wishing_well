package main

type User struct {
	ID        int    `gorm:"primaryKey"`
	DiscordId uint64 `gorm:"column:discord_id"`
}

func (User) TableName() string {
	return "users"
}

type Log struct {
	ID        int    `gorm:"primaryKey,autoIncrement"`
	UserId    int    `gorm:"column:user_id"`
	MessageId uint64 `gorm:"column:message_id"`
	Count     int
}

func (Log) TableName() string {
	return "log"
}
