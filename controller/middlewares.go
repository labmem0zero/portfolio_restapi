package controller

import (
	gorillactx "github.com/gorilla/context"
	"log"
	"net/http"
	"wbEmployeeApi/RFC7807"
)

func MiddlewareJSONCheck(handler http.Handler)http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request){
		if r.Header.Get("Content-Type")!="application/json"{
			RFC7807.RequestSetError(r, "content-type-in", "Хидер 'Content-Type' должен быть 'application/json'", nil, http.StatusUnsupportedMediaType)
			return
		}
		handler.ServeHTTP(w,r)
	})
}

func MiddlewarePackData(handler http.Handler)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w,r)
		if errortype:=gorillactx.Get(r, "errortype");errortype!=nil{
			return
		}
		customData:=gorillactx.Get(r, "customdata")
		gorillactx.Delete(r,"customdata")
		if customData==nil{
			return
		}
		var responseBody []byte
		var formatErr error
		mediatype:= GetAcceptableMediatype(r.Header.Get("Accept"))
		switch  mediatype{
		case "application/json":
			responseBody, formatErr = FormJSON(customData)
		case "application/xml":
			responseBody, formatErr = FormXML(customData)
		default:
			RFC7807.RequestSetError(r, "content-type-out", "Хидер 'Accept' должен быть 'application/json' либо 'application/xml'", nil, http.StatusNotAcceptable)
			return
		}
		if formatErr!=nil{
			log.Println("Ошибка при конвертации данных в json/xml:",formatErr)
			return
		}
		w.Header().Set("Content-Type",mediatype)
		w.WriteHeader(http.StatusOK)
		_,err:=w.Write(responseBody)
		if err!=nil{
			log.Println("Ошибка при отправке ответа:",err)
		}
		return
	})
}


