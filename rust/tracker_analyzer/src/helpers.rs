use std::collections::{HashMap, HashSet};

use crate::types::transactions::Transaction;

pub fn to_transactions_slice(transactions: &[Transaction]) -> Vec<&Transaction> {
    let mut result: Vec<&Transaction> = Vec::with_capacity(transactions.len());
    for transaction in transactions.iter() {
        result.push(&transaction);
    }
    result
}

pub fn transactions_by_account<'a>(transactions: &'a [Transaction]) -> HashMap<&'a str, Vec<&'a Transaction>> {
    let mut map: HashMap<&'a str, Vec<&'a Transaction>> = HashMap::new();

    for transaction in transactions {
        map.entry(&transaction.account_id)
            .or_insert_with(Vec::new)
            .push(transaction);
    }

    map
}

pub fn extract_symbols(transactions: &Vec<&Transaction>) -> HashSet<String> {
    transactions.iter().map(|t| t.symbol.clone()).collect::<HashSet<_>>()
}

pub fn sort_transactions_by_date(transactions: &mut Vec<Transaction>) {
    transactions.sort_by(|a, b| a.date.cmp(&b.date));
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_transactions_by_account() {
        let transactions = vec![
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "2".to_string(),
                symbol: "AAPL".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                ..Default::default()
            },
        ];

        let result = transactions_by_account(&transactions);

        assert_eq!(result.len(), 2);
        assert_eq!(result.get("1").unwrap().len(), 2);
        assert_eq!(result.get("2").unwrap().len(), 1);
    }

    #[test]
    fn test_extract_symbols() {
        let transactions = vec![
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "GOOGL".to_string(),
                ..Default::default()
            },
        ];

        let result = extract_symbols(&to_transactions_slice(transactions.as_slice()));

        assert_eq!(result.len(), 2);
        assert!(result.contains("AAPL"));
        assert!(result.contains("GOOGL"));
    }

    #[test]
    fn test_sort_transactions_by_date() {
        let mut transactions = vec![
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                date: "2021-01-01".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                date: "2021-01-03".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                date: "2021-01-02".to_string(),
                ..Default::default()
            },
        ];

        sort_transactions_by_date(&mut transactions);

        assert_eq!(transactions[0].date, "2021-01-01");
        assert_eq!(transactions[1].date, "2021-01-02");
        assert_eq!(transactions[2].date, "2021-01-03");
    }
}
