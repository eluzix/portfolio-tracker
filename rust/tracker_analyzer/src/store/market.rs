use crate::store::tracker_config;
use serde::{Deserialize, Serialize};
use std::{
    collections::{HashMap, HashSet},
    fmt::Display,
    str::FromStr,
};

use serde_json::Value;

use crate::store::cache::Cache;

#[derive(Debug, Deserialize, Serialize)]
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

pub struct PricesClient;

pub trait PriceFetcher {
    #![allow(async_fn_in_trait)]
    async fn fetch_prices(symbols: &Vec<String>) -> Option<HashMap<String, SymbolPrice>>;
}

#[derive(Debug, Deserialize)]
struct MarketStackResponse {
    data: Vec<SymbolPrice>,
}

#[cfg(not(test))]
impl PriceFetcher for PricesClient {
    async fn fetch_prices(symbols: &Vec<String>) -> Option<HashMap<String, SymbolPrice>> {
        let key: String = tracker_config::get("marketstack_key").unwrap();
        println!("KEY >>>>>> {:?}", key);
        let client = reqwest::Client::new();
        let res: MarketStackResponse = client
            .get("https://api.marketstack.com/v1/eod/latest")
            .query(&[("symbols", symbols.join(",")), ("access_key", key)])
            .send()
            .await
            .unwrap()
            .json::<MarketStackResponse>()
            .await
            .unwrap();

        // println!("Response: {:?}", res);

        //let data:serde_json::Array = res.get("data").unwrap();
        //println!("DATA: {:?}", data);
        let mut ret: HashMap<String, SymbolPrice> = HashMap::with_capacity(res.data.len());
        for price in res.data {
            ret.insert(price.symbol.clone(), price);
        }
        // println!("RETURN: {:?}", ret);

        Some(ret)
    }
}

#[cfg(test)]
impl PriceFetcher for PricesClient {
    async fn fetch_prices(symbols: &Vec<String>) -> Option<HashMap<String, SymbolPrice>> {
        None
    }
}

pub async fn load_prices<C: Cache + Send + Sync>(
    cache: &C,
    symbols: &[&str],
) -> Option<HashMap<String, SymbolPrice>> {
    let mut prices = HashMap::new();
    let cached_prices = cache.get("prices").await;
    let mut missing_symbols: HashSet<String> = symbols.iter().map(|s| s.to_string()).collect();

    if let Some(cached_prices) = cached_prices {
        for (symbol, price) in cached_prices.as_object().unwrap() {
            missing_symbols.remove(symbol);
            prices.insert(symbol.clone(), SymbolPrice::from_value(price));
        }

        if missing_symbols.is_empty() {
            return Some(prices);
        }
    }

    if missing_symbols.is_empty() {
        return Some(prices);
    }

    let missing_symbols: Vec<String> = missing_symbols.into_iter().collect();
    println!("Fetching prices from the market API: {:?}", missing_symbols);

    let fetched_prices = PricesClient::fetch_prices(&missing_symbols).await;
    // println!(">>>> {:?}", fetched_prices);
    if let Some(price_map) = fetched_prices {
        prices.extend(price_map);
    }

    let s: String = serde_json::to_string(&prices).unwrap();
    cache.set("prices", s, 60 * 60 * 12).await;

    Some(prices)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::store::cache::MockedCache;
    use rand::Rng;

    #[test]
    fn test_symbol_price_from_value() {
        let val = serde_json::json!({
            "symbol": "AAPL",
            "adj_close": 123.45,
        });

        let price = SymbolPrice::from_value(&val);
        assert_eq!(price.symbol, "AAPL");
        assert_eq!(price.adj_close, 123.45);
    }

    #[tokio::test]
    async fn test_load_prices() {
        let mut rng = rand::thread_rng();
        let mocked_symbols = vec!["AAPL", "GOOGL", "MSFT", "XXX"];
        let mocked_prices = vec![rng.gen::<f64>(), rng.gen::<f64>(), rng.gen::<f64>()];
        let mut cache = MockedCache::new();
        let val = serde_json::json!({
            mocked_symbols[0]: {
                "symbol": mocked_symbols[0],
                "adj_close": mocked_prices[0],
            },
            mocked_symbols[1]: {
                "symbol": mocked_symbols[1],
                "adj_close": mocked_prices[1],
            },
            mocked_symbols[2]: {
                "symbol": mocked_symbols[2],
                "adj_close": mocked_prices[2],
            },
        });
        cache.cache.insert("prices".to_string(), val);

        let prices = load_prices(&cache, mocked_symbols.as_slice())
            .await
            .unwrap();
        assert_eq!(prices.len(), 3);
        assert_eq!(prices.get("AAPL").unwrap().adj_close, mocked_prices[0]);
        assert_eq!(prices.get("GOOGL").unwrap().adj_close, mocked_prices[1]);
        assert_eq!(prices.get("MSFT").unwrap().adj_close, mocked_prices[2]);
    }
}
