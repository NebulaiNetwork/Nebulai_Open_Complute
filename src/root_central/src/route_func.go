package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
)

func http_root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        // Handle preflight requests
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func API_Start_Cluster(w http.ResponseWriter, r *http.Request) {
	workernum := r.URL.Query().Get("workernum")
	apikey := r.URL.Query().Get("apikey")
	central_id := r.URL.Query().Get("central_id")

	ssci := Start_Stop_Cluster_Info{apikey:apikey, workernum:workernum, header_central_id:central_id}

	retData := Central_Task(convertHEXStrToInt(central_id), StartCluster, ssci)

	DBG_LOG("Start Cluster resullt[", string(retData.([]byte)), "]")
	
	fmt.Fprintf(w, "%s", string(retData.([]byte)))
}

func API_Stop_Cluster(w http.ResponseWriter, r *http.Request) {
	frp_port := r.URL.Query().Get("frp_port")
	apikey := r.URL.Query().Get("apikey")
	central_id := r.URL.Query().Get("central_id")
	
	ssci := Start_Stop_Cluster_Info{apikey:apikey, frp_port:frp_port, header_central_id:central_id}

	retData := Central_Task(convertHEXStrToInt(central_id), StopCluster, ssci)
	
	DBG_LOG("Stop Cluster resullt[", string(retData.([]byte)), "]")
	
	fmt.Fprintf(w, "%s", string(retData.([]byte)))
}

func API_Reg_Node(w http.ResponseWriter, r *http.Request) {
	uid_ := r.URL.Query().Get("uid")
	account_ := r.URL.Query().Get("account")
	password_ := r.URL.Query().Get("password")
	isHeader_ := r.URL.Query().Get("isHeader")
	central_id := r.URL.Query().Get("central_id")

	DBG_LOG("reg uid[", uid_, "], account[", account_, "], password[", password_, "] isheader[", isHeader_, "] central_id[", central_id, "]")

	retData := Central_Task(convertHEXStrToInt(central_id), RegNode, Node_Info{uid:uid_, account:account_, password:password_, isHeader:isHeader_})
	
	fmt.Fprintf(w, "%s", string(retData.([]byte)))
}

func API_Unreg_Node(w http.ResponseWriter, r *http.Request) {
	uid_ := r.URL.Query().Get("uid")
	account_ := r.URL.Query().Get("account")
	password_ := r.URL.Query().Get("password")
	isHeader_ := r.URL.Query().Get("isHeader")
	central_id := r.URL.Query().Get("central_id")

	DBG_LOG("reg uid[", uid_, "], account[", account_, "], password[", password_, "] isheader[", isHeader_, "] central_id[", central_id, "]")

	retData := Central_Task(convertHEXStrToInt(central_id), UnregNode, Node_Info{uid:uid_, account:account_, password:password_, isHeader:isHeader_})
	
	fmt.Fprintf(w, "%s", string(retData.([]byte)))
}

func API_Reg_New_Node(w http.ResponseWriter, r *http.Request) {
	account 		:= r.URL.Query().Get("account")
	password	:= r.URL.Query().Get("password")
	isHeader	:= r.URL.Query().Get("isHeader")
	central_id	:= r.URL.Query().Get("central_id")
	
	DBG_LOG("reg new node account[", account, "] password[", password, "] isHeader[", isHeader, "] central_id[", central_id, "]")

	var ret interface{}
	
	if isHeader == "true"{
		ret = Mysql_Op_Ret(mysql_op_data{operator:RegNewHeader, account:account, password:password, central_id:central_id})
	}else if isHeader == "false"{
		ret = Mysql_Op_Ret(mysql_op_data{operator:RegNewWorker, account:account, password:password, central_id:central_id})
	}else{
		DBG_LOG("isHeader error")
		fmt.Fprintf(w, "%s", string(err_jsonData))
	}
	DBG_LOG("finish reg node")
	fmt.Fprintf(w, "%s", string(Set_Return_Code_Json(ret.(int))))
}

func API_Reg_New_Account(w http.ResponseWriter, r *http.Request) {
	account		:= r.URL.Query().Get("account")
	password	:= r.URL.Query().Get("password")

	DBG_LOG("reg account account[", account, "] password[", password, "]")

	ret := Mysql_Op_Ret(mysql_op_data{operator:RegNewAccount, account:account, password:password})

	fmt.Fprintf(w, "%s", string(Set_Return_Code_Json(ret.(int))))
}

