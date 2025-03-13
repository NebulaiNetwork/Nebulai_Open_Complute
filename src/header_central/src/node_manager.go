package main

import (
	"sync"
)

var node_manager NodeManager

type Node struct {
	uid		string
	ip		string
	port	string
}

type FrpClient struct{
	ip					string
	port				string
	obj_port			string
	dashboard_port		string
	under_use_api_key	string
}

type Node_Info struct{
	node			Node
	isWorking		bool
	workingAtFrp	string
}

type NodeManager struct {
	headers			[]Node_Info
	workers			[]Node_Info

	frp_clients		[]FrpClient
	frpPortIsWork	[]bool
	
	workerSetMutex	sync.Mutex
	headerSetMutex	sync.Mutex
	frpSetMutex		sync.Mutex
}

func (nm *NodeManager) NewNode(uid_ string, isHeader bool) {
	if isHeader == true {
		nm.headerSetMutex.Lock()

		nm.headers = append(nm.headers, Node_Info{node:Node{uid: uid_}, isWorking:false, workingAtFrp:""})

		nm.headerSetMutex.Unlock()
	} else {
		nm.workerSetMutex.Lock()

		nm.workers = append(nm.workers, Node_Info{node:Node{uid: uid_}, isWorking:false, workingAtFrp:""})

		nm.workerSetMutex.Unlock()
	}
}

func (nm *NodeManager) DeleteNode(uid_ string, isHeader bool) {
	DBG_LOG("delete uid[", uid_, "]")
	
	if isHeader == true {
		nm.headerSetMutex.Lock()

		for i, val := range nm.headers{

			if val.node.uid == uid_{
				DBG_LOG("delete header uid[", uid_, "] node")
				nm.headers = append(nm.headers[:i], nm.headers[i + 1:]...)
				break
			}
		}
		
		nm.headerSetMutex.Unlock()
	} else {
		nm.workerSetMutex.Lock()

		for i, val := range nm.workers{

			if val.node.uid == uid_{
				DBG_LOG("delete worker uid[", uid_, "] node")
				nm.workers = append(nm.workers[:i], nm.workers[i + 1:]...)
				break
			}
		}

		nm.workerSetMutex.Unlock()
	}
}


func (nm *NodeManager) SetNodeWorkingStatus(isHeader bool, index int, frp_port string) {

	if isHeader{
		nm.headers[index].isWorking = true
		nm.headers[index].workingAtFrp = frp_port
	}else{
		nm.workers[index].isWorking = true
		nm.workers[index].workingAtFrp = frp_port
	}
}

func (nm *NodeManager) ClrNodeWorkingStatus(isHeader bool, index int) {

	if isHeader{
		nm.headers[index].isWorking = false
		nm.headers[index].workingAtFrp = ""
	}else{
		nm.workers[index].isWorking = false
		nm.workers[index].workingAtFrp = ""
	}
}


func (nm *NodeManager) GetUnUesHeaderNode(frpPort string) (Node, bool) {

	nm.headerSetMutex.Lock()
	defer nm.headerSetMutex.Unlock()

	for i, val := range nm.headers {
		if val.isWorking == false {
			nm.SetNodeWorkingStatus(true, i, frpPort)
			return nm.headers[i].node, true
		}
	}	
	return Node{}, false
}

func (nm *NodeManager) GetUnUesWorkerNode(frpPort string) (Node, bool) {

	nm.workerSetMutex.Lock()
	defer nm.workerSetMutex.Unlock()

	for i, val := range nm.workers {
		if val.isWorking == false {
			nm.SetNodeWorkingStatus(false, i, frpPort)
			return nm.workers[i].node, true
		}
	}
	return Node{}, false
}

func (nm *NodeManager) HeaderNodeFinishWork(uid_ string) bool {

	nm.headerSetMutex.Lock()
	defer nm.headerSetMutex.Unlock()

	for i, val := range nm.headers {
		if val.node.uid == uid_ {
			nm.ClrNodeWorkingStatus(true, i)
			return true
		}
	}
	return false
}

func (nm *NodeManager) GetClusterNode(workerNum int, apikey string) ([]Node, FrpClient, bool) {
	var ret_node []Node
	var ret_frp FrpClient

	if workerNum > len(nm.workers) {
		DBG_ERR("worker num")
		return ret_node, ret_frp, false
	}

	//allocate frp
	{
		nm.frpSetMutex.Lock()
		defer nm.frpSetMutex.Unlock()

		unfinish := true

		for i, val := range nm.frpPortIsWork {
			if val == false {
				nm.frpPortIsWork[i] = true
				ret_frp = nm.frp_clients[i]
				nm.frp_clients[i].under_use_api_key = apikey

				unfinish = false
				break
			}
		}

		if unfinish {
			DBG_ERR("Insufficient frp resources")
			return ret_node, ret_frp, false
		}
	}

	//allocate header 
	headerNode, have := nm.GetUnUesHeaderNode(ret_frp.port)

	if have != true {
		DBG_ERR("Insufficient number of header")
		return ret_node, ret_frp, false
	}

	ret_node = append(ret_node, headerNode)

	//allocate worker
	{
		nm.workerSetMutex.Lock()
		defer nm.workerSetMutex.Unlock()

		can_use_num := 0
		var use_worker_pos []int

		for i, val := range nm.workers {
			if val.isWorking == false {
				nm.SetNodeWorkingStatus(false, i, ret_frp.port)
				ret_node = append(ret_node, nm.workers[i].node)

				can_use_num++
				use_worker_pos = append(use_worker_pos, i)

				if(can_use_num == workerNum) {
					return ret_node, ret_frp, true
				}
			}
		}

		DBG_LOG("Insufficient number of workers")
		for _, index := range use_worker_pos{
			nm.ClrNodeWorkingStatus(false, index)
		}

		nm.HeaderNodeFinishWork(headerNode.uid)
	}

	DBG_ERR("can`t find")
	return ret_node, ret_frp, false
}

