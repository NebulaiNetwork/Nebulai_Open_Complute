use std::process::{Command, Stdio};
//use std::io::{self, Read, BufRead, BufReader, Write};
use std::io::{Read, Write};
use std::net::TcpStream;

use crate::public_func;
use crate::DBG_LOG;
use crate::config;

//static SERVER_ADDRESS:&str = "arbitrumpepe.com:6502";
static RAY_HEADER_ADDRESS:&str = "arbitrumpepe.com:";
static CONFIG_FILE_PATH:&str = "./config.json"; // "/etc/node_config/config.json"
static FRPC_FILE_PATH:&str = "./frpc.toml"; // "/usr/bin/frp/frpc.toml"
static FRPC_PATH:&str = "./frpc"; // "/usr/bin/frp/frpc"

pub struct NodeTCPClient {
	pub uid: String,
    pub account: String,
	pub password: String,
	
	pub is_header: String,
	
	pub frpc_dip: u32,
}

impl NodeTCPClient {
    pub fn new(is_header:String) -> NodeTCPClient {
		
		
		let worker_config = config::read_config(CONFIG_FILE_PATH);
		DBG_LOG!("Init by uid[", worker_config.uid, "] account[", worker_config.account, "] password[", worker_config.password, "]");
		NodeTCPClient {
			uid: worker_config.uid,
			account: worker_config.account,
			password: worker_config.password,
			is_header:is_header,
			frpc_dip:0
        }
    }
	
	pub fn start_client(&mut self, ip:&str, port:&str){

		let server_address = ip.to_owned() + ":" + port;

		// connect address
		let mut stream = TcpStream::connect(server_address.clone()).expect("Unable to connect server");
		DBG_LOG!("Connected to :", server_address);

		let reg_data = self.uid.to_owned() + "|" + &self.account + "|" + &self.password + "|" + &self.is_header;

		stream.write_all(reg_data.as_bytes()).expect("Send Data to server failed");

		// create a buffer to read msg from server
		let mut buffer = [0; 512];

		// loop read msg from server
		loop {
			match stream.read(&mut buffer) {
				Ok(0) => {
					// server close connect
					DBG_LOG!("Server closed connection");
					break;
				}
				Ok(n) => {
					let received = String::from_utf8_lossy(&buffer[..n]);
					DBG_LOG!("Received: :", received);
					
					if received.len() <= 3 {
						DBG_LOG!("error cmd len to low");
						continue;
					}
					match &received[..2]{
						"00" => {
							DBG_LOG!("recv server msg:", &received[3..]);
							stream.write_all((&received[3..]).as_bytes()).expect("Send Data to server failed");
						},
						"01" => {
							let result = self.start_ray_head_node(&received[3..]);
							if result == false{
								let ret_msg = reg_data.to_owned() + "|4001";
								stream.write_all(ret_msg.as_bytes()).expect("Send Data to server failed");
							}else {
								let ret_msg = reg_data.to_owned() + "|1";
								stream.write_all(ret_msg.as_bytes()).expect("Send Data to server failed");
							}
						},
						"02" => {
							let result = self.start_ray_worker_node(&received[3..]);
							
							if result == false{
								let ret_msg = reg_data.to_owned() + "|4002";
								stream.write_all(ret_msg.as_bytes()).expect("Send Data to server failed");
							}else{
								let ret_msg = reg_data.to_owned() + "|1";
								stream.write_all(ret_msg.as_bytes()).expect("Send Data to server failed");
							}
						},
						"03" => self.stop_ray_node(),
						_ => DBG_LOG!("do not"),
					}
				}
				Err(e) => {
					DBG_LOG!("Error reading from server: ", e);
					break;
				}
			}
		}
		
	}
	
