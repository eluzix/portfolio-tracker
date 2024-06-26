use std::collections::HashMap;

use aws_sdk_dynamodb::types::AttributeValue;
use once_cell::sync::Lazy;
use serde_json::Value;
use tokio::sync::Mutex;

use crate::store::ddb::get_client;

static CACHE: Lazy<Mutex<HashMap<String, Value>>> = Lazy::new(|| {
    Mutex::new(HashMap::new())
});

pub async fn get(key: &str) -> Option<Value> {
    {
        let cache_lock = CACHE.lock().await;
        if let Some(value) = cache_lock.get(key) {
            return Some(value.clone());
        }
    } // Release the lock before performing the network request

    let client = get_client().await.unwrap();
    let result = client.get_item()
        .table_name("tracker-data")
        .key("PK", AttributeValue::S("CACHE".into()))
        .key("SK", AttributeValue::S(key.to_string()))
        .send()
        .await.unwrap();

    if let Some(item) = result.item {
        if let Some(AttributeValue::B(blob)) = item.get("value") {
            if let Ok(json_str) = std::str::from_utf8(blob.as_ref()) {
                if let Ok(json) = serde_json::from_str(json_str) {
                    let mut cache_lock = CACHE.lock().await;
                    let json: Value = json;
                    cache_lock.insert(key.to_string(), json.clone());
                    return Some(json);
                }
            }
        }
    }

    None
}
