package models

type Customer struct{
	Email		string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password	string `gorm:"not null"`
	CustomerId	string `gorm:"not null"`
}
type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Register struct {
	Email     		string `json:"email"`
	Name  			string `json:"name"`
	Password  		string `json:"password"`
	Number	  		string `json:"number"`
}


