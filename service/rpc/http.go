package rpc

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func dealHome(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("hacash rpc."))
}

func dealQuery(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	params := make(map[string]string, 0)
	for k, v := range request.Form {
		//fmt.Println("key:", k)
		//fmt.Println("val:", strings.Join(v, ""))
		params[k] = strings.Join(v, "")
	}
	if _, ok := params["action"]; !ok {
		response.Write([]byte("must action"))
		return
	}

	// call controller
	routeQueryRequest(params["action"], params, response, request)

}

func dealOperate(response http.ResponseWriter, request *http.Request) {

	//request.Body.Read()

}

func RunHttpRpcService() {

	initRoutes()

	http.HandleFunc("/", dealHome)           //设置访问的路由
	http.HandleFunc("/query", dealQuery)     //设置访问的路由
	http.HandleFunc("/operate", dealOperate) //设置访问的路由

	port := "3334"

	err := http.ListenAndServe(":"+port, nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	} else {
		fmt.Println("RunHttpRpcService on " + port)
	}
}
