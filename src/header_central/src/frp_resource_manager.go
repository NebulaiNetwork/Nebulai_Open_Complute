package main

var frp_resource_manager FRPResourceManager

type IP_Port struct{
	ip		uint32
	port	uint32
}

type FRPResourceManager struct{
	ip		[]IP_Port
	working []bool
}

func (frp_r_m *FRPResourceManager) InitFrpByDatabase() {
/*
	ch := make(chan interface{})
	mysql_ch <- mysql_op_data{innerChannel:ch, operator:GetFRPResource}
	retData := <-ch
	
	nm.ip, ok := retData.([]IP_Port)
	if !ok {
		DBG_LOG("Type assertion failed")
		return
	}
	
	for range nm.ip {
		nm.working = append(nm.working, false)
	}

	DBG_LOG("header num:", len(nm.headerIsWork), " worker num:", len(nm.workerIsWork))
	
	*/
}