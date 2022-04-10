package service

import (
	"database/sql"
	"fmt"
	gorillactx "github.com/gorilla/context"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"time"
	"wbEmployeeApi/RFC7807"
	"wbEmployeeApi/model"
)

var Db *sql.DB

func bdErrProcessor(err error)(string,int){
	errText:="хранилище временно не доступно"
	errStatusCode:=500
	if pqerr, ok := err.(*pq.Error); ok {
		errCode,_:=strconv.Atoi(string(pqerr.Code))
		if errCode>50000{
			return err.Error(),400
		}
	}
	return errText,errStatusCode
}

func EmployeeAdd(employee model.Employee, r *http.Request)error{
	if r.Context().Err()!=nil{
		log.Println(time.Now(),"Контекст закрыт: ",r.Context().Err())
		return nil
	}
	startTime:=time.Now()
	_,err:=Db.Exec(`SELECT employees.employee_add($1,$2,$3,$4,$5,$6)`,employee.Name,employee.LastName,employee.Patronymic,
		employee.Phone,employee.Position,employee.GoodJobCount)
	if err!=nil{
		errText, errStatusCode:= bdErrProcessor(err)
		RFC7807.RequestSetError(r, "db-exec", errText, nil, errStatusCode)
		log.Println(time.Now(),": Ошибка при выполнении запроса employee_add к бд: ",err)
		return err
	}
	reqTime:=time.Since(startTime)
	gorillactx.Set(r,"db-time",reqTime.Seconds())
	gorillactx.Set(r,"db-request","EmployeeAdd")
	return nil
}

func EmployeeRemove(id int, r *http.Request)error{
	if r.Context().Err()!=nil{
		log.Println(time.Now(),"Контекст закрыт: ",r.Context().Err())
		return nil
	}
	startTime:=time.Now()
	_,err:=Db.Exec(`SELECT employees.employee_remove($1)`,id)
	if err!=nil{
		errText, errStatusCode:= bdErrProcessor(err)
		RFC7807.RequestSetError(r, "db-exec", errText, nil, errStatusCode)
		log.Println(time.Now(),"Ошибка при выполнении запроса employee_remove к бд: ",err)
		return err
	}
	reqTime:=time.Since(startTime)
	gorillactx.Set(r,"db-time",reqTime.Seconds())
	gorillactx.Set(r,"db-request","EmployeeRemove")
	return nil
}

func EmployeeUpdate(employee model.Employee,r *http.Request)error{
	if r.Context().Err()!=nil{
		log.Println(time.Now(),"Контекст закрыт: ",r.Context().Err())
		return nil
	}
	startTime:=time.Now()
	_,err:=Db.Exec(`SELECT employees.employee_upd($1,$2,$3,$4,$5,$6,$7)`,employee.Id,employee.Name,employee.LastName,employee.Patronymic,
		employee.Phone,employee.Position,employee.GoodJobCount)
	if err!=nil{
		errText, errStatusCode:= bdErrProcessor(err)
		RFC7807.RequestSetError(r, "db-exec", errText, nil, errStatusCode)
		log.Println(time.Now(),"Ошибка при выполнении запроса employee_upd к бд: ",err)
		return err
	}
	reqTime:=time.Since(startTime)
	gorillactx.Set(r,"db-time",reqTime.Seconds())
	gorillactx.Set(r,"db-request","EmployeeUpdate")
	return nil
}

func EmployeeGet(employeeId int,r *http.Request) error{
	var empl model.Employee
	startTime:=time.Now()
	row:=Db.QueryRow("SELECT * FROM employees.employee_get($1)", employeeId)
	err:=row.Scan(&empl.Name,&empl.LastName,&empl.Id,&empl.Patronymic,&empl.Phone,&empl.Position,&empl.GoodJobCount)
	if err!=nil{
		errText, errStatusCode:= bdErrProcessor(err)
		RFC7807.RequestSetError(r, "db-query", errText, nil, errStatusCode)
		log.Println(time.Now(),"Ошибка при выполнении запроса employee_get к бд: ",err)
		return err
	}
	reqTime:=time.Since(startTime)
	gorillactx.Set(r,"db-time",reqTime.Seconds())
	gorillactx.Set(r,"db-request","EmployeeGet")
	gorillactx.Set(r, "customdata",empl)
	if r.Context().Err()!=nil{
		log.Println(time.Now(),"Контекст закрыт: ",r.Context().Err())
		return r.Context().Err()
	}
	return nil
}

func EmployeeGetAllPartOne(r *http.Request)(map[int]model.EmployeePartOne, error){
	var emplID int
	var empl model.EmployeePartOne
	empls:=make(map[int]model.EmployeePartOne)
	startTime:=time.Now()
	rows, err := Db.Query(`SELECT * FROM employees.employees_get_all_part1()`)
	if err != nil {
		errText, errStatusCode := bdErrProcessor(err)
		RFC7807.RequestSetError(r, "db-query", errText, nil, errStatusCode)
		log.Println(time.Now(), "Ошибка при запросе employees.employees_get_all_part1 к БД: ", err)
		return nil,err
	}
	for rows.Next() {
		err = rows.Scan(&empl.Name, &empl.LastName, &emplID)
		if err != nil {
			log.Println(time.Now(), "При сканировании строки из БД в запросе employees.employees_get_all_part1 произошла ошибка: ", err)
		}else{
			empls[emplID]=empl
		}
	}
	reqTime:=time.Since(startTime)
	gorillactx.Set(r,"db-time",reqTime.Seconds())
	gorillactx.Set(r,"db-request","EmployeeGetAllPartOne")
	if r.Context().Err() != nil {
		log.Println(time.Now(), "Контекст закрыт: ", r.Context().Err())
		return nil,r.Context().Err()
	}
	return empls,nil
}

func EmployeeGetAllPartTwo(r *http.Request)(map[int]model.EmployeePartTwo,error){
	var emplID int
	var empl model.EmployeePartTwo
	empls:=make(map[int]model.EmployeePartTwo)
	startTime:=time.Now()
	rows, err := Db.Query(`SELECT * FROM employees.employees_get_all_part2()`)
	if err != nil {
		errText, errStatusCode := bdErrProcessor(err)
		RFC7807.RequestSetError(r, "db-query", errText, nil, errStatusCode)
		log.Println(time.Now(), "Ошибка при запросе employees.employees_get_all_part2 к БД: ", err)
		return nil,err
	}
	for rows.Next() {
		err = rows.Scan(&emplID, &empl.Patronymic, &empl.Phone, &empl.Position, &empl.GoodJobCount)
		if err != nil {
			log.Println(time.Now(), "При сканировании строки из БД в запросе employees.employees_get_all_part2 произошла ошибка: ", err)
		}else{
			empls[emplID] = empl
		}
	}
	reqTime:=time.Since(startTime)
	gorillactx.Set(r,"db-time",reqTime.Seconds())
	gorillactx.Set(r,"db-request","EmployeeGetAllPartTwo")
	if r.Context().Err() != nil {
		log.Println(time.Now(), "Контекст закрыт: ", r.Context().Err())
		return nil,r.Context().Err()
	}
	return empls,nil
}

func DBInitialize(config model.DBConfig) (*sql.DB,error){
	initString:=fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)
	db, err:=sql.Open("postgres",initString)
	if err!=nil{
		log.Fatal("Ошибка при открытии бд: ",err)
	}
	err=db.Ping()
	if err!=nil{
		log.Fatal("Ошибка при пинге бд: ",err)
	}
	return db,nil
}
