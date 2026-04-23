package models

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
