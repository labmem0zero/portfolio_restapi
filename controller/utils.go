package controller

import (
	"encoding/json"
	"encoding/xml"
	"strings"
)

func GetAcceptableMediatype(mediatypes string)string{
	mType:=""
	if strings.Contains(mediatypes,"application/xml"){
		mType="application/xml"
	}
	if strings.Contains(mediatypes,"application/json"){
		mType="application/json"
	}
	return mType
}

func FormJSON(obj interface{}) ([]byte,error){
	jsonBody,err:=json.Marshal(obj)
	if err!=nil{
		return nil, err
	}
	return jsonBody,nil
}

func FormXML(obj interface{}) ([]byte,error){
	xmlBody,err:=xml.Marshal(obj)
	if err!=nil{
		return nil, err
	}
	return xmlBody,nil
}
