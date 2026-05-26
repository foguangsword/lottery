package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex;size:64"`
	Email    string `json:"email" gorm:"size:128"`
	Address  string `json:"address" gorm:"size:128"` // 链上地址
}
