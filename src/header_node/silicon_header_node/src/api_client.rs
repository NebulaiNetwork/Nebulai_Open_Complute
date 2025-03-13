use reqwest::blocking::Client;
use std::error::Error;
use crate::config;
use config::TCPServerConfig;

static CENTRAL_API_ADDRESS:&str = "http://arbitrumpepe.com:6402";
static CONFIG_FILE_PATH:&str = "./config.json"; // "/etc/node_config/config.json"

pub struct APIClient {
	pub uid: String,
    pub account: String,
	pub password: String,
	pub central_id: String,
	
	pub is_header: String,
	
	pub client: Client,
}


impl APIClient{
	
	pub fn new(is_header:String) -> APIClient {
		
		let worker_config = config::read_config(CONFIG_FILE_PATH);
		
		APIClient {
			uid: worker_config.uid,
			account: worker_config.account,
			password: worker_config.password,
			central_id: worker_config.central_id,
			is_header: is_header,
			client: Client::new(),
        }
    }
	
	pub fn reg_this_node(&self) -> Result<TCPServerConfig,  Box<dyn Error>> {
		let get_url = CENTRAL_API_ADDRESS.to_owned() + "/reg?uid=" + &self.uid + "&account=" + &self.account + "&password=" + &self.password + "&isHeader=" + &self.is_header + "&central_id=" + &self.central_id;
		
		let get_response:TCPServerConfig = self.client.get(&get_url).send()?.json()?;
		Ok(get_response)
	}
}