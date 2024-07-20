package models

type Signupusers struct {
	Fullname string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
	Username string `gorm:"not null"`
	Password string `gorm:"not null"`
}

type Invalidsignup struct {
	InvalidFullname string
	InvalidEmail    string
	InvalidUsername string
	InvalidPassword string
}

type Invalidlogin struct {
	Username string
	Password string
}

type Home struct {
	Username string
}

type Searchdata struct {
	Users       []Signupusers
	SearchError string
}

type UserUpdate struct {
	Error    Invalidsignup
	Userdata Signupusers
}