func API_Query_Header(w http.ResponseWriter, r *http.Request) {
	account		:= r.URL.Query().Get("account")
	password	:= r.URL.Query().Get("password")

	DBG_LOG("API_Query_Header account[", account, "] password[", password, "]")

	ret := Mysql_Op_Ret(mysql_op_data{operator:QueryHeader, account:account, password:password})

	jsonData, err := json.Marshal(ret.(USER_NODES_DATA))
	if err != nil {
		DBG_ERR("Error encoding JSON:", err)
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}

	fmt.Fprintf(w, "%s", string(jsonData))
}

func API_Query_Worker(w http.ResponseWriter, r *http.Request) {
	account		:= r.URL.Query().Get("account")
	password	:= r.URL.Query().Get("password")

	DBG_LOG("API_Query_Worker account[", account, "] password[", password, "]")

	ret := Mysql_Op_Ret(mysql_op_data{operator:QueryWorker, account:account, password:password})

	jsonData, err := json.Marshal(ret.(USER_NODES_DATA))
	if err != nil {
		DBG_ERR("Error encoding JSON:", err)
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}

	fmt.Fprintf(w, "%s", string(jsonData))
}

func API_Query_Header_Status(w http.ResponseWriter, r *http.Request) {
	uid 		:= r.URL.Query().Get("uid")
	account		:= r.URL.Query().Get("account")
	password	:= r.URL.Query().Get("password")

	DBG_LOG("account[", account, "] password[", password, "] uid[", uid, "]")

	ret := Mysql_Op_Ret(mysql_op_data{operator:QueryCentralIdByHeaderUid, uid:uid, account:account, password:password})

	header_central_id := ret.(int)

	if header_central_id == -1{
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}
	
}

func API_Query_Worker_Status(w http.ResponseWriter, r *http.Request) {
	account		:= r.URL.Query().Get("account")
	password	:= r.URL.Query().Get("password")

	DBG_LOG("API_Query_Header account[", account, "] password[", password, "]")

	ret := Mysql_Op_Ret(mysql_op_data{operator:QueryHeader, account:account, password:password})

	jsonData, err := json.Marshal(ret.(USER_NODES_DATA))
	if err != nil {
		DBG_ERR("Error encoding JSON:", err)
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}

	fmt.Fprintf(w, "%s", string(jsonData))
}

func Start_Http_Route(){

	ch := make(chan interface{})

	mysql_ch <- mysql_op_data{innerChannel:ch, operator:GetRouteIp}

	route_ip := <-ch

	if len(route_ip.(string)) == 0 {
		DBG_LOG("error ! ip is null")
		return 
	}

	ip_and_port := route_ip.(string)


	r := mux.NewRouter()

    // Apply the CORS middleware to all routes
    r.Use(corsMiddleware)

    // Define routes
	r.HandleFunc("/", http_root).Methods("GET", "POST")
	r.HandleFunc("/start/cluster", API_Start_Cluster).Methods("GET", "POST")
	r.HandleFunc("/stop/cluster", API_Stop_Cluster).Methods("GET", "POST")
	r.HandleFunc("/reg", API_Reg_Node).Methods("GET", "POST")
	r.HandleFunc("/unreg", API_Unreg_Node).Methods("GET", "POST")
	r.HandleFunc("/new/node", API_Reg_New_Node).Methods("GET", "POST")
	r.HandleFunc("/new/account", API_Reg_New_Account).Methods("GET", "POST")
	r.HandleFunc("/query/header", API_Query_Header).Methods("GET", "POST")
	r.HandleFunc("/query/worker", API_Query_Worker).Methods("GET", "POST")
	r.HandleFunc("/query/header/status", API_Query_Header_Status).Methods("GET", "POST")
	r.HandleFunc("/query/worker/status", API_Query_Worker_Status).Methods("GET", "POST")


	/*
	http.HandleFunc("/", http_root)
	http.HandleFunc("/start/cluster", API_Start_Cluster)
	http.HandleFunc("/stop/cluster", API_Stop_Cluster)
	http.HandleFunc("/reg", API_Reg_Node)
	http.HandleFunc("/unreg", API_Unreg_Node)
	http.HandleFunc("/new/node", API_Reg_New_Node)
	http.HandleFunc("/new/account", API_Reg_New_Account)
	*/
	
	_, port := splitStrAfterChar(ip_and_port, ':')
	DBG_LOG("Starting server at port ", ip_and_port)
	if err := http.ListenAndServe("0.0.0.0:" + port, r); err != nil {
		DBG_LOG("Error starting server:", err)
	}
}
