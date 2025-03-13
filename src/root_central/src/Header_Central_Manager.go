package main

import(
	"net/http"
	"io/ioutil"
	"time"
	"encoding/json"
)

const (
    StartCluster = iota
    StopCluster
    RegNode
    UnregNode
    QueryHeaderStatus
    QueryWorkerStatus
)

var header_central_manager Header_Central_Manager

type Header_Central_Op_Data struct{
	central_id	int
	op			int
	data		interface{}
	ret			chan interface{}
}

type Header_Central_Info struct{
	id			int
	ip			string
	api_port	string
	tcp_port	string
}

type Header_Central_Manager struct{
	sub_header_central map[int]Header_Central_Info

	op_chan chan Header_Central_Op_Data
}

func (hcm *Header_Central_Manager) GET(get_url string) string{
	resp, err := http.Get(get_url)
	if err != nil {
		DBG_LOG("Error:", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		DBG_LOG("Error:", err)
		return ""
	}

	return string(body)
}

func (hcm *Header_Central_Manager) Init_By_Database(){
	ch := make(chan interface{})

	mysql_ch <- mysql_op_data{innerChannel:ch, operator:GetAllHeaderCentralInfo};

	headerCentralsIP := <-ch

	for _, val := range headerCentralsIP.([]Header_Central_Info){
		hcm.sub_header_central[val.id] = val
	}
	
	for _, val := range hcm.sub_header_central{
		DBG_LOG("central[", val.id, "] ip[", val.ip, "] api_port[", val.api_port, "] tcp_port[", val.tcp_port, "]")
	}
}

func (hcm *Header_Central_Manager) Start_Cluster(cluster_info Start_Stop_Cluster_Info) []byte{

	info_data, exsist := hcm.sub_header_central[convertIntStrToInt(cluster_info.header_central_id)]
	
	if !exsist{
		DBG_LOG("Error header_central_id:", cluster_info.header_central_id)
		return err_jsonData
	}

	var auth_str string 

	_, msg_uint32 := StrMsgToUint32Array(cluster_info.apikey + "|" + cluster_info.workernum + "|")

	currentTimestamp := uint32(time.Now().Unix())
	auth_str = Encode(currentTimestamp)
	
	for _, val := range msg_uint32{
		auth_str += Encode(val ^ currentTimestamp)
	}
	
	
	url := "http://" + info_data.ip + ":" + info_data.api_port + "/start/cluster?auth=" + auth_str
	result := hcm.GET(url)

	var cluster_data CLUSTER_DATA_S

	err := json.Unmarshal([]byte(result), &cluster_data)
	if err != nil {
		DBG_LOG("Error decoding JSON:", err)
		return err_jsonData
	}

	DBG_LOG(cluster_data)

	ret := CLUSTER_DATA_C{
		Code: 		cluster_data.Code,
		Frp_ip:		cluster_data.Frp_ip,
		Frp_port:	cluster_data.Frp_port,
	}

	jsonData, err := json.Marshal(ret)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
		return err_jsonData
	}
	
	return jsonData
}

func (hcm *Header_Central_Manager) Stop_Cluster(cluster_info Start_Stop_Cluster_Info) []byte{

	info_data, exsist := hcm.sub_header_central[convertIntStrToInt(cluster_info.header_central_id)]
	
	if !exsist{
		DBG_LOG("Error header_central_id:", cluster_info.header_central_id)
		return err_jsonData
	}

	var auth_str string 

	_, msg_uint32 := StrMsgToUint32Array(cluster_info.apikey + "|" + cluster_info.frp_port + "|")

	currentTimestamp := uint32(time.Now().Unix())
	auth_str = Encode(currentTimestamp)
	
	for _, val := range msg_uint32{
		auth_str += Encode(val ^ currentTimestamp)
	}
	
	url := "http://" + info_data.ip + ":" + info_data.api_port + "/stop/cluster?auth=" + auth_str
	result := hcm.GET(url)
	
	return []byte(result)
}

