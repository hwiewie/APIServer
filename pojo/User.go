package pojo

import "github.com/hwiewie/APIServer/database"

type User struct {
	Id       int    `json:"UserId"`
	Name     string `json:"UserName"`
	Password string `json:"UserPassword"`
	Email    string `json:"UserEmail"`
}

func FindAllUsers() []User {
	var users []User
	database.DBconnect.Find(&users)
	return users
}

func FindByUserId(userId int) User {
	var user User
	database.DBconnect.Where("id = ?", userId).First(&user)
	return user
}