	fn start_ray_head_node(&mut self, passwd_and_ports: &str) -> bool {
		DBG_LOG!("Starting Ray header node...");

		let mut result = false;
		let mut try_time = 5;

		let (passwd, next_port_msg) = public_func::split_str_by_char(passwd_and_ports, '|');
		let (frpc_port, next_port_msg) = public_func::split_str_by_char(next_port_msg, '|');
		let (obj_port, dashboard_port) = public_func::split_str_by_char(next_port_msg, '|');

		//DBG_LOG!("passwd[", passwd, "] frpc_port[", frpc_port, "] pbj_port[", obj_port, "] dashboard_port[", dashboard_port, "]");

		while !result && try_time > 0{
			let output = Command::new("ray")
			.args(&["start", "--head", "--port=7799", &("--redis-password='".to_owned() + passwd + "'")])
			.stdout(Stdio::null()) //Stdio::inherit() to show all msg from ray
			.stderr(Stdio::null())
			.output()
			.expect("Failed to start Ray header node");

			if output.status.success() {
				DBG_LOG!("Ray header node started successfully");
				result = true;
			} else {
				DBG_LOG!("Failed to start Ray header node.");
				try_time -= 1;
			}
		}
		
		result = false;
		
		while !result && try_time > 0{
			//set frpc_port
			result = public_func::change_file_setting(FRPC_FILE_PATH, "remotePort ", frpc_port, 1);
			result = result && public_func::change_file_setting(FRPC_FILE_PATH, "remotePort ", obj_port, 2);
			result = result && public_func::change_file_setting(FRPC_FILE_PATH, "remotePort ", dashboard_port, 3);
			
			DBG_LOG!("frpc_port = ", frpc_port, "  FRPC_FILE_PATH = ", FRPC_FILE_PATH);
			
			if result == false {
				DBG_LOG!("Failed to write setting");
				try_time -= 1;
			}else{
				let child = Command::new(FRPC_PATH)
					.args(&["-c", FRPC_FILE_PATH])
					.stdout(Stdio::null())
					.stderr(Stdio::null())
					.spawn()
					.expect("Failed to start frpc");
				DBG_LOG!("frpc started with PID:[", child.id(), "]");
				self.frpc_dip = child.id();
			}
		}
		
		result
	}

	fn start_ray_worker_node(&self, passwd_and_frpc: &str) -> bool {
		DBG_LOG!("Starting Ray worker node...");

		let mut result = false;
		let mut try_time = 5;

		let (passwd, frpc_port) = public_func::split_str_by_char(passwd_and_frpc, '|');

		while !result && try_time > 0{
			
			let output = Command::new("ray")
			.args(&["start", &("--address=".to_owned() + RAY_HEADER_ADDRESS + frpc_port), &("--redis-password='".to_owned() + passwd + "'")])
			.stdout(Stdio::null()) //Stdio::inherit() to show all msg from ray
			.stderr(Stdio::null())
			.output()
			.expect("Failed to start Ray worker node");

			if output.status.success() {
				DBG_LOG!("Ray worker node started successfully");
				result = true;
			} else {
				DBG_LOG!("Failed to start Ray worker node.");
				try_time -= 1;
			}
		}

		result
	}

	fn stop_ray_node(&mut self) {
		DBG_LOG!("Stopping Ray worker node...");

		let output = Command::new("ray")
			.args(&["stop"])
			.stdout(Stdio::null())
			.stderr(Stdio::null())
			.output()
			.expect("Failed to stop Ray worker node");

		if output.status.success() {
			DBG_LOG!("Ray worker node stopped successfully.");
		} else {
			DBG_LOG!("Failed to stop Ray worker node.");
		}
		
		if self.frpc_dip != 0{
			if kill_process(self.frpc_dip as i32) == true{
				self.frpc_dip = 0;
			}else {
				DBG_LOG!("Failed to stop Frpc.");
			}
		}
	}
}


fn kill_process(pid: i32) -> bool {
	DBG_LOG!("kill process[", pid, "]");
    let output = Command::new("kill")
        .arg("-TERM")  
        .arg(pid.to_string())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .output()
        .expect("Failed to send SIGTERM signal");

    if output.status.success() {
        // wait process stop
        std::thread::sleep(std::time::Duration::from_secs(1));
        
        // check
        let output = Command::new("kill")
            .arg("-0")  //still exist 
            .arg(pid.to_string())
            .output()
            .expect("Failed to check process status");

        if output.status.success() {
            // force stop it 
            let output = Command::new("kill")
                .arg("-KILL")
                .arg(pid.to_string())
                .stdout(Stdio::null())
                .stderr(Stdio::null())
                .output()
                .expect("Failed to send SIGKILL signal");

            return output.status.success();
        }

        true
    } else {
        false
    }
}