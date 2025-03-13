package main

import(
	"encoding/json"
)

type RETURN_DATA struct{
	Code int `json:"code"`
}

type ERR_DATA struct{
	Code int `json:"code"`
}

type SUCC_DATA struct{
	Code int `json:"code"`
}

type HEADER_CENTRAL_INFO_DATA struct{
	Code	int		`json:"code"`
	Ip		string	`json:"ip"`
	Port	string	`json:"port"`
}

type CLUSTER_DATA_S struct {
	Code		int			`json:"code"`
	Header_node string		`json:"header_node_id"`
	Worker_node []string	`json:"worker_node_id"`
	Frp_ip		string		`json:"frp_ip"`
	Frp_port	string		`json:"frp_port"`
}

type CLUSTER_DATA_C struct {
	Code		int			`json:"code"`
	Frp_ip		string		`json:"frp_ip"`
	Frp_port	string		`json:"frp_port"`
}

type USER_NODES_DATA struct {
	Code		int			`json:"code"`
	Node_Uid	[]string	`json:"node_uid"`
	Central_Id	[]int		`json:"central_id"`
}

type NODE_STATUS_DATA struct{
	Code			int		`json:"code"`
	Node_Uid		string	`json:"node_uid"`
	Working_Status	int 	`json:"working_status"`
}

var err_jsonData []byte
var err_header_central_info_data []byte
func Init_Default_Json_Type(){
	ret := ERR_DATA{
		Code:FAILED,
	}
	
	jsonData, err := json.Marshal(ret)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
	}

	err_jsonData = jsonData


	ret2 := HEADER_CENTRAL_INFO_DATA{
		Code:FAILED,
		Ip:"",
		Port:"",
	}

	jsonData, err = json.Marshal(ret2)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
	}

	err_header_central_info_data = jsonData
}

func Set_Return_Code_Json(return_code int)[]byte{
	ret := RETURN_DATA{
		Code:return_code,
	}

	jsonData, err := json.Marshal(ret)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
	}

	return jsonData
}

