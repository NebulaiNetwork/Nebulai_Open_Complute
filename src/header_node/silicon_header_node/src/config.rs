
use serde::{Serialize, Deserialize};
use serde_json;
use std::fs;

#[derive(Serialize, Deserialize, Debug)]
pub struct Config {
    pub uid: String,
	pub account: String,
	pub password: String,
	pub central_id: String,
}

pub fn read_config(file_path: &str) -> Config {
	let config_data = fs::read_to_string(file_path).expect("Unable to read file");
	let c: Config = serde_json::from_str(&config_data).expect("JSON was not well-formatted");

    c
}

#[derive(Serialize, Deserialize, Debug)]
pub struct TCPServerConfig {
	pub code: i32,
    pub ip: String,
	pub port: String,
}

/*
pub fn read_header_central_data(data_info: &str) -> TCPServerConfig {
	let c: TCPServerConfig = serde_json::from_str(&data_info).expect("JSON was not well-formatted");

    c
}
*/