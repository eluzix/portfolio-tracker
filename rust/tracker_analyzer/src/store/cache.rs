use std::sync::Arc;
use std::time::SystemTime;
use std::{collections::HashMap, hash::Hash};

use aws_config::imds::client::error::ErrorResponse;
use aws_sdk_dynamodb::types::AttributeValue;
use chrono::{Datelike, Duration, NaiveDate};
use once_cell::sync::Lazy;
use serde_json::{json, Value};
use tokio::sync::Mutex;

use crate::store::ddb::get_client;

pub trait Cache {
    fn get(&self, key: &str) -> impl std::future::Future<Output = Option<Value>> + Send;
    fn set(&self, key: &str, value: String, ttl: u64) -> impl std::future::Future<Output = ()>;
}

pub struct DynamoCache {
    cache: Mutex<HashMap<String, Value>>,
}

impl DynamoCache {
    pub fn new() -> Self {
        Self {
            cache: Mutex::new(HashMap::new()),
        }
    }
}

impl Cache for DynamoCache {
    async fn get(&self, key: &str) -> Option<Value> {
        {
            let cache_lock = self.cache.lock().await;
            if let Some(value) = cache_lock.get(key) {
                return Some(value.clone());
            }
        } // Release the lock before performing the network request

        let client = get_client().await.unwrap();
        let result = client
            .get_item()
            .table_name("tracker-data")
            .key("PK", AttributeValue::S("CACHE".into()))
            .key("SK", AttributeValue::S(key.to_string()))
            .send()
            .await
            .unwrap();

        if let Some(item) = result.item {
            let now = SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs();

            println!("Found item: {:?}, NOW ==> {}", item, now);

            if let Some(AttributeValue::N(ttl)) = item.get("ttl") {
                if now > ttl.as_str().parse::<u64>().unwrap() {
                    return None;
                }
            }

            let val: Option<_> = match item.get("value") {
                Some(AttributeValue::B(blob)) => {
                    if let Ok(json_str) = std::str::from_utf8(blob.as_ref()) {
                        Some(json_str)
                    } else {
                        None
                    }
                }

                Some(AttributeValue::S(s)) => Some(s.as_str()),

                _ => None,
            };

            if let Some(json_str) = val {
                if let Ok(json) = serde_json::from_str(json_str) {
                    let mut cache_lock = self.cache.lock().await;
                    let json: Value = json;
                    cache_lock.insert(key.to_string(), json.clone());
                    return Some(json);
                }
            }
        }

        None

        //     if let Some(AttributeValue::B(blob)) = item.get("value") {
        //         if let Ok(json_str) = std::str::from_utf8(blob.as_ref()) {
        //             if let Ok(json) = serde_json::from_str(json_str) {
        //                 let mut cache_lock = self.cache.lock().await;
        //                 let json: Value = json;
        //                 cache_lock.insert(key.to_string(), json.clone());
        //                 return Some(json);
        //             }
        //         }
        //     }
        // }

        // None
    }

    async fn set(&self, key: &str, value: String, ttl: u64) {
        let client = get_client().await.unwrap();
        // let val = serde_json::to_string(&value).unwrap();
        let ttl = SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)
            .unwrap()
            .as_secs()
            + ttl;

        println!("SETTING item {:?} with ttl ====>>> {}", value, ttl);

        let res = client
            .put_item()
            .table_name("tracker-data")
            .item("PK", AttributeValue::S("CACHE".to_string()))
            .item("SK", AttributeValue::S(key.to_string()))
            .item("value", AttributeValue::S(value))
            .item("ttl", AttributeValue::N(format!("{}", ttl)))
            .send()
            .await;

        if let Result::Err(err) = res {
            println!("ERROR Setting cache for {:?}, err: {:?}", key, err);
        }
    }
}

static CACHE: Lazy<Arc<DynamoCache>> = Lazy::new(|| Arc::new(DynamoCache::new()));

pub fn default_cache() -> Arc<DynamoCache> {
    CACHE.clone()
}

#[cfg(test)]
pub struct MockedCache {
    pub cache: HashMap<String, Value>,
}

#[cfg(test)]
impl MockedCache {
    pub fn new() -> Self {
        Self {
            cache: HashMap::new(),
        }
    }
}

#[cfg(test)]
impl Cache for MockedCache {
    fn get(&self, key: &str) -> impl std::future::Future<Output = Option<Value>> + Send {
        async move { self.cache.get(key).cloned() }
    }

    fn set(&self, key: &str, value: String, ttl: u64) -> impl std::future::Future<Output = ()> {
        ()
    }
}
