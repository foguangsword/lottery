package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username   string `json:"username" gorm:"uniqueIndex;size:64;not null"`
	Email      string `json:"email" gorm:"size:128;not null"`
	Password   string `json:"-" gorm:"size:256;not null"`
	EthAddress string `json:"address" gorm:"size:42;not null"` // 链上地址
}
