use std::collections::HashMap;
use serde_json::Value;
use crate::store::cache;

#[derive(Debug)]
pub struct SymbolPrice {
    pub symbol: String,
    pub adj_close: f64,
}

impl SymbolPrice {
    pub fn from_value(val: &Value) -> SymbolPrice {
        SymbolPrice {
            symbol: val["symbol"].as_str().unwrap().to_string(),
            adj_close: val["adj_close"].as_f64().unwrap(),
        }
    }
}

pub async  fn load_prices() -> Option<HashMap<String, SymbolPrice>> {
    let cached_prices = cache::get("prices").await;
    if let Some(cached_prices) = cached_prices {
        let mut prices = HashMap::new();
        for (symbol, price) in cached_prices.as_object().unwrap() {
            prices.insert(symbol.clone(), SymbolPrice::from_value(price));
        }
        return Some(prices);
    }

    None
}
