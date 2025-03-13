use std::collections::HashMap;
use std::sync::{mpsc, Arc, Mutex};
use std::{thread};
use std::thread::JoinHandle;

pub type SendMsg = String;
pub type RecvMsg = bool;

pub struct TaskInfo {
    pub handle: Option<JoinHandle<()>>,
    pub tx: mpsc::Sender<SendMsg>,
    pub rx: Arc<Mutex<mpsc::Receiver<RecvMsg>>>,
}

pub struct ThreadManager {
    pub task_map: HashMap<i32, TaskInfo>,
    pub now_index: i32,
}

impl ThreadManager {
    pub fn new() -> ThreadManager {
        ThreadManager {
            task_map: HashMap::new(),
            now_index: 0,
        }
    }

    pub fn new_task<T: Send + 'static>(&mut self, sub_task: fn(mpsc::Sender<RecvMsg>, Arc<Mutex<mpsc::Receiver<SendMsg>>>, T), default_value : T) -> i32 {
        let (tx, rx) = mpsc::channel::<SendMsg>();
        
        let rx = Arc::new(Mutex::new(rx));

        // Clone the tx to use inside the closure
        let _tx_clone = tx.clone();
        let rx_clone = Arc::clone(&rx);


        let (tx2, rx2) = mpsc::channel::<RecvMsg>();
        
        let rx2 = Arc::new(Mutex::new(rx2));

        // Clone the tx to use inside the closure
        let tx_clone2 = tx2.clone();
        let _rx_clone2 = Arc::clone(&rx2);


        // Spawn the thread
        let handle = thread::spawn(move || {
            sub_task(tx_clone2, rx_clone, default_value);
        });

        // Insert the task info into the map
        self.task_map.insert(self.now_index, TaskInfo { handle: Some(handle), tx:tx, rx:rx2 });

        self.now_index += 1;

        self.now_index - 1
    }

    pub fn send_to_task(&self, task_index: i32, data: SendMsg) -> Result<(), String> {
        if let Some(task_info) = self.task_map.get(&task_index) {
            task_info.tx.send(data).map_err(|e| e.to_string())
        } else {
            Err("Task not found".to_string())
        }
    }

    pub fn recv_from_task(&self, task_index: i32) -> Result<RecvMsg, String> {
        if let Some(task_info) = self.task_map.get(&task_index) {
            let rx = task_info.rx.lock().unwrap();
            rx.recv().map_err(|e| e.to_string())
        } else {
            Err("Task not found".to_string())
        }
    }

    pub fn wait_all_task_finish(&mut self) {
        for (_, value) in &mut self.task_map {
            if let Some(handle) = value.handle.take() {
                handle.join().unwrap();
            }
        }
    }
}

/*
fn example_task(tx: mpsc::Sender<RecvMsg>, rx: Arc<Mutex<mpsc::Receiver<SendMsg>>>, defualt_value: i32) {
    println!("defualt_value {}", defualt_value);
    
    while let Ok(data) = rx.lock().unwrap().recv() {
        println!("Received data1: {}", data);
        
        tx.send(true).unwrap();
    }
}

fn main() {
    let mut manager = ThreadManager::new();
    let task_index = manager.new_task::<i32>(example_task, 2);

    manager.send_to_task(task_index, (&"hello world").to_string()).unwrap();


    let result = manager.recv_from_task(task_index).unwrap();
    println!("Received result2: {}", result);

    manager.wait_all_task_finish();
}
*/