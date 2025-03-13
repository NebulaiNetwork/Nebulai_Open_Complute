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

fn main(){
	let node_client = NodeTCPClient::new();
	node_client.start_client();
}