func (hcm *Header_Central_Manager) Reg_Node(header_central_id int, ni Node_Info) []byte{

	info_data, exsist := hcm.sub_header_central[header_central_id]

	if !exsist{
		DBG_LOG("Error header_central_id:", header_central_id)
		return err_header_central_info_data
	}

	encode_msg := ni.uid + "|" + ni.account + "|" + ni.password + "|" + ni.isHeader + "|"

	_, msg_uint32 := StrMsgToUint32Array(encode_msg)

	currentTimestamp := uint32(time.Now().Unix())

	auth_str := Encode(currentTimestamp)

	for _, val := range msg_uint32{
		auth_str += Encode(currentTimestamp ^ val)
	}

	DBG_LOG("reg str:", auth_str)
	
	url := "http://" + info_data.ip + ":" + info_data.api_port + "/reg?auth=" + auth_str
	result := hcm.GET(url)

	var succ_status SUCC_DATA
	
	err := json.Unmarshal([]byte(result), &succ_status)
	if err != nil {
		DBG_LOG("Error decoding JSON:", err)
		return err_header_central_info_data
	}

	if succ_status.Code == FAILED{
		DBG_LOG("Error request")
		return err_header_central_info_data
	}


	ret := HEADER_CENTRAL_INFO_DATA{
		Code:SUCC,
		Ip:info_data.ip,
		Port:info_data.tcp_port,
	}

	jsonData, err := json.Marshal(ret)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
		return err_header_central_info_data
	}
	
	return jsonData
}

func (hcm *Header_Central_Manager) Unreg_Node(header_central_id int, ni Node_Info) []byte{

	info_data, exsist := hcm.sub_header_central[header_central_id]

	if !exsist{
		DBG_LOG("Error header_central_id:", header_central_id)
		return err_header_central_info_data
	}
	
	encode_msg := ni.uid + "|" + ni.account + "|" + ni.password + "|" + ni.isHeader

	_, msg_uint32 := StrMsgToUint32Array(encode_msg)

	currentTimestamp := uint32(time.Now().Unix())

	auth_str := Encode(currentTimestamp)

	for _, val := range msg_uint32{
		auth_str += Encode(currentTimestamp ^ val)
	}
	
	url := "http://" + info_data.ip + ":" + info_data.api_port + "/unreg?auth=" + auth_str
	result := hcm.GET(url)
	
	return []byte(result)
}

func (hcm *Header_Central_Manager) Query_Node_Status(header_central_id int, ni Node_Info) []byte{

	info_data, exsist := hcm.sub_header_central[header_central_id]

	if !exsist{
		DBG_LOG("Error header_central_id:", header_central_id)
		return err_header_central_info_data
	}
	
	encode_msg := ni.uid + "|" + ni.account + "|" + ni.password + "|" + ni.isHeader

	_, msg_uint32 := StrMsgToUint32Array(encode_msg)

	currentTimestamp := uint32(time.Now().Unix())

	auth_str := Encode(currentTimestamp)

	for _, val := range msg_uint32{
		auth_str += Encode(currentTimestamp ^ val)
	}
	
	url := "http://" + info_data.ip + ":" + info_data.api_port + "/unreg?auth=" + auth_str
	result := hcm.GET(url)
	
	return []byte(result)
}

func (hcm *Header_Central_Manager) Server(){
	for {
		select {
		case hc_op := <-hcm.op_chan:
			switch hc_op.op {
				case StartCluster:
					hc_op.ret <- hcm.Start_Cluster(hc_op.data.(Start_Stop_Cluster_Info))
				case StopCluster:
					hc_op.ret <- hcm.Stop_Cluster(hc_op.data.(Start_Stop_Cluster_Info))
				case RegNode:
					hc_op.ret <- hcm.Reg_Node(hc_op.central_id, hc_op.data.(Node_Info))
				case UnregNode:
					hc_op.ret <- hcm.Unreg_Node(hc_op.central_id, hc_op.data.(Node_Info))
				case QueryHeaderStatus:
					hc_op.ret <- hcm.Query_Node_Status(hc_op.central_id, hc_op.data.(Node_Info))
				case QueryWorkerStatus:
					hc_op.ret <- hcm.Query_Node_Status(hc_op.central_id, hc_op.data.(Node_Info))
				default:
			}
		default:			
		}
	}
}

func (hcm *Header_Central_Manager) Start_Server(){
	hcm.sub_header_central = make(map[int]Header_Central_Info)
	hcm.op_chan = make(chan Header_Central_Op_Data)
	
	hcm.Init_By_Database()
	
	go hcm.Server()
}

func Central_Task_no_ret(central_id_ int, op_ int, data_ interface{}){
	header_central_manager.op_chan <- Header_Central_Op_Data{central_id:central_id_, op:op_, data:data_}
}

func Central_Task(central_id_ int, op_ int, data_ interface{}) interface{}{
	ret_chan := make(chan interface{})
	header_central_manager.op_chan <- Header_Central_Op_Data{central_id:central_id_, op:op_, data:data_, ret:ret_chan}
	retData := <-ret_chan
	return retData
}

