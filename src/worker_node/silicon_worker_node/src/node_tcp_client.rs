use std::process::{Command, Stdio};
use std::io::{self, Read, BufRead, BufReader, Write};
use std::net::TcpStream;

use crate::public_func;
use crate::DBG_LOG;
use crate::config;

static SERVER_ADDRESS:&str = "arbitrumpepe.com:6400";
static RAY_HEADER_ADDRESS:&str = "arbitrumpepe.com:";
static CONFIG_FILE_PATH:&str = "./config.json"; // "/etc/node_config/config.json"
static FRPC_FILE_PATH:&str = "./frpc.toml"; // "/usr/bin/frp/frpc.toml"
static FRPC_PATH:&str = "./frpc"; // "/usr/bin/frp/frpc"

pub struct NodeTCPClient {
	pub uid: String,
    pub account: String,
	pub password: String,
}

impl NodeTCPClient {
    pub fn new() -> NodeTCPClient {
		
		
		let worker_config = config::read_config(CONFIG_FILE_PATH);
		
		NodeTCPClient {
			uid: worker_config.uid,
			account: worker_config.account,
			password: worker_config.password,
        }
    }
	
	pub fn start_client(&self){

		// connect address
		let mut stream = TcpStream::connect(SERVER_ADDRESS).expect("Unable to connect server");
		DBG_LOG!("Connected to :", SERVER_ADDRESS);

		let reg_data = self.uid.to_owned() + "|" + &self.account + "|" + &self.password + "|false";

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
	
	fn start_ray_head_node(&self, passwd_and_frpc: &str) -> bool {
		DBG_LOG!("Starting Ray header node...");

		let mut result = true;

		let (passwd, frpc_port) = public_func::split_str_by_char(passwd_and_frpc, '|');

		let output = Command::new("ray")
			.args(&["start", "--head", "--port=7799", &("--redis-password='".to_owned() + passwd + "'")])
			.stdout(Stdio::null()) //Stdio::inherit() to show all msg from ray
			.stderr(Stdio::null())
			.output()
			.expect("Failed to start Ray header node");

		if output.status.success() {
			DBG_LOG!("Ray header node started successfully");
		} else {
			DBG_LOG!("Failed to start Ray header node.");
			result = false
		}
		
		//set frpc_port
		let op_result = public_func::change_file_setting(FRPC_FILE_PATH, "remotePort ", frpc_port);
		DBG_LOG!("frpc_port = ", frpc_port, "  FRPC_FILE_PATH = ", FRPC_FILE_PATH);
		
		if op_result == false {
			DBG_LOG!("Failed to write setting");
			result = false
		}else{
			let child = Command::new(FRPC_PATH)
				.args(&["-c", FRPC_FILE_PATH])
				.stdout(Stdio::null())
				.stderr(Stdio::null())
				.spawn()
				.expect("Failed to start frpc");
			DBG_LOG!("frpc started with PID:[", child.id(), "]");
		}
		
		result
	}

	fn start_ray_worker_node(&self, passwd_and_frpc: &str) -> bool {
		DBG_LOG!("Starting Ray worker node...");

		let mut result = true;

		let (passwd, frpc_port) = public_func::split_str_by_char(passwd_and_frpc, '|');

		let addr_cmd = &("--address=".to_owned() + RAY_HEADER_ADDRESS + frpc_port);
		DBG_LOG!("target ip[", addr_cmd, "]");

		let output = Command::new("ray")
			.args(&["start", addr_cmd, &("--redis-password='".to_owned() + passwd + "'")])
			.stdout(Stdio::inherit()) //Stdio::inherit() to show all msg from ray
			.stderr(Stdio::inherit())
			.output()
			.expect("Failed to start Ray worker node");

		if output.status.success() {
			DBG_LOG!("Ray worker node started successfully");
		} else {
			DBG_LOG!("Failed to start Ray worker node.");
			result = false
		}
		
		result
	}

	fn stop_ray_node(&self) {
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
	}
}
