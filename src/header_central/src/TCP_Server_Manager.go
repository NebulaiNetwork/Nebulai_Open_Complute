package main

import (
	"net"
	"sync"
	"time"
	"encoding/json"
)

var tcp_server_manager TCP_Server_Manager

type Header_Central_Info struct{
	id			int
	ip			string
	api_port	string
	tcp_port	string
}


type Node_info struct{
	uid			string
	account		string
	password	string
	isHeader	string
}

func (ni *Node_info) Parser_Node_Info(node_msg string) bool{

	var info_array [4]string
	k := 0
	last_i := 0
	for i, char_ := range node_msg {
		if char_ == '|' {
			info_array[k] = node_msg[last_i:i]
			last_i = i + 1
			k++
			if k == 3{
				info_array[k] = node_msg[last_i:]
				break
			}
		}
	}
	
	if k != 3{
		return false
	}else{
		uid			:= info_array[0]
		account 	:= info_array[1]
		password 	:= info_array[2]
		isHeader 	:= info_array[3]

		if len(uid) > 8 || len(account) > 128 || len(password) > 64 || (len(isHeader) != 4 && len(isHeader) != 5){
			return false
		}

		ni.uid		= uid
		ni.account	= account
		ni.password	= password
		ni.isHeader	= isHeader
	}

	return true
}

type Net_Conn_Info struct{
	conn			net.Conn
	now_check_time	uint32
	max_check_time	uint32
	node_info		Node_info

	failed_heart	uint32
}

type TCP_Server_Manager struct{
	worker_nodes	map[string]Net_Conn_Info
	header_nodes	map[string]Net_Conn_Info
	
	this_header_central_info	Header_Central_Info

	worker_setMutex sync.Mutex
	header_setMutex sync.Mutex
	header_central_setMutex sync.Mutex

	clientConnectChannel chan net.Conn
}

func (tcp_sm *TCP_Server_Manager) InitTCPServer() {

	ch := make(chan interface{})

	mysql_ch <- mysql_op_data{innerChannel:ch, operator:GetHeaderCentralTCP};

	ip_and_port := <-ch

	tcp_sm.this_header_central_info = ip_and_port.(Header_Central_Info)

	listener, err := net.Listen("tcp", "0.0.0.0:" + tcp_sm.this_header_central_info.tcp_port)
	if err != nil {
		DBG_LOG("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	DBG_LOG("Server listening on port 6400...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			DBG_LOG("listener new connet err:", err)

			//do not
			continue
		}
		DBG_LOG("recv new connet:", conn)
		tcp_sm.clientConnectChannel <- conn
	}
}

func (tcp_sm *TCP_Server_Manager) DispatchThread() {
	for {
		select {
		case conn := <-tcp_sm.clientConnectChannel:
			tcp_sm.NewNode(conn)
		//case <-time.After(100 * time.Millisecond):
			// Prevent busy waiting
		default:
			
		}
	}
}

func (tcp_sm *TCP_Server_Manager) Init_Server(){
	tcp_sm.worker_nodes = make(map[string]Net_Conn_Info)
	tcp_sm.header_nodes = make(map[string]Net_Conn_Info)
	tcp_sm.clientConnectChannel = make(chan net.Conn)

	timer_manager.Reg_Timer(10, func(){
	
		for i, net_conn_info := range tcp_sm.worker_nodes{
			go tcp_sm.Hert_Beat(i, net_conn_info, false, "00|heartBeat")
		}

		for i, net_conn_info := range tcp_sm.header_nodes{
			go tcp_sm.Hert_Beat(i, net_conn_info, true, "00|heartBeat")
		}
	})

	go tcp_sm.InitTCPServer()
	go tcp_sm.DispatchThread()
}

