use aws_sdk_dynamodb::types::AttributeValue;
use chrono::NaiveDate;
use core::panic;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;

#[derive(Debug, PartialEq, Clone, Copy, Default, Deserialize, Serialize)]
pub enum TransactionType {
    #[default]
    Buy,
    Sell,
    Dividend,
}

impl From<&str> for TransactionType {
    fn from(value: &str) -> Self {
        match value {
            "buy" => TransactionType::Buy,
            "sell" => TransactionType::Sell,
            "dividend" => TransactionType::Dividend,
            _ => panic!("Invalid transaction type"),
        }
    }
}

#[derive(Debug, PartialEq, Deserialize, Serialize)]
pub struct Transaction {
    pub id: String,
    pub account_id: String,
    pub symbol: String,
    pub date: String,
    pub transaction_type: TransactionType,
    pub quantity: u32,
    pub pps: f64,
}

impl Transaction {
    pub fn from_dynamodb(item: HashMap<String, AttributeValue>) -> Self {
        let id = item.get("id").unwrap().as_s().unwrap().to_string();
        let account_id = item.get("account_id").unwrap().as_s().unwrap().to_string();
        let symbol = item.get("symbol").unwrap().as_s().unwrap().to_string();
        let date = item.get("date").unwrap().as_s().unwrap().to_string();

        let transaction_type = match item.get("type").unwrap().as_s().unwrap().as_str() {
            "buy" => TransactionType::Buy,
            "sell" => TransactionType::Sell,
            "dividend" => TransactionType::Dividend,
            _ => panic!("Invalid transaction type"),
        };
        let quantity = item
            .get("quantity")
            .unwrap()
            .as_n()
            .unwrap()
            .parse()
            .unwrap();
        let pps = item.get("pps").unwrap().as_n().unwrap().parse().unwrap();

        Transaction {
            id,
            account_id,
            symbol,
            date,
            transaction_type,
            quantity,
            pps,
        }
    }

    pub fn naive_date(&self) -> NaiveDate {
        NaiveDate::parse_from_str(&self.date, "%Y-%m-%d").unwrap()
    }
}

impl From<&Value> for Transaction {
    fn from(value: &Value) -> Self {
        match Transaction::deserialize(value) {
            Ok(t) => t,
            Err(_) => {
                let mut t = Transaction::default();

                let id = value.get("id").unwrap();
                let id = match id {
                    Value::Null => "".to_string(),
                    _ => id.to_string(),
                };

                let account_id = value.get("account_id").unwrap();
                let account_id = match account_id {
                    Value::Null => "".to_string(),
                    _ => account_id.to_string(),
                };

                Transaction {
                    id,
                    account_id,
                    symbol: value.get("symbol").unwrap().as_str().unwrap().to_string(),
                    date: value.get("date").unwrap().as_str().unwrap().to_string(),
                    transaction_type: TransactionType::from(
                        value.get("type").unwrap().as_str().unwrap(),
                    ),
                    quantity: value.get("quantity").unwrap().as_u64().unwrap() as u32,
                    pps: value.get("pps").unwrap().as_f64().unwrap(),
                }
            }
        }
    }
}

impl Default for Transaction {
    fn default() -> Self {
        Transaction {
            id: "1".to_string(),
            account_id: "1".to_string(),
            symbol: "AAPL".to_string(),
            date: "2024-06-23".to_string(),
            transaction_type: TransactionType::Buy,
            quantity: 0,
            pps: 0.0,
        }
    }
}

#[cfg(test)]
mod tests {
    use serde_json::json;

    use super::*;

    #[test]
    fn test_transaction_type() {
        assert_eq!(TransactionType::from("dividend"), TransactionType::Dividend);
    }

    #[test]
    fn test_transaction_from_json() {
        let js = serde_json::from_str("{\"account_id\":null,\"date\":\"2024-06-11\",\"id\":null,\"pps\":0.93,\"quantity\":0,\"symbol\":\"HDV\",\"type\":\"dividend\"}").unwrap();
        let t = Transaction::from(&js);
        assert_eq!(t.symbol, "HDV".to_string());
        assert_eq!(t.date, "2024-06-11".to_string());
        assert_eq!(t.pps, 0.93);
    }
}
