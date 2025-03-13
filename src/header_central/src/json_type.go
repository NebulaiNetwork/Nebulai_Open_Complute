package main

import(
	"encoding/json"
)


type ERR_DATA struct{
	Code int `json:"code"`
}

type SUCC_DATA struct{
	Code int `json:"code"`
}

type CLUSTER_DATA_S struct {
	Code		int			`json:"code"`
	Header_node string		`json:"header_node_id"`
	Worker_node []string	`json:"worker_node_id"`
	Frp_ip		string		`json:"frp_ip"`
	Frp_port	string		`json:"frp_port"`
}

var err_jsonData	[]byte
var succ_jsonData	[]byte
var err_cluster		[]byte
func Init_Default_Json_Type(){
	err_ret := ERR_DATA{
		Code:FAILED,
	}
	
	jsonData, err := json.Marshal(err_ret)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
	}

	err_jsonData = jsonData

	succ_ret := SUCC_DATA{
		Code:SUCC,
	}
	
	jsonData2, err := json.Marshal(succ_ret)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
	}

	succ_jsonData = jsonData2
	

	err_ret2 := CLUSTER_DATA_S{
		Code:FAILED,
	}
	
	jsonData3, err := json.Marshal(err_ret2)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
	}

	err_cluster = jsonData3
}

