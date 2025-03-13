/*
use std::process::{Command, Stdio};
use std::io::{self, Read, BufRead, BufReader, Write};
use std::net::TcpStream;
use std::borrow::Cow;


*/
mod config;

#[macro_use]
mod public_func;
mod node_tcp_client;
use node_tcp_client::NodeTCPClient;

mod api_client;
use api_client::APIClient;

fn main(){
	
	let is_header = String::from("true");
	
	let api_client_ = APIClient::new(is_header.clone());

	let tcp_server_config = match api_client_.reg_this_node(){
		Ok(result) => result,
		Err(e) => {
			DBG_LOG!("Error ", e);
			config::TCPServerConfig{code:-1, ip:(&"").to_string(), port:(&"").to_string()}
		},
	};
	
	if tcp_server_config.code == -1{
		DBG_LOG!("api client init failed");
		return 
	}
	
	let mut node_client = NodeTCPClient::new(is_header.clone());
	node_client.start_client(&tcp_server_config.ip, &tcp_server_config.port);
}