func (tcp_sm *TCP_Server_Manager) NewNode(conn_ net.Conn) {

	buffer := make([]byte, 1024)

	DBG_LOG("wait coon client send msg")
	
	n, err := conn_.Read(buffer)
	if err != nil {
		conn_.Close()
		DBG_LOG("err read reason:", err)
		return
	}
	DBG_LOG("Received uid connet: ", string(buffer[:n]))

	var tmp_node Node_info

	if tmp_node.Parser_Node_Info(string(buffer[:n])) == false{
		conn_.Close()
		DBG_LOG("msg type error", err)
		return
	}
	
	if tmp_node.isHeader == "true"{

		ch := make(chan interface{})
		mysql_ch <- mysql_op_data{innerChannel:ch, operator:CheckHeaderUidExist, uid:tmp_node.uid, account:tmp_node.account, password:tmp_node.password};
		result := <-ch

		if result.(bool) == false{
			conn_.Close()
			DBG_LOG("uid[", tmp_node.uid, "] no in header database")
			return
		}
		_, exist :=	tcp_sm.header_nodes[tmp_node.uid]

		if exist{
			conn_.Close()
			DBG_LOG("uid[", tmp_node.uid, "] already exist")
			return
		}
	
		tcp_sm.header_setMutex.Lock()
		tcp_sm.header_nodes[tmp_node.uid] = Net_Conn_Info{conn:conn_, now_check_time:0, max_check_time:6, node_info:tmp_node, failed_heart:0}
		tcp_sm.header_setMutex.Unlock()

		node_manager.NewNode(tmp_node.uid, true)
	} else if tmp_node.isHeader == "false"{

		ch := make(chan interface{})
		mysql_ch <- mysql_op_data{innerChannel:ch, operator:CheckWorkerUidExist, uid:tmp_node.uid, account:tmp_node.account, password:tmp_node.password};
		result := <-ch

		if result.(bool) == false{
			conn_.Close()
			DBG_LOG("uid[", tmp_node.uid, "] no in worker database")
			return
		}

		_, exist :=	tcp_sm.worker_nodes[tmp_node.uid]

		if exist{
			conn_.Close()
			DBG_LOG("uid[", tmp_node.uid, "] already exist")
			return
		}
	
		tcp_sm.worker_setMutex.Lock()
		tcp_sm.worker_nodes[tmp_node.uid] = Net_Conn_Info{conn:conn_, now_check_time:0, max_check_time:6, node_info:tmp_node, failed_heart:0}
		tcp_sm.worker_setMutex.Unlock()

		node_manager.NewNode(tmp_node.uid, false)
	} else{
		DBG_LOG("error isHeader")
		conn_.Close()
	}
}

func (tcp_sm *TCP_Server_Manager) Start_Cluster(workerNum int, apikey string) []byte{
	/*
	type CLUSTER_DATA_S struct {
		Code		int 		`json:"code"`
		Header_node string		`json:"header_node_id"`
		Worker_node []string	`json:"worker_node_id"`
		Frp_ip		string		`json:"frp_ip"`
		Frp_port	string		`json:"frp_port"`
	}
	*/

	var err error
	var jsonData []byte
	var cluster_info CLUSTER_DATA_S
	var ret_worker_uid []string

	var ret bool

	var redis = "123123"

	clustNodes, frpClientInfo, result := node_manager.GetClusterNode(int(workerNum), apikey)
	header_message := "01|" + redis + "|" + frpClientInfo.port + "|" + frpClientInfo.obj_port + "|" + frpClientInfo.dashboard_port
	worker_message := "02|" + redis + "|" + frpClientInfo.port
	
	buffer := make([]byte, 1024)

	if !result{
		DBG_LOG("select cluster failed")
		goto ERR_RELEASE
	}

	DBG_LOG("start cluster")

	for true{

		_, exsist := tcp_sm.header_nodes[clustNodes[0].uid]

		if !exsist{
			DBG_LOG("header id[" + clustNodes[0].uid + "] no login")

			anotherNode, op_result := node_manager.GetAnotherNode(clustNodes[0], true, frpClientInfo.port)

			if op_result == false{
				DBG_LOG("retry get header failed")
				goto ERR_RELEASE
			}

			clustNodes[0] = anotherNode
			
			continue
		}
	
		 _, err := tcp_sm.header_nodes[clustNodes[0].uid].conn.Write([]byte(header_message))
		if err != nil {
			DBG_LOG("Error writing to header:", err)
			
			anotherNode, op_result := node_manager.GetAnotherNode(clustNodes[0], true, frpClientInfo.port)

			if op_result == false{
				DBG_LOG("retry get header failed")
				goto ERR_RELEASE
			}

			clustNodes[0] = anotherNode

			continue
		}

		tcp_sm.header_nodes[clustNodes[0].uid].conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		n, err := tcp_sm.header_nodes[clustNodes[0].uid].conn.Read(buffer)
		if err != nil {
			DBG_LOG("read header msg failed:", err)

			anotherNode, op_result := node_manager.GetAnotherNode(clustNodes[0], true, frpClientInfo.port)

			if op_result == false{
				DBG_LOG("retry get header failed")
				goto ERR_RELEASE
			}

			clustNodes[0] = anotherNode

			continue
		}
		
		ret = tcp_sm.Parser_Return_Code(string(buffer[:n]))
		if ret == false{
			DBG_LOG("init header failed")

			anotherNode, op_result := node_manager.GetAnotherNode(clustNodes[0], true, frpClientInfo.port)

			if op_result == false{
				DBG_LOG("retry get header failed")
				goto ERR_RELEASE
			}

			clustNodes[0] = anotherNode
			continue
		}

		break
	}

	for i, val := range clustNodes{

		if i == 0{
			continue
		}

		for true{
			_, exsist := tcp_sm.worker_nodes[val.uid]
		
			if !exsist{
				DBG_LOG("worker id[" + val.uid + "] no login")
	
				anotherNode, op_result := node_manager.GetAnotherNode(val, false, frpClientInfo.port)
	
				if op_result == false{
					DBG_LOG("retry get worker failed")
					goto ERR_RELEASE
				}
	
				clustNodes[i] = anotherNode
				
				continue
			}
		
			 _, err := tcp_sm.worker_nodes[val.uid].conn.Write([]byte(worker_message))
			if err != nil {
				DBG_LOG("Error writing to server:", err)
				
				anotherNode, op_result := node_manager.GetAnotherNode(val, false, frpClientInfo.port)
	
				if op_result == false{
					DBG_LOG("retry get worker failed")
					goto ERR_RELEASE
				}
	
				clustNodes[i] = anotherNode
				
				continue
			}
	
			tcp_sm.worker_nodes[val.uid].conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	
			n, err := tcp_sm.worker_nodes[val.uid].conn.Read(buffer)
			if err != nil {
				DBG_LOG("read worker msg failed:", err)

				anotherNode, op_result := node_manager.GetAnotherNode(val, false, frpClientInfo.port)
	
				if op_result == false{
					DBG_LOG("retry get worker failed")
					goto ERR_RELEASE
				}
	
				clustNodes[i] = anotherNode
				
				continue
			}
			
			ret = tcp_sm.Parser_Return_Code(string(buffer[:n]))
			if ret == false{
				DBG_LOG("init worker failed")

				anotherNode, op_result := node_manager.GetAnotherNode(val, false, frpClientInfo.port)
	
				if op_result == false{
					DBG_LOG("retry get worker failed")
					goto ERR_RELEASE
				}
	
				clustNodes[i] = anotherNode
				
				continue
			}	
			break
		}
	}
	
	for i, val := range clustNodes{
		if i == 0{
			continue 
		}
		ret_worker_uid = append(ret_worker_uid, val.uid)
	}

	cluster_info = CLUSTER_DATA_S {
		Code:SUCC,
		Header_node:clustNodes[0].uid,
		Worker_node:ret_worker_uid,
		Frp_ip:frpClientInfo.ip,
		Frp_port:frpClientInfo.dashboard_port,
	}

	
	jsonData, err = json.Marshal(cluster_info)
	if err != nil {
		DBG_LOG("Error encoding JSON:", err)
		return err_cluster
	}
	

	return jsonData

ERR_RELEASE:

	node_manager.StopCluster(apikey, frpClientInfo.port)

	return err_cluster
}

