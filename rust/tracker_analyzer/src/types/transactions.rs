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
    Split,
}

impl From<&str> for TransactionType {
    fn from(value: &str) -> Self {
        match value {
            "buy" => TransactionType::Buy,
            "sell" => TransactionType::Sell,
            "dividend" => TransactionType::Dividend,
            "split" => TransactionType::Split,
            _ => panic!("Invalid transaction type"),
        }
    }
}

impl From<String> for TransactionType {
    fn from(value: String) -> Self {
        value.as_str().into()
    }
}

impl From<&TransactionType> for String {
    fn from(value: &TransactionType) -> Self {
        match value {
            TransactionType::Buy => "buy".to_string(),
            TransactionType::Sell => "sell".to_string(),
            TransactionType::Dividend => "dividend".to_string(),
            TransactionType::Split => "split".to_string(),
        }
    }
}

#[derive(Debug, PartialEq, Clone, Deserialize, Serialize)]
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

    pub fn generate_id(&self) -> String {
        format!(
            "{}#{}#{}#{:?}#{}#{}",
            self.account_id, self.symbol, self.date, self.transaction_type, self.quantity, self.pps
        )
    }
}

impl From<&Value> for Transaction {
    fn from(value: &Value) -> Self {
        match Transaction::deserialize(value) {
            Ok(t) => t,
            Err(_) => {
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

impl Into<HashMap<String, AttributeValue>> for Transaction {
    fn into(self) -> HashMap<String, AttributeValue> {
        let mut item: HashMap<String, AttributeValue> = HashMap::with_capacity(9);
        item.insert("id".to_string(), AttributeValue::S(self.id));
        item.insert("account_id".to_string(), AttributeValue::S(self.account_id));
        item.insert("symbol".to_string(), AttributeValue::S(self.symbol));
        item.insert("date".to_string(), AttributeValue::S(self.date));
        item.insert(
            "type".to_string(),
            AttributeValue::S(String::from(&self.transaction_type)),
        );
        item.insert(
            "quantity".to_string(),
            AttributeValue::N(self.quantity.to_string()),
        );
        item.insert("pps".to_string(), AttributeValue::N(self.pps.to_string()));

        item
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
