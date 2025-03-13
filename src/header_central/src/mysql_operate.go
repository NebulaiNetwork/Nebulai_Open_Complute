package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

const (
    GetRouteIp = iota
	RegWorker
	UnregWorker
	RegHeader
	UnregHeader
	GetHeaderNode
	GetWorkerNode
	GetFRPInfo
	GetHeaderCentralTCP
	CheckWorkerUidExist
	CheckHeaderUidExist
	CheckApiKeyExist
)

type mysql_op_data struct {
	innerChannel	chan interface{}
	operator		int
	payload			string
	ip				string
	port			string
	uid				string
	account			string
	password		string
}

var mysql_ch chan mysql_op_data

func Start_SQL_Server(){

	mysql_ch = make(chan mysql_op_data)

	dsn := dsn_str
	DBG_LOG("database cmd==>", dsn)
	
	db, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(err)
    }
    defer db.Close()
	
	for{
		data := <-mysql_ch

		switch data.operator {
		case CheckApiKeyExist:
			CheckApiKeyExistOp(db, data)
		case CheckWorkerUidExist:
			CheckWorkerUidExistOp(db, data)
		case CheckHeaderUidExist:
			CheckHeaderUidExistOp(db, data)
		case GetRouteIp:
			QueryRouteIpAddress(db, data.innerChannel)
		case RegWorker:
			RegWorkerOp(db, data)
		case UnregWorker:
			UnRegWorkerOp(db, data)
		case RegHeader:
			RegHeaderOp(db, data)
		case UnregHeader:
			UnRegHeaderOp(db, data)
		case GetHeaderNode:
			GetHeaderNodeOp(db, data)
		case GetWorkerNode:
			GetWorkerNodeOp(db, data)
		case GetFRPInfo:
			GetFRPInfoOp(db, data)
		case GetHeaderCentralTCP:
			GetHeaderCentralTCPOp(db, data);
		}
	}
}

func ApiKeyIsExist(api_key string) bool{
	ch := make(chan interface{})

	mysql_ch <- mysql_op_data{innerChannel:ch, operator:CheckApiKeyExist, uid:api_key}

	result := <-ch

	return result.(bool)
}

func CheckApiKeyExistOp(db *sql.DB, data mysql_op_data){
	
    rows, err := db.Query("SELECT count(api_key) FROM api_info where api_key = \"" + data.uid + "\";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
        var count int
        err := rows.Scan(&count)
        if err != nil {
            panic(err)
        }

		if count == 1{
			data.innerChannel <- true
		}else{
			data.innerChannel <- false
		}
		return 
    }
}

func QueryRouteIpAddress(db *sql.DB, ret_chan chan interface{}){
	DBG_LOG("try query Route Ip")
	
    rows, err := db.Query("SELECT ip_address, api_port FROM header_central_info where id = " + convertToString(this_central_id) + ";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
        var ipAddress string
        var port string
        err := rows.Scan(&ipAddress, &port)
        if err != nil {
            panic(err)
        }
		
		ret := ipAddress + ":" + port
		DBG_LOG("get Route IP:", ret)
		ret_chan <- ret
		return 
    }
	DBG_LOG("error database read nil")
	ret_chan <- ""
}

func RegWorkerOp(db *sql.DB, data mysql_op_data){

	ret := false

	sqlCmd := "select count(t2.uid) from account_info t1 join worker_info t2 on t1.id = t2.account_id where t2.uid = \"" + data.uid + "\" and t1.account = \"" + data.account + "\" and t1.password = \"" + data.password + "\";"

	rows, err := db.Query(sqlCmd)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
		return 
    }
    defer rows.Close()

	var count_ int

    for rows.Next() {

        err := rows.Scan(&count_)
        if err != nil {
        	data.innerChannel <- ret
            DBG_LOG("reg worker err:", err)
            return 
        }

        break
    }

    // check error
    err = rows.Err()
    if err != nil {
   		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
        return
    }

    if count_ > 0{
		
    }

	sqlCmd2 := "update worker_info set header_central_id = ? where uid = ?;"
	
    _, err = db.Exec(sqlCmd2, this_central_id, data.uid)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
    }
    ret = true
	data.innerChannel <- ret
}

