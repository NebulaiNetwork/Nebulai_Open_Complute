

pub struct StatusManager{
	pub cacl_time_ms i64
}


impl StatusManager{
	
	pub fn test_calc_speed() -> (i64, usize){
		let mut rng = rand::thread_rng();
		
		let mut vec = Vec::new();
		let start_time: DateTime<Utc> = Utc::now();
		let start_timestamp_nanos = start_time.timestamp_nanos_opt().unwrap();
		
		for _ in 0..2000000 {
			
			let random_float1: f64 = rng.gen();
			let random_float2: f64 = rng.gen();

			let float_result = random_float1 + random_float2;
			vec.push(float_result);
			
			let float_result = random_float1 - random_float2;
			vec.push(float_result);
			
			let float_result = random_float1 * random_float2;
			vec.push(float_result);
			
			let float_result = random_float1 / random_float2;
			vec.push(float_result);
		}

		//thread::sleep(Duration::from_secs(2));

		let end_time: DateTime<Utc> = Utc::now();
		let end_timestamp_nanos = end_time.timestamp_nanos_opt().unwrap();

		(((end_timestamp_nanos - start_timestamp_nanos) / 1000000), vec.len())
	}
	
}