func (tcp_sm *TCP_Server_Manager) Hert_Beat(index string, net_conn_info Net_Conn_Info, isHeader bool, heart_beat_msg string) {
	buffer := make([]byte, 16)

	after_delete := false
	net_conn_info.now_check_time++

	if net_conn_info.now_check_time > net_conn_info.max_check_time{
		net_conn_info.now_check_time = 0

		_, err := net_conn_info.conn.Write([]byte(heart_beat_msg))
		if err != nil {
			after_delete = true
			DBG_LOG("Error writing to Worker:", err)
		}

		if after_delete != true{

			net_conn_info.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		
			n, err2 := net_conn_info.conn.Read(buffer)
			if err2 != nil {
				after_delete = true
				DBG_LOG("err read reason:", err)
			}

			if after_delete != true{
				if string(buffer[:n]) != "heartBeat"{
					after_delete = true
					DBG_LOG("worker node return err:", err)
				}else{
					//finish heartbeat
					net_conn_info.failed_heart = 0
				}
			}					
		}
	}

	if isHeader{
		if after_delete{
			DBG_LOG("uid[", index, "] after delete[", after_delete, "] net_conn_info.failed_heart[", net_conn_info.failed_heart, "]")
			if net_conn_info.failed_heart < 2{
				net_conn_info.failed_heart ++
				tcp_sm.header_nodes[index] = net_conn_info
			}else{
				net_conn_info.conn.Close()
			
				tcp_sm.header_setMutex.Lock()
				delete(tcp_sm.header_nodes, index)
				tcp_sm.header_setMutex.Unlock()

				node_manager.DeleteNode(index, true)
			}
			
		}else{
			tcp_sm.header_nodes[index] = net_conn_info
		}
	}else{
		if after_delete{
			DBG_LOG("uid[", index, "] after delete[", after_delete, "] net_conn_info.failed_heart[", net_conn_info.failed_heart, "]")
			if net_conn_info.failed_heart < 2{
				net_conn_info.failed_heart ++
				tcp_sm.worker_nodes[index] = net_conn_info
			}else{
				net_conn_info.conn.Close()
			
				tcp_sm.worker_setMutex.Lock()
				delete(tcp_sm.worker_nodes, index)
				tcp_sm.worker_setMutex.Unlock()

				node_manager.DeleteNode(index, false)
			}
		}else{
			tcp_sm.worker_nodes[index] = net_conn_info
		}
	}
}