func UnRegWorkerOp(db *sql.DB, data mysql_op_data){
	ret := false

	sqlCmd := "select count(t2.uid) from account_info t1 join worker_info t2 on t1.id = t2.account_id where t2.uid = \"" + data.uid + "\" and t1.account = \"" + data.account + "\" and t1.password = \"" + data.password + "\";"

	rows, err := db.Query(sqlCmd)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
		return 
    }
    defer rows.Close()

	var count_ int

    for rows.Next() {

        err := rows.Scan(&count_)
        if err != nil {
        	data.innerChannel <- ret
            DBG_LOG("reg worker err:", err)
            return 
        }

        break
    }

    // check error
    err = rows.Err()
    if err != nil {
   		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
        return
    }

    if count_ > 0{
		
    }

	sqlCmd2 := "update worker_info set header_central_id = ? where uid = ?;"
	
    _, err = db.Exec(sqlCmd2, -1, data.uid)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
    }
    ret = true
	data.innerChannel <- ret
}

func RegHeaderOp(db *sql.DB, data mysql_op_data){
	ret := false

	sqlCmd := "select count(t2.uid) from account_info t1 join header_info t2 on t1.id = t2.account_id where t2.uid = \"" + data.uid + "\" and t1.account = \"" + data.account + "\" and t1.password = \"" + data.password + "\";"

	rows, err := db.Query(sqlCmd)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
		return 
    }
    defer rows.Close()

	var count_ int

    for rows.Next() {

        err := rows.Scan(&count_)
        if err != nil {
        	data.innerChannel <- ret
            DBG_LOG("reg worker err:", err)
            return 
        }

        break
    }

    // check error
    err = rows.Err()
    if err != nil {
   		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
        return
    }

    if count_ > 0{
		
    }

	sqlCmd2 := "update header_info set header_central_id = ? where uid = ?;"
	
    _, err = db.Exec(sqlCmd2, this_central_id, data.uid)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
    }
    ret = true
	data.innerChannel <- ret
}

func UnRegHeaderOp(db *sql.DB, data mysql_op_data){
	ret := false

	sqlCmd := "select count(t2.uid) from account_info t1 join header_info t2 on t1.id = t2.account_id where t2.uid = \"" + data.uid + "\" and t1.account = \"" + data.account + "\" and t1.password = \"" + data.password + "\";"

	rows, err := db.Query(sqlCmd)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
		return 
    }
    defer rows.Close()

	var count_ int

    for rows.Next() {

        err := rows.Scan(&count_)
        if err != nil {
        	data.innerChannel <- ret
            DBG_LOG("reg worker err:", err)
            return 
        }

        break
    }

    // check error
    err = rows.Err()
    if err != nil {
   		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
        return
    }

    if count_ > 0{
		
    }

	sqlCmd2 := "update header_info set header_central_id = ? where uid = ?;"
	
    _, err = db.Exec(sqlCmd2, -1, data.uid)
    if err != nil {
		data.innerChannel <- ret
        DBG_LOG("reg worker err:", err)
    }
    ret = true
	data.innerChannel <- ret
}

func GetHeaderNodeOp(db *sql.DB, sql_op_data mysql_op_data){

	var ret []Node_Info

	sqlCmd := "select uid from header_info where header_central_id = " + convertToString(this_central_id) + ";"
	
	rows, err := db.Query(sqlCmd)
    if err != nil {
		sql_op_data.innerChannel <- ret
        panic(err)
		return 
    }
    defer rows.Close()

    for rows.Next() {
        var uid_ string

        err := rows.Scan(&uid_)
        if err != nil {
            panic(err)
        }
		ret = append(ret, Node_Info{node:Node{uid:uid_}, isWorking:false, workingAtFrp:""})
    }

    // check error
    err = rows.Err()
    if err != nil {
        panic(err)
    }
	
	sql_op_data.innerChannel <- ret
}

