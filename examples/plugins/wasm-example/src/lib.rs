use serde::{Deserialize, Serialize};
use std::ffi::CStr;
use std::os::raw::c_char;

#[derive(Deserialize)]
struct Input {
    text: String,
}

#[derive(Serialize)]
struct Output {
    result: String,
    status: String,
    length: usize,
}

#[no_mangle]
pub extern "C" fn execute(input_ptr: *const c_char) -> *mut c_char {
    // Safety: Convert C string to Rust string
    let input_str = unsafe {
        if input_ptr.is_null() {
            return std::ptr::null_mut();
        }
        CStr::from_ptr(input_ptr)
            .to_str()
            .unwrap_or("{}")
    };
    
    // Parse input JSON
    let input: Input = match serde_json::from_str(input_str) {
        Ok(i) => i,
        Err(_) => return create_error_response("Invalid input JSON"),
    };
    
    // Process
    let result = format!("Processed: {}", input.text);
    let length = result.len();
    
    // Create output
    let output = Output {
        result: result.clone(),
        status: "success".to_string(),
        length,
    };
    
    // Serialize to JSON
    let json = match serde_json::to_string(&output) {
        Ok(j) => j,
        Err(_) => return create_error_response("Failed to serialize output"),
    };
    
    // Convert to C string
    let c_string = std::ffi::CString::new(json).unwrap();
    c_string.into_raw()
}

fn create_error_response(error: &str) -> *mut c_char {
    let error_json = format!(r#"{{"status":"error","error":"{}"}}"#, error);
    let c_string = std::ffi::CString::new(error_json).unwrap();
    c_string.into_raw()
}

// Memory deallocation function
#[no_mangle]
pub extern "C" fn free_string(ptr: *mut c_char) {
    if !ptr.is_null() {
        unsafe {
            let _ = std::ffi::CString::from_raw(ptr);
        }
    }
}
