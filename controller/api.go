package controller

import (
	"encoding/json"
	gorillactx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"wbEmployeeApi/model"
	"wbEmployeeApi/service"
)

func EmployeeAddHandler(w http.ResponseWriter, r *http.Request){
	var empl model.Employee
	body,err:=ioutil.ReadAll(r.Body)
	if err!=nil{
		log.Println("Ошибка при чтении тела запроса:",err)
		return
	}
	err=json.Unmarshal(body,&empl)
	if err!=nil{
		log.Println("Ошибка при анмаршале(EmployeeAddHandler):",err)
		return
	}
	err= service.EmployeeAdd(empl, r)
	if err==nil{
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("201 сreated"))
	}
	return
}

func EmployeeRemoveHandler(w http.ResponseWriter, r *http.Request){
	var id int
	body,err:=ioutil.ReadAll(r.Body)
	if err!=nil{
		log.Println("Ошибка при чтении тела запроса:",err)
		return
	}
	err=json.Unmarshal(body,&id)
	if err!=nil{
		log.Println("Ошибка при анмаршале(EmployeeRemoveHandler):",err)
		return
	}
	err= service.EmployeeRemove(id, r)
	if err==nil{
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 deleted"))
	}
}

func EmployeeUpdateHandler(w http.ResponseWriter, r *http.Request){
	var empl model.Employee
	body,err:=ioutil.ReadAll(r.Body)
	if err!=nil{
		log.Println("Ошибка при чтении тела запроса:",err)
		return
	}
	err=json.Unmarshal(body,&empl)
	if err!=nil{
		log.Println("Ошибка при анмаршале(EmployeeUpdateHandler):",err)
		return
	}
	err= service.EmployeeUpdate(empl,r)
	if err==nil{
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 updated"))
	}
}

func EmloyeeGetAllHandler(w http.ResponseWriter, r* http.Request){
	employeesAllPart1:=make(map[int]model.EmployeePartOne)
	employeesAllPart2:=make(map[int]model.EmployeePartTwo)
	var employeesAllResult []model.Employee
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		employeesAllPart1,err=service.EmployeeGetAllPartOne(r)
		wg.Done()
	}()
	if err!=nil{
		return
	}
	wg.Add(1)
	go func(){
		employeesAllPart2,err=service.EmployeeGetAllPartTwo(r)
		wg.Done()
	}()
	if err!=nil{
		return
	}
	wg.Wait()
	for id,_:=range employeesAllPart1{
		if _,ok:=employeesAllPart2[id];ok==false{
			continue
		}
		tmpEmpl:=model.Employee{
			id,
			employeesAllPart1[id].LastName,
			employeesAllPart1[id].Name,
			employeesAllPart2[id].Patronymic,
			employeesAllPart2[id].Phone,
			employeesAllPart2[id].Position,
			employeesAllPart2[id].GoodJobCount,
		}
		employeesAllResult=append(employeesAllResult,tmpEmpl)
	}
	gorillactx.Set(r,"customdata",employeesAllResult)
	return
}

func EmployeeGetHandler(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	var err error
	employeeId,err:=strconv.Atoi(vars["employeeId"])
	if err!=nil{
		log.Println("Ошибка при конвертации STR в INT:",err)
		return
	}
	service.EmployeeGet(employeeId, r)
	return
}