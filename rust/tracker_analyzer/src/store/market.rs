use crate::{
    store::tracker_config,
    types::transactions::{Transaction, TransactionType},
};
use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};

use serde_json::Value;

use crate::store::cache::Cache;

#[derive(Debug, Deserialize, Serialize, Default)]
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

pub struct MarketDataClient;

pub trait MarketDataFetcher {
    #![allow(async_fn_in_trait)]
    async fn fetch_prices(symbols: &[String]) -> Option<HashMap<String, SymbolPrice>>;
    async fn fetch_dividends(symbols: &[String]) -> Option<HashMap<String, Vec<Transaction>>>;
}

#[derive(Debug, Deserialize)]
pub struct MarketStackResponse {
    data: Vec<SymbolPrice>,
    // pagination: HashMap<String, u32>,
}

#[derive(Debug, Deserialize)]
pub struct MarketDividend {
    date: String,
    dividend: f64,
    symbol: String, // pagination: HashMap<String, u32>,
}

#[derive(Debug, Deserialize)]
pub struct MarketDividendsResponse {
    data: Vec<MarketDividend>,
    // pagination: HashMap<String, u32>,
}

#[cfg(not(test))]
impl MarketDataFetcher for MarketDataClient {
    async fn fetch_prices(symbols: &[String]) -> Option<HashMap<String, SymbolPrice>> {
        let key: String = tracker_config::get("marketstack_key").unwrap();

        let client = reqwest::Client::new();

        let res: String = client
            .get("https://api.marketstack.com/v1/eod/latest")
            .query(&[("symbols", symbols.join(",")), ("access_key", key)])
            .send()
            .await
            .unwrap()
            .text()
            .await
            .unwrap();

        let res = res.replace(",[]", "");

        let res = serde_json::from_str::<MarketStackResponse>(&res);

        match res {
            Ok(js) => {
                let mut ret: HashMap<String, SymbolPrice> = HashMap::with_capacity(js.data.len());
                for price in js.data {
                    ret.insert(price.symbol.clone(), price);
                }

                Some(ret)
            }

            Err(err) => {
                println!(
                    "[fetch_prices] Error loading data from marketstack, error: {:?}",
                    err
                );
                None
            }
        }
    }

    async fn fetch_dividends(symbols: &[String]) -> Option<HashMap<String, Vec<Transaction>>> {
        let key: String = tracker_config::get("marketstack_key").unwrap();
        let client = reqwest::Client::new();
        let res = client
            .get("https://api.marketstack.com/v1/dividends")
            .query(&[
                ("symbols", symbols.join(",")),
                ("access_key", key),
                ("limit", "1000".to_string()),
            ])
            .send()
            .await
            .unwrap()
            .json::<MarketDividendsResponse>()
            .await;

        if let Ok(dividend_response) = res {
            let mut res: HashMap<String, Vec<Transaction>> =
                HashMap::with_capacity(dividend_response.data.len());

            for div in dividend_response.data.iter() {
                let symbol_dividends = res
                    .entry(div.symbol.clone())
                    .or_insert(Vec::<Transaction>::new());
                symbol_dividends.push(Transaction {
                    id: "".to_string(),
                    account_id: "".to_string(),
                    symbol: div.symbol.clone(),
                    date: div.date.clone(),
                    transaction_type: TransactionType::Dividend,
                    quantity: 0,
                    pps: div.dividend,
                });
            }

            return Some(res);
        }

        None
    }
}

#[cfg(test)]
impl MarketDataFetcher for MarketDataClient {
    async fn fetch_prices(symbols: &[String]) -> Option<HashMap<String, SymbolPrice>> {
        None
    }

    async fn fetch_dividends(symbols: &[String]) -> Option<HashMap<String, Vec<Transaction>>> {
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

    let existing_size = prices.len();
    let fetched_prices = MarketDataClient::fetch_prices(&missing_symbols).await;
    // println!(">>>> {:?}", fetched_prices);
    if let Some(price_map) = fetched_prices {
        prices.extend(price_map);
    }

    if prices.len() > existing_size {
        let s: String = serde_json::to_string(&prices).unwrap();
        cache.set("prices", s, 60 * 60 * 12).await;
    }

    Some(prices)
}

pub async fn load_dividends<C: Cache + Send + Sync>(
    cache: &C,
    symbols: &[&str],
) -> Option<HashMap<String, Vec<Transaction>>> {
    let mut dividends: HashMap<String, Vec<Transaction>> = HashMap::with_capacity(symbols.len());
    let cached_dividends = cache.get("dividends").await;
    let mut missing_symbols: HashSet<String> = symbols.iter().map(|s| s.to_string()).collect();

    if let Some(cached_dividends) = cached_dividends {
        for (symbol, transactions) in cached_dividends.as_object().unwrap() {
            missing_symbols.remove(symbol);
            let div_list = transactions
                .as_array()
                .unwrap()
                .iter()
                .map(|tr| Transaction::from(tr))
                .collect();
            dividends.insert(symbol.clone(), div_list);
        }

        if missing_symbols.is_empty() {
            return Some(dividends);
        }
    }

    if missing_symbols.is_empty() {
        return Some(dividends);
    }

    let missing_symbols: Vec<String> = missing_symbols.into_iter().collect();
    println!(
        "[fetch_dividends] >>>> Going to network with: {:?}",
        missing_symbols
    );
    let market_dividend = MarketDataClient::fetch_dividends(&missing_symbols).await;
    println!("[fetch_dividends] market_dividend: {:?}", market_dividend);

    if let Some(market_dividend) = market_dividend {
        dividends.extend(market_dividend);
    }

    for lookup_symbol in missing_symbols {
        if !dividends.contains_key(&lookup_symbol) {
            dividends.insert(lookup_symbol.clone(), Vec::with_capacity(0));
        }
    }

    let s: String = serde_json::to_string(&dividends).unwrap();
    cache.set("dividends", s, 60 * 60 * 24 * 3).await;

    Some(dividends)
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
