package main

import (
	"fmt"
	"net/http"
	"time"
)

func GetTypeBitOfNum(num uint32, checkLen uint8, isSet bool) uint32 {
	var ret uint32 = 0

	if isSet {
		for ; checkLen != 0; checkLen-- {
			if (num&1) == 1 {
				ret++
			}
			num >>= 1
		}
	} else {
		for ; checkLen != 0; checkLen-- {
			if (num&1) == 0 {
				ret++
			}
			num >>= 1
		}
	}

	return ret
}

func Get_X_Bit_Num(byteArray [8]uint8, bitStart uint32, bitEnd uint32) uint32 {
	var retNum uint32 = 0
	var tmpNum uint8 = 0

	var startPos int = int(bitStart / 8)
	var endPos int = int(bitEnd / 8)

	var j int = 0
	for j = endPos; j >= startPos; j-- {
		if j == endPos {
			tmpOffset := 8 - ((bitEnd % 8) + 1)
			tmpNum = byteArray[endPos]
			tmpNum <<= tmpOffset
			tmpNum >>= tmpOffset

			retNum = uint32(tmpNum)
		} else if j == startPos {
			tmpOffset := bitStart % 8
			retNum <<= 8 - (tmpOffset)

			tmpNum = byteArray[startPos]
			tmpNum >>= tmpOffset

			retNum |= uint32(tmpNum)

		} else {
			retNum <<= 8
			retNum |= uint32(byteArray[j])
		}
	}

	return retNum
}

func Get_XOR_Proof(authStr string) (uint32, string) {
	ret := [4]uint8{0, 0, 0, 0}

	if len(authStr) != 16 {
		return 0, "authStr len error"
	}
	
	var tmpByte [8]uint8
	
	for i, data := range authStr{
		
		var tmpNum uint8 = 0
		if data >= '0' && data <= '9'{
			tmpNum = uint8(data - '0')
		}else if data >= 'a' && data <= 'f'{
			tmpNum = uint8(data - 'a' + 10)
		}else if data >= 'A' && data <= 'F'{
			tmpNum = uint8(data - 'A' + 10)
		}
		
		if i % 2 == 0{
			tmpByte[i / 2] = tmpNum << 4
		}else{
			tmpByte[i / 2] |= tmpNum
			
			//fmt.Printf("i[%d] tmpByte[%x]\n", i / 2, tmpByte[i / 2])
		}
	}
	
	ret[0] = tmpByte[0] ^ tmpByte[2] ^ tmpByte[3] ^ tmpByte[4]
	ret[1] = ret[0] ^ tmpByte[1] ^ tmpByte[5]
	ret[2] = ret[1] ^ tmpByte[0] ^ tmpByte[6]
	ret[3] = ret[2] ^ tmpByte[1] ^ tmpByte[7]
	
	//fmt.Printf("ret %x %x %x %x\n", ret[3], ret[2], ret[1], ret[0])
	
	var retOriginNum uint32
	var originNum uint32
	var tmpNumAandB uint32
	
	tmpNumAandB = uint32(tmpByte[2])
	tmpNumAandB <<= 8
	tmpNumAandB |= uint32(tmpByte[3])
	
	originNum = uint32(ret[3])
	originNum <<= 8 
	originNum |= uint32(ret[2])
	originNum <<= 8
	originNum |= uint32(ret[1])
	originNum <<= 8
	originNum |= uint32(ret[0])
	
	retOriginNum = originNum
	//fmt.Printf("A+B[%x] origin[%x]\n", tmpNumAandB, originNum)
	
	var retTypeBitOfOriginNum uint32 = 0
	
	if tmpNumAandB & 1 == 1{
		retTypeBitOfOriginNum = GetTypeBitOfNum(originNum, 32, true)
	}else{
		retTypeBitOfOriginNum = GetTypeBitOfNum(originNum, 32, false)
	}
	tmpNumAandB >>= 1
	
	//fmt.Printf("tmpNumAandB[%x] retTypeBitOfOriginNum[%d]\n", tmpNumAandB, retTypeBitOfOriginNum)
	
	if (tmpNumAandB & 0x1F) != retTypeBitOfOriginNum{
		return 0, "error auth 1"
	}
	tmpNumAandB >>= 5
	
	originNum >>= tmpNumAandB & 1
	checkFirstBit := originNum & 1
	tmpNumAandB >>= 1
	
	if checkFirstBit != (tmpNumAandB & 1) {
		return 0, "error auth 2"
	}
	tmpNumAandB >>= 1
	
	originNum >>= tmpNumAandB & 3
	checkFirstBit = originNum & 1
	tmpNumAandB >>= 2
	
	if checkFirstBit != (tmpNumAandB & 1) {
		return 0, "error auth 3"
	}
	tmpNumAandB >>= 1
	
	originNum >>= tmpNumAandB & 0xF
	checkFirstBit = originNum & 1
	tmpNumAandB >>= 4
	
	if checkFirstBit != (tmpNumAandB & 1) {
		return 0, "error auth 4"
	}

	return retOriginNum, ""
}

