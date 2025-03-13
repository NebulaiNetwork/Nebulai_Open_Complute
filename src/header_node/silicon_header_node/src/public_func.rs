use std::fs;

#[macro_export]
macro_rules! DBG_LOG {
    ($($arg:expr),*) => {
        {
            let file = std::file!();
            let line = std::line!();

            let args = vec![$(format!("{}", $arg)),*];

            let args_str = args.join("");

            println!("{:<20}|{:^5}| logs: {}", file, line, args_str);
        }
    };
}

pub fn split_str_by_char(cut_str: &str, by_char: char) -> (&str, &str){
    
    if cut_str.len() == 0{
        return (cut_str, "");
    }
    
    let mut index_i = 0;
    
    for i in cut_str.chars(){
        index_i += 1;
        if i == by_char{
            break; 
        }
    }
    
    (&cut_str[..(index_i - 1)], &cut_str[index_i..])
}

pub fn change_file_setting(file_path: &str, target_line_str: &str, new_option: &str, change_config_index : i32) -> bool{
	//DBG_LOG!("read file[", file_path, "]");
	let file_data = fs::read_to_string(file_path).expect("Unable to read file");
	
	let target_settting_name_len = target_line_str.len();
	
	let mut result = String::from("");
	
	let mut find_set = false;
	
	let mut index = change_config_index;
	
	for line in file_data.lines(){
	    
	    if line.len() < target_settting_name_len{
	        continue;
	    }
	   
	    let (set_name, _) = split_str_by_char(line, '=');
	    
		//DBG_LOG!("target_line_str[", target_line_str, "] set_name[", set_name, "]");
		
	    if set_name == target_line_str{
			index -= 1;
			if index == 0{
				
				result.push_str(&(set_name.to_owned() + "=" + new_option + "\n"));
				find_set = true;
				continue;
			}
	    }
        
        result.push_str(&(line.to_owned() + "\n"));
    }
    
	let mut wf_succ = false;
	
	match fs::write(file_path, result) {  
        Ok(_) => wf_succ = true,  
        Err(_) => {
			DBG_LOG!("write file error:");
        }  
    }  
	
	//DBG_LOG!("wf_succ:", wf_succ, "  find_set:", find_set);
	wf_succ && find_set
}