func (nm *NodeManager) ReleaseWorkerNode(nodes []Node) {

	for i, val1 := range nm.workers {
		for _, val2 := range nodes{
			if val1.node.uid == val2.uid{
				nm.workerSetMutex.Lock()
				defer nm.workerSetMutex.Unlock()
				
				nm.ClrNodeWorkingStatus(false, i)
				break
			}
		}
	}
}

func (nm *NodeManager) ReleaseHeaderNode(node Node) bool {

	for i, val := range nm.headers {
		if val.node.uid == node.uid{
			
			nm.headerSetMutex.Lock()
			defer nm.headerSetMutex.Unlock()

			nm.ClrNodeWorkingStatus(true, i)
			return true
		}
	}
	return false
}


func (nm *NodeManager) GetAnotherNode(node_ Node, isHeader bool, frp_port string) (Node, bool){
	var ret_node Node

	find_new := false
	find_old := false
	
	if isHeader{

		nm.headerSetMutex.Lock()
		defer nm.headerSetMutex.Unlock()
	
		for i, val := range nm.headers {
			if !find_new && val.isWorking == false {
			
				nm.SetNodeWorkingStatus(true, i, frp_port)
				ret_node = nm.headers[i].node

				find_new = true
			}else if !find_old && val.node.uid == node_.uid{

				nm.ClrNodeWorkingStatus(true, i)

				find_old = true
			}

			if find_new && find_old {
				break
			}
		}
	}else{

		nm.workerSetMutex.Lock()
		defer nm.workerSetMutex.Unlock()
	
		for i, val := range nm.workers {
			if !find_new && val.isWorking == false {
			
				nm.SetNodeWorkingStatus(false, i, frp_port)
				ret_node = nm.workers[i].node

				find_new = true
			}else if !find_old && val.node.uid == node_.uid{

				nm.ClrNodeWorkingStatus(false, i)

				find_old = true
			}

			if find_new && find_old {
				break
			}
		}
	}

	return ret_node, (find_new && find_old)
}

func (nm *NodeManager) StopCluster(apikey string, frp_port_str string) bool{

	close_message := "03|_"
	
	{
		nm.frpSetMutex.Lock()
		defer nm.frpSetMutex.Unlock()

		unrelease_frp := true

		for i, val := range nm.frp_clients{
			DBG_LOG("Compare frp_port_str[", frp_port_str, "] val.port[", val.port, "]")
			if val.port == frp_port_str && val.under_use_api_key == apikey{
				nm.frpPortIsWork[i] = false
				nm.frp_clients[i].under_use_api_key = ""
				DBG_LOG("release FRP port[", frp_port_str, "]")
				unrelease_frp = false
	
				break
			}
		}

		if unrelease_frp{
			return false
		}
	}
	{
		for i, val := range nm.headers {
			DBG_LOG("header workingFrp[", val.workingAtFrp, "] val.isWorking[", val.isWorking, "]")
			if val.workingAtFrp == frp_port_str && val.isWorking == true {

				ret := tcp_server_manager.Send_Msg_To_Node(val.node.uid, close_message, true)
				 
				if !ret{
					DBG_LOG("Error close header")
					break
				}

				DBG_LOG("release header[", val.node.uid, "]")

				nm.headerSetMutex.Lock()
				defer nm.headerSetMutex.Unlock()
				nm.ClrNodeWorkingStatus(true, i)
				
				break;
			}
		}
	}
	
	{
		for i, val := range nm.workers {
			DBG_LOG("worker workingFrp[", val.workingAtFrp, "] val.isWorking[", val.isWorking, "]")
			if val.workingAtFrp == frp_port_str && val.isWorking == true {

				ret := tcp_server_manager.Send_Msg_To_Node(val.node.uid, close_message, false)
				 
				if !ret{
					DBG_LOG("Error close worker")
					break
				}

				DBG_LOG("release worker[", val.node.uid, "]")

				nm.workerSetMutex.Lock()
				defer nm.workerSetMutex.Unlock()
				nm.ClrNodeWorkingStatus(false, i)
			}
		}
	}

	return true
}

func (nm *NodeManager) InitByDatabase() {

	var ok bool
	ch := make(chan interface{})
	var retData interface{}

	/*
	//init header
	mysql_ch <- mysql_op_data{innerChannel:ch, operator:GetHeaderNode}
	retData = <-ch
	
	nm.headers, ok = retData.([]Node_Info)
	if !ok {
		DBG_LOG("Type assertion failed")
		return
	}

	//init worker
	mysql_ch <- mysql_op_data{innerChannel:ch, operator:GetWorkerNode}
	retData = <-ch
	
	nm.workers, ok = retData.([]Node_Info)
	if !ok {
		DBG_LOG("Type assertion failed")
		return
	}
	*/
	
	//init frp 
	mysql_ch <- mysql_op_data{innerChannel:ch, operator:GetFRPInfo}
	retData = <-ch
	
	nm.frp_clients, ok = retData.([]FrpClient)
	if !ok {
		DBG_LOG("Type assertion failed")
		return
	}

    for range nm.frp_clients{
        nm.frpPortIsWork = append(nm.frpPortIsWork, false)
    }

	/*
	DBG_LOG("frp num:", len(nm.frp_clients))

	for _, val := range nm.headers{
		DBG_LOG("Init header ", val.node.uid)
	}
	
	for _, val := range nm.workers{
		DBG_LOG("Init worker ", val.node.uid)
	}
	*/
}
