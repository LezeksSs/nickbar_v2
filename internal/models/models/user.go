package models

import (
	"time"
)

type User struct {
	Id              string
	Nickname        string
	Register_date   time.Time
	Last_login_date time.Time
	Picture         string
	Role            int
}
