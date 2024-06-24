use std::collections::HashMap;
use aws_sdk_dynamodb::types::AttributeValue;
use chrono::{NaiveDate};

#[derive(Debug, PartialEq, Clone, Copy)]
pub enum TransactionType {
    Buy,
    Sell,
    Dividend,
}

impl Default for TransactionType {
    fn default() -> Self {
        TransactionType::Buy
    }
}

#[derive(Debug, PartialEq)]
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
        let quantity = item.get("quantity").unwrap().as_n().unwrap().parse().unwrap();
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