func http_root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}


func Get_Value_From_Auth(authStr string, ret_vals ...*uint32)string{
	//test value is "94AA6759BA88F288" result is 0x76543210
	checkRetValueLen := len(ret_vals)
	
	if len(authStr) != (checkRetValueLen * 16){
		return "auth len error"
	}

	for i, ptr := range ret_vals{
		tmpAuthStr := authStr[(i * 16): (16 + i * 16)]
		
		result_num, result_reason := Get_XOR_Proof(tmpAuthStr)
		if len(result_reason) != 0{
			return result_reason
		}
		*ptr = result_num
	}

	return "";
}

func This_Timestamp_Is_Invalid(check_timestamp uint32)bool{
	var ret bool = false
	
	timeStamp2 := int64(check_timestamp)
	currentTimestamp := time.Now().Unix()
	
	if currentTimestamp < timeStamp2 {
		DBG_LOG("error time")
		ret = true
	} else if (currentTimestamp - timeStamp2) > 30 {
		DBG_LOG("old time")
		ret = true
	}
	return ret
}

func Get_Auth_Msg(authParam string)(string, bool){

	tmpAuthStr := authParam[0:16]
	result_num, result_reason := Get_XOR_Proof(tmpAuthStr)
	if len(result_reason) != 0{
		DBG_LOG("get timeStamp invalid err:", result_reason)
		return "", false
	}
	timeStamp := result_num

	if This_Timestamp_Is_Invalid(timeStamp){
		DBG_LOG("timeStamp invalid")
		return "", false
	}

	authParam = authParam[16:]
	authParamLen := len(authParam)

	if authParamLen % 16 != 0 || authParamLen == 0{
		DBG_LOG("authParamLen invalid")
		return "", false
	}
	
	array_uint32 := make([]uint32, authParamLen / 16)

	for i, _ := range array_uint32{
		tmp_num, err_msg := Get_XOR_Proof(authParam[(i * 16): (16 + i * 16)])
		if len(err_msg) != 0{
			DBG_LOG("get reg msg invalid err:", err_msg)
			return "", false
		}
		array_uint32[i] = timeStamp ^ tmp_num
	}

	result_msg := RevertUint32ToStr(array_uint32)
	
	return result_msg, true
}

func Reg_Unreg_Node(w http.ResponseWriter, r *http.Request, isReg bool){
	authParam := r.URL.Query().Get("auth")

	DBG_LOG("recv authParam[", authParam, "]")

	node_info_msg, result := Get_Auth_Msg(authParam)

	if result == false{
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}

	var uid_, account_, password_, isHeader string
	
	for i, val := range node_info_msg{
		if val == '|'{
			uid_ = node_info_msg[:i]
			node_info_msg = node_info_msg[i + 1:]
			break
		}
	}

	for i, val := range node_info_msg{
		if val == '|'{
			account_ = node_info_msg[:i]
			node_info_msg = node_info_msg[i + 1:]
			break
		}
	}

	for i, val := range node_info_msg{
		if val == '|'{
			password_ = node_info_msg[:i]
			node_info_msg = node_info_msg[i + 1:]
			break
		}
	}

	for i, val := range node_info_msg{
		if val == '|'{
			isHeader = node_info_msg[:i]
			//node_info_msg = node_info_msg[i + 1:]
			break
		}
	}

	if isReg{
		DBG_LOG("reg uid[", uid_, "] account[", account_, "] password[", password_, "] isHeader[", isHeader, "]")
	}else{
		DBG_LOG("unreg uid[", uid_, "] account[", account_, "] password[", password_, "] isHeader[", isHeader, "]")
	}
	

	if isHeader == "true"{
	
		if isReg{
			ch := make(chan interface{})
			mysql_ch <- mysql_op_data{innerChannel:ch, operator:RegWorker, uid:uid_, account:account_, password:password_}
			retData := <-ch

			if retData.(bool){
				fmt.Fprintf(w, "%s", string(succ_jsonData))
			}else{
				DBG_LOG("database check reg header invalid")
				fmt.Fprintf(w, "%s", string(err_jsonData))
			}
		}else{
			ch := make(chan interface{})
			mysql_ch <- mysql_op_data{innerChannel:ch, operator:UnregWorker, uid:uid_, account:account_, password:password_}
			retData := <-ch

			if retData.(bool){
				fmt.Fprintf(w, "%s", string(succ_jsonData))
			}else{
				DBG_LOG("database check unreg header invalid")
				fmt.Fprintf(w, "%s", string(err_jsonData))
			}
		}
		
	}else if isHeader == "false"{

		if isReg{
			ch := make(chan interface{})
			mysql_ch <- mysql_op_data{innerChannel:ch, operator:RegHeader, uid:uid_, account:account_, password:password_}
			retData := <-ch

			if retData.(bool){
				fmt.Fprintf(w, "%s", string(succ_jsonData))
			}else{
				DBG_LOG("database check reg worker invalid")
				fmt.Fprintf(w, "%s", string(err_jsonData))
			}
		}else{
			ch := make(chan interface{})
			mysql_ch <- mysql_op_data{innerChannel:ch, operator:UnregHeader, uid:uid_, account:account_, password:password_}
			retData := <-ch

			if retData.(bool){
				fmt.Fprintf(w, "%s", string(succ_jsonData))
			}else{
				DBG_LOG("database check unreg header invalid")
				fmt.Fprintf(w, "%s", string(err_jsonData))
			}
		}
		
	}else{
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}
}

