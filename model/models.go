package model

type Employee struct{
	Id int				`json:"id,omitempty"`
	LastName string		`json:"lastName"`
	Name string			`json:"name"`
	Patronymic string	`json:"patronymic"`
	Phone string		`json:"phone"`
	Position string		`json:"position"`
	GoodJobCount int	`json:"goodJobCount"`
}

type EmployeePartOne struct {
	LastName string		`json:"lastName"`
	Name string			`json:"name"`
}

type EmployeePartTwo struct {
	Patronymic string	`json:"patronymic"`
	Phone string		`json:"phone"`
	Position string		`json:"position"`
	GoodJobCount int	`json:"goodJobCount"`
}

type DBConfig struct {
	Host string
	Port string
	DBName string
	User string
	Password string
}
