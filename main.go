package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"wbEmployeeApi/RFC7807"
	"wbEmployeeApi/controller"
	"wbEmployeeApi/service"
)

const(
	APIName="employees"
	APIVersion="1.0.0"
)

var db *sql.DB
var cfg *service.Config

func APIInfoHandler(w http.ResponseWriter, r *http.Request){
	info:=make(map[string]string)
	info["name"]=APIName
	info["version"]=APIVersion
	jsonBody,err:=json.Marshal(info)
	if err!=nil{
		log.Println("Ошибка при маршале JSON, APIInfoHandler:",err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Метод не робит("))
	}else{
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBody)
	}
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func main(){
	os.Setenv("user_employees","postgresadmin")
	os.Setenv("pass_employees","admin123")
	cfg= service.ConfigInitialize("config.toml")
	db,err:= service.DBInitialize(cfg.GetDBFromConfig("employees"))
	if err!=nil{
		return
	}
	service.Db=db
	db.SetMaxOpenConns(100)

	empAddHandler:=http.HandlerFunc(controller.EmployeeAddHandler)
	empRemHandler:=http.HandlerFunc(controller.EmployeeRemoveHandler)
	empUpdHandler:=http.HandlerFunc(controller.EmployeeUpdateHandler)
	empGetAllHandler:=http.HandlerFunc(controller.EmloyeeGetAllHandler)
	empGetHandler:=http.HandlerFunc(controller.EmployeeGetHandler)
	apiInfoHandler:=http.HandlerFunc(APIInfoHandler)

	muxRouter:=mux.NewRouter()
	muxRouter.StrictSlash(true)
	muxRouter.Handle("/api/v1/employees/", RFC7807.ErrRFC7807Middleware(controller.MiddlewareJSONCheck(empAddHandler))).Methods("PUT")
	muxRouter.Handle("/api/v1/employees/", RFC7807.ErrRFC7807Middleware(controller.MiddlewareJSONCheck(empRemHandler))).Methods("DELETE")
	muxRouter.Handle("/api/v1/employees/", RFC7807.ErrRFC7807Middleware(controller.MiddlewareJSONCheck(empUpdHandler))).Methods("POST")
	muxRouter.Handle("/api/v1/employees/", RFC7807.ErrRFC7807Middleware(controller.MiddlewarePackData(empGetAllHandler))).Methods("GET")
	muxRouter.Handle("/api/v1/employees/{employeeId:[0-9]+}", RFC7807.ErrRFC7807Middleware(controller.MiddlewarePackData(empGetHandler))).Methods("GET")
	muxRouter.Handle("/tech/info", apiInfoHandler)
	muxRouter.Handle("/metrics", promhttp.Handler())

	metrics:=controller.InitiateMetrics()
	muxRouter.Use(metrics.MiddlewareMetrics)
	log.Println(http.ListenAndServe("localhost:8000",muxRouter))

}
