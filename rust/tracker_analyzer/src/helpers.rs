use std::collections::{HashMap, HashSet};

use rayon::prelude::*;

use crate::portfolio_analyzer::analyze_transactions;
use crate::store::market;
use crate::store::user_data::load_user_data;
use crate::types::account::AccountMetadata;
use crate::types::portfolio::AnalyzedPortfolio;
use crate::types::transactions::Transaction;
use crate::types::user_portfolio::UserPortfolio;

pub fn to_transactions_slice(transactions: &[Transaction]) -> Vec<&Transaction> {
    let mut result: Vec<&Transaction> = Vec::with_capacity(transactions.len());
    for transaction in transactions.iter() {
        result.push(transaction);
    }
    result
}

pub fn transactions_by_account<'a>(transactions: &'a [Transaction]) -> HashMap<&'a str, Vec<&'a Transaction>> {
    let mut map: HashMap<&'a str, Vec<&'a Transaction>> = HashMap::new();

    for transaction in transactions {
        map.entry(&transaction.account_id)
            .or_default()
            .push(transaction);
    }

    map
}

pub fn extract_symbols(transactions: &[&Transaction]) -> HashSet<String> {
    transactions.iter().map(|t| t.symbol.clone()).collect::<HashSet<_>>()
}

pub fn sort_transactions_by_date(transactions: &mut [Transaction]) {
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

pub async fn analyze_user_portfolio(user_id: &str) -> Option<UserPortfolio> {
    let prices = market::load_prices().await?;
    let resp = load_user_data(user_id).await.unwrap();
    let transactions = resp.0;
    let accounts_metadata: HashMap<String, AccountMetadata> = resp.1.into_iter().map(|account| (account.id.clone(), account)).collect();

    let account_transactions = transactions_by_account(&transactions);
    let results = account_transactions.par_iter().map(|(account_id, transactions)| {
        let portfolio_data = analyze_transactions(transactions, &prices).unwrap();
        (account_id, portfolio_data)
    }).collect::<Vec<_>>();

    let mut results_map: HashMap<String, AnalyzedPortfolio> = HashMap::with_capacity(results.len());
    for (account_id, portfolio_data) in results {
        results_map.insert(account_id.to_string(), portfolio_data);
    }

    let all_portfolio_data = analyze_transactions(&to_transactions_slice(&transactions), &prices).unwrap();

    Some(UserPortfolio {
        accounts_metadata,
        accounts: results_map,
        portfolio: all_portfolio_data,
    })
}