func Register_Node_To_This_Header_Central(w http.ResponseWriter, r *http.Request) {
	Reg_Unreg_Node(w, r, true)
}

func UnRegister_Node_To_This_Header_Central(w http.ResponseWriter, r *http.Request) {
	Reg_Unreg_Node(w, r, false)
}

func Start_Cluster(w http.ResponseWriter, r *http.Request) {
	authParam := r.URL.Query().Get("auth")
	
	DBG_LOG("recv start cluster authParam[", authParam, "]")

	cluster_info_msg, result := Get_Auth_Msg(authParam)

	DBG_LOG("recv cluster_info_msg[", cluster_info_msg, "]")

	if result == false{
		fmt.Fprintf(w, "%s", string(err_cluster))
		return
	}

	var api_key, workernum string

	splitStrByChar(cluster_info_msg, '|', &api_key, &workernum)

	if !ApiKeyIsExist(api_key){
		DBG_LOG("error no api key:", api_key)
		fmt.Fprintf(w, "%s", string(err_cluster))
		return
	}

	cluster_data := tcp_server_manager.Start_Cluster(convertIntStrToInt(workernum), api_key);

	DBG_LOG("start cluster:", string(cluster_data))		
	fmt.Fprintf(w, "%s", string(cluster_data))
}

func Stop_Cluster(w http.ResponseWriter, r *http.Request) {
	authParam := r.URL.Query().Get("auth")
	
	cluster_info_msg, result := Get_Auth_Msg(authParam)

	DBG_LOG("recv cluster_info_msg[", cluster_info_msg, "]")

	if result == false{
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}

	var api_key, frp_port string

	splitStrByChar(cluster_info_msg, '|', &api_key, &frp_port)

	if !ApiKeyIsExist(api_key){
		DBG_LOG("error no api key:", api_key)
		fmt.Fprintf(w, "%s", string(err_jsonData))
		return
	}

	ret := node_manager.StopCluster(api_key, frp_port);

	if ret{
		fmt.Fprintf(w, "%s", string(succ_jsonData))
	}else{
		fmt.Fprintf(w, "%s", string(err_jsonData))
	}
}

func Start_One_Worker(w http.ResponseWriter, r *http.Request) {
	authParam := r.URL.Query().Get("auth")
	
	var timeStamp uint32
	result_reason := Get_Value_From_Auth(authParam, &timeStamp)
	if len(result_reason) != 0{
		DBG_LOG("error", result_reason)
		return
	}
	
	if This_Timestamp_Is_Invalid(timeStamp){
		return
	}
}

func Start_One_Header(w http.ResponseWriter, r *http.Request) {
	authParam := r.URL.Query().Get("auth")
	
	var timeStamp uint32
	result_reason := Get_Value_From_Auth(authParam, &timeStamp)
	if len(result_reason) != 0{
		DBG_LOG("error", result_reason)
		return
	}
	
	if This_Timestamp_Is_Invalid(timeStamp){
		return
	}
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

	http.HandleFunc("/", http_root)
	http.HandleFunc("/reg", Register_Node_To_This_Header_Central)
	http.HandleFunc("/unreg", UnRegister_Node_To_This_Header_Central)
	http.HandleFunc("/start/cluster", Start_Cluster)
	http.HandleFunc("/stop/cluster", Stop_Cluster)
	http.HandleFunc("/start/worker", Start_One_Worker)
	http.HandleFunc("/start/header", Start_One_Header)
	
	
	DBG_LOG("Starting server at port ", ip_and_port)
	if err := http.ListenAndServe("0.0.0.0:" + splitStrAfterChar(ip_and_port, ':'), nil); err != nil {
		DBG_LOG("Error starting server:", err)
	}
}
