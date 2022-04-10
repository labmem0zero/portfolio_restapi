package RFC7807

import (
	"encoding/json"
	gorillactx "github.com/gorilla/context"
	"log"
	"net/http"
	"time"
)


//базовые типы ошибок, используются для удобства оформления
const (
	BaseTypeRead = "read"
	BaseTypeWrite = "write"
	BaseTypeContentTypeIN= "content-type-in"
	BaseTypeContentTypeOUT= "content-type-out"
	BaseTypeMETHOD= "method"
	BaseTypeDBPrepare= "db-prepare"
	BaseTypeDBExec= "db-exec"
	BaseTypeCONTEXT= "context"
	BaseTypeATOI= "atoi"
)

//ErrRFC7807Body структура ошибки по стандарту RFC7807:
//ErrType - адрес к странице ошибки
//Title - название ошибки
//Detail - детальная информация об ошибке
//Instance - путь к узлу, при запросе к которому возникла ошибка
//Params - опциональные параметры, `omitempty`(при значении nil, не включаются в финальный JSON)
type ErrRFC7807Body struct{
	ErrType string				`json:"type"`
	Title string				`json:"title"`
	Detail string				`json:"detail"`
	Instance string				`json:"instance"`
	Params map[string]string	`json:"params,omitempty"`
}

//ErrRFC7807BasicErr (basetype string, detail string, r *http.Request, params map[string]string) ErrRFC7807Body
//basetype - базовый тип ошибки, detail - детали ошибки, r - *http.Request хэндлера,
//params - карта с дополнителньыми параметрами
//возвращает структуру ErrRFC7807Body с полями:
//ErrType - адрес к странице ошикбик с сигнатурой http://ВашХост:порт/probs/тип-ошибки(в зависимости от basetype)
//Title - название ошибки в зависимости от входа basetype, если готового типа нет, то используется "/probs/unhandled"
//Detail - детали ошибки
//Instance - путь к узлу, при запросе к которому возникла ошибка
//Params - опциональные параметры, `omitempty`(при значении nil, не включаются в финальный JSON)
func ErrRFC7807BasicErr(basetype string, detail string, r *http.Request, params map[string]string) ErrRFC7807Body {
	hostname:=r.Host
	port:=r.URL.Port()
	errBody:= ErrRFC7807Body{
		"",
		"",
		detail,
		"http://"+hostname+port+r.URL.Path,
		params,
	}
	if port!=""{
		port=":"+port
	}
	switch basetype {
	case "read":
		errBody.ErrType = "http://" + hostname + port + "/probs/read"
		errBody.Title = "Ошибка при считывании данных из reader."
	case "write":
		errBody.ErrType= "http://" + hostname + port + "/probs/write"
		errBody.Title= "Ошибка при записи данных в writer."
	case "atoi":
		errBody.ErrType= "http://" + hostname + port + "/probs/atoi"
		errBody.Title= "Ошибка при конвертации строки в число."
	case "method":
		errBody.ErrType= "http://" + hostname + port + "/probs/wrong-method"
		errBody.Title= "Ошибка в методе запроса"
	case "content-type-in":
		errBody.ErrType= "http://" + hostname + port + "/probs/content-type-input"
		errBody.Title= "Ошибка, некорректный 'Content-Type'"
	case "content-type-out":
		errBody.ErrType= "http://" + hostname + port + "/probs/content-type-output"
		errBody.Title= "Ошибка, некорректный хидер 'Accept'"
	case "db-query":
		errBody.ErrType= "http://" + hostname + port + "/probs/db-query"
		errBody.Title= "Ошибка при выполнении запроса в БД"
	case "db-exec":
		errBody.ErrType= "http://" + hostname + port + "/probs/db-exec"
		errBody.Title= "Ошибка при записи изменений в БД"
	case "context":
		errBody.ErrType= "http://" + hostname + port + "/probs/context"
		errBody.Title= "Ошибка контекста"
	default:
		errBody.ErrType= "http://" + hostname + port + "/probs/unhandled"
		errBody.Title= "Непредвиденная ошибка: " + basetype
	}
	return errBody
}


//AddKV - добавляет ключ в структуру ошибки. Важно понимать, что при передачи значений основных ключей, старые будут
//переписаны. Если ключ не основной, то он и значение будут добавлены в опциональные параметры.
func (t *ErrRFC7807Body) AddKV(key string,val string){
	switch key {
	case "ErrType":
		t.ErrType=val
	case "Title":
		t.Title=val
	case "Detail":
		t.Detail=val
	case "Instance":
		t.Instance=val
	default:
		params:=t.Params
		if params!=nil{
			t.Params[key]=val
		}else{
			params=make(map[string]string)
			params[key]=val
			t.Params=params
		}
	}
	return
}

//ErrRFC7807Middleware (http.Handler)http.handler - стандартный миддлвейр, заворачивает ваш хэндер
//для удобной обработки ошибок. Обратите внимание: для использования функционала, при обработке ошибки в вашем хэндлере,
//необходимо добавлять ключи и значения в контект реквеста:
//"errortype" - тип ошибки, можно использовать константы из этого пакета, либо прописать кастомный тип
//"errordetail" - детали ошибки, в идеале передавать сюда значение err.Error()
//"errorparams" - map[string]string с описанием дополнительных параметров, можно оставить nil
//"statuscode" - статускод для w.WriteHeader
func ErrRFC7807Middleware(handler http.Handler)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w,r)
		errortype:=gorillactx.Get(r,"errortype")
		if errortype!=nil{
			basetype:=gorillactx.Get(r,"errortype").(string)
			detail:=gorillactx.Get(r,"errordetail").(string)
			params:=gorillactx.Get(r, "errorparams").(map[string]string)
			statuscode:=gorillactx.Get(r, "statuscode").(int)
			RequestRemoveError(r)
			errBody:= ErrRFC7807BasicErr(basetype,detail, r, params)
			w.Header().Set("Content-Type","application/prob+json")
			w.WriteHeader(statuscode)
			jsonAnswer,err:=json.Marshal(errBody)
			if err!=nil{
				log.Printf("%T: Невозможно замаршаллить JSON из:\n---->\n%v\n",time.Now(),errBody)
				return
			}
			_,err=w.Write(jsonAnswer)
			if err!=nil{
				log.Printf("%T: Невозможно отправить данные ошибки:\n---->\n%v\n",time.Now(),jsonAnswer)
			}
			return
		}
		return
	})
}

//RequestSetError - функция для добавление ключей и значений в контекст реквеста, где:
//r - реквест нашего хндлера
//basetype - тип ошибки, можно использовать константы из этого пакета, либо прописать кастомный тип
//detail - детали ошибки, в идеале передавать сюда значение err.Error()
//params - map[string]string с описанием дополнительных параметров, можно оставить nil
//statuscode - статускод для w.WriteHeader
func RequestSetError(r *http.Request, errortype string, detail string, params map[string]string, statuscode int){
	gorillactx.Set(r,"errortype", errortype)
	gorillactx.Set(r,"errordetail", detail)
	gorillactx.Set(r, "errorparams", params)
	gorillactx.Set(r, "statuscode", statuscode)
}

//RequestRemoveError - функция удаления ключей из конекста реквеста,
//автоматически вызывается в теле комплектного миддлвейра
func RequestRemoveError(r *http.Request){
	gorillactx.Delete(r, "errortype")
	gorillactx.Delete(r, "errordetail")
	gorillactx.Delete(r, "errorparams")
	gorillactx.Delete(r, "statuscode")
}

