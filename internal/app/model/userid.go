package model

type UserID string

func NewUserID(userId string) UserID {
	return UserID(userId)
}