func (tcp_sm *TCP_Server_Manager) Trigger_Hert_Beat(index string, net_conn_info Net_Conn_Info, isHeader bool, heart_beat_msg string) bool{
	buffer := make([]byte, 16)

	after_delete := false

	net_conn_info.now_check_time = 0

	_, err := net_conn_info.conn.Write([]byte(heart_beat_msg))
	if err != nil {
		net_conn_info.conn.Close()
		after_delete = true
		DBG_LOG("Error writing to Worker:", err)
	}

	if after_delete != true{

		net_conn_info.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
		n, err2 := net_conn_info.conn.Read(buffer)
		if err2 != nil {
			net_conn_info.conn.Close()
			after_delete = true
			DBG_LOG("err read reason:", err)
		}

		if after_delete != true{
			if string(buffer[:n]) != "heartBeat"{
				net_conn_info.conn.Close()
				after_delete = true
				DBG_LOG("worker node return err:", err)
			}
		}					
	}


	if isHeader{
		if after_delete{
			tcp_sm.header_setMutex.Lock()
			delete(tcp_sm.header_nodes, index)
			tcp_sm.header_setMutex.Unlock()

			node_manager.DeleteNode(index, true)
		}else{
			tcp_sm.header_nodes[index] = net_conn_info
		}
	}else{
		if after_delete{
			tcp_sm.worker_setMutex.Lock()
			delete(tcp_sm.worker_nodes, index)
			tcp_sm.worker_setMutex.Unlock()

			node_manager.DeleteNode(index, false)
		}else{
			tcp_sm.worker_nodes[index] = net_conn_info
		}
	}

	return !after_delete
}

// pub func:

func (tcp_sm *TCP_Server_Manager) Parser_Return_Code(msg string) bool{
	var uid, account, password, isHeader, code string

	for i, val := range msg {
		if val == '|' {
			uid = msg[:i]
			msg = msg[(i + 1):]
			break
		}
	}

	for i, val := range msg {
		if val == '|' {
			account = msg[:i]
			msg = msg[(i + 1):]
			break
		}
	}

	for i, val := range msg {
		if val == '|' {
			password = msg[:i]
			msg = msg[(i + 1):]
			break
		}
	}

	for i, val := range msg {
		if val == '|' {
			isHeader = msg[:i]
			msg = msg[(i + 1):]
			break
		}
	}

	code = msg

	var nci Net_Conn_Info
	var exsist bool

	if isHeader == "true"{
		nci, exsist = tcp_sm.header_nodes[uid]
		if !exsist{
			DBG_LOG("uid[" + uid + "] account[" + account + "] password[" + password + "] isHeader[" + isHeader + "] code[" + code + "] not in our data")
			return false
		}
	}else if isHeader == "false"{
		nci, exsist = tcp_sm.worker_nodes[uid]
		if !exsist{
			DBG_LOG("uid[" + uid + "] account[" + account + "] password[" + password + "] isHeader[" + isHeader + "] code[" + code + "] not in our data")
			return false
		}
	}else{
		DBG_LOG("uid[" + uid + "] account[" + account + "] password[" + password + "] isHeader[" + isHeader + "] code[" + code + "] not in our data")
		return false
	}

	ni := nci.node_info

	if ni.account == account && ni.password == password && ni.isHeader == isHeader{
		switch code{
			case "1":{
				return true
			}
			case "4001":{
				DBG_LOG("header init failed")
				return false
			}
			case "4002":{
				DBG_LOG("worker init failed")
				return false
			}
		}
	}else{
		DBG_LOG("uid[" + uid + "] account[" + account + "] password[" + password + "] isHeader[" + isHeader + "] code[" + code + "] not in our data")
		return false
	} 

	return false
}

func (tcp_sm *TCP_Server_Manager) Send_Msg_To_Node(uid string, msg string, isHeader bool) bool{

	var exsist bool

	if isHeader{
		_, exsist = tcp_sm.header_nodes[uid]
	}else{
		_, exsist = tcp_sm.worker_nodes[uid]
	}

	if !exsist{
		DBG_LOG("uid[" + uid + "] isHeader[", isHeader, "]no login")

		return false
	}

	var err error

	if isHeader{
		_, err = tcp_sm.header_nodes[uid].conn.Write([]byte(msg))
	}else{
		_, err = tcp_sm.worker_nodes[uid].conn.Write([]byte(msg))
	}

	if err != nil {
		DBG_LOG("Error writing to node:", err)
		return false
	}
	return true
}