func GetWorkerNodeOp(db *sql.DB, sql_op_data mysql_op_data){

	var ret []Node_Info

	sqlCmd := "select uid from worker_info where header_central_id = " + convertToString(this_central_id) + ";"
	
    rows, err := db.Query(sqlCmd)
    if err != nil {
		sql_op_data.innerChannel <- ret
        panic(err)
		return 
    }
    defer rows.Close()

    for rows.Next() {
        var uid_ string

        err := rows.Scan(&uid_)
        if err != nil {
            panic(err)
        }
		ret = append(ret, Node_Info{node:Node{uid:uid_}, isWorking:false, workingAtFrp:""})
    }
	
	// check error
    err = rows.Err()
    if err != nil {
        panic(err)
    }
	
	sql_op_data.innerChannel <- ret
}

func GetFRPInfoOp(db *sql.DB, sql_op_data mysql_op_data){
	var ret []FrpClient

	sqlCmd := "select ip_address, header_port, obj_port, dashboard_port from frpc_info where header_central_id = " + convertToString(this_central_id) + ";"
	
    rows, err := db.Query(sqlCmd)
    if err != nil {
		sql_op_data.innerChannel <- ret
        panic(err)
		return 
    }
    defer rows.Close()

    for rows.Next() {
        var ip_address string
		var header_port string
		var obj_port string
		var dashboard_port string

        err := rows.Scan(&ip_address, &header_port, &obj_port, &dashboard_port)
        if err != nil {
            panic(err)
        }
		ret = append(ret, FrpClient{ip:ip_address, port:header_port, obj_port:obj_port, dashboard_port:dashboard_port})
    }
	
	// check error
    err = rows.Err()
    if err != nil {
        panic(err)
    }
	
	sql_op_data.innerChannel <- ret
}

func GetHeaderCentralTCPOp(db *sql.DB, sql_op_data mysql_op_data){
	var ret Header_Central_Info
	
    rows, err := db.Query("SELECT id, ip_address, api_port, tcp_port FROM header_central_info where id = " + convertToString(this_central_id) + ";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
    	var id_ int
        var ipAddress string
        var api_port_ string
        var tcp_port_ string
        err := rows.Scan(&id_, &ipAddress, &api_port_, &tcp_port_)
        if err != nil {
            panic(err)
        }

		ret = Header_Central_Info{id:id_, ip:ipAddress, api_port:api_port_, tcp_port:tcp_port_}
		break
    }

    sql_op_data.innerChannel <- ret
}

func CheckWorkerUidExistOp(db *sql.DB, user_data mysql_op_data) {

    rows, err := db.Query("select count(t2.uid) from account_info t1 join worker_info t2 on t1.id = t2.account_id where t2.uid = \"" + user_data.uid + "\" and t1.account = \"" + user_data.account + "\" and t1.password = \"" + user_data.password + "\";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
		var count_ int
		err := rows.Scan(&count_)
		if err != nil {
			panic(err)
		}

		if count_ > 0{
			user_data.innerChannel <- true
		}else{
			user_data.innerChannel <- false
		}

		return 
	}

	user_data.innerChannel <- false
}

func CheckHeaderUidExistOp(db *sql.DB, user_data mysql_op_data) {

    rows, err := db.Query("select count(t2.uid) from account_info t1 join header_info t2 on t1.id = t2.account_id where t2.uid = \"" + user_data.uid + "\" and t1.account = \"" + user_data.account + "\" and t1.password = \"" + user_data.password + "\";")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	
    for rows.Next() {
		var count_ int
		err := rows.Scan(&count_)
		if err != nil {
			panic(err)
		}

		if count_ > 0{
			user_data.innerChannel <- true
		}else{
			user_data.innerChannel <- false
		}

		return 
	}

	user_data.innerChannel <- false
}


