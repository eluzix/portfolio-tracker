use std::borrow::Borrow;
use std::collections::{HashMap, HashSet};

use rayon::prelude::*;

use crate::portfolio_analyzer::analyze_transactions;
use crate::store::cache::default_cache;
use crate::store::market::{self, CurrencyMetadata};
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

pub fn transactions_by_account(transactions: &[Transaction]) -> HashMap<String, Vec<Transaction>> {
    let mut map: HashMap<String, Vec<Transaction>> = HashMap::new();

    for transaction in transactions {
        map.entry(transaction.account_id.clone())
            .or_default()
            .push(transaction.clone());
    }

    map
}

pub fn extract_symbols(transactions: &[Transaction]) -> HashSet<String> {
    transactions
        .iter()
        .map(|t| t.symbol.clone())
        .collect::<HashSet<_>>()
}

pub fn sort_transactions_by_date(transactions: &mut [Transaction]) {
    transactions.sort_by(|a, b| a.date.cmp(&b.date));
}

pub fn merge_dividends(
    transactions: &mut Vec<Transaction>,
    dividends: &HashMap<String, Vec<Transaction>>,
    // dividends: &Option<HashMap<String, Vec<Transaction>>>,
) {
    // if let Some(dividends) = dividends {
    let mut symbols_first_tr: HashMap<String, String> = HashMap::new();
    let symbols = transactions
        .iter()
        .map(|t| {
            if let Some(tr_date) = symbols_first_tr.get(&t.symbol) {
                if tr_date > &t.date {
                    symbols_first_tr.insert(t.symbol.clone(), t.date.clone());
                }
            } else {
                symbols_first_tr.insert(t.symbol.clone(), t.date.clone());
            }

            t.symbol.clone()
        })
        .collect::<HashSet<_>>();

    // let symbols = extract_symbols(transactions);
    for symbol in symbols.iter() {
        let symbol_dividends = dividends.get(symbol);
        if let Some(symbol_dividends) = symbol_dividends {
            if let Some(start_date) = symbols_first_tr.get(&symbol.to_string()) {
                symbol_dividends.iter().for_each(|t| {
                    if &t.date > start_date {
                        transactions.push(t.clone());
                    }
                });
            }
        }
    }
    // }
}

pub async fn analyze_user_portfolio(user_id: &str, currency: &str) -> Option<UserPortfolio> {
    let cache = default_cache();
    let resp = load_user_data(user_id).await.unwrap();
    let accounts_metadata: hashmap<string, accountmetadata> = resp
        .1
        .into_iter()
        .map(|account| (account.id.clone(), account))
        .collect();
    let mut transactions = resp.0;
    // let transactions_slice = to_transactions_slice(&transactions);

    let symbols = extract_symbols(&transactions);
    let all_symbols = symbols.iter().map(|s| s.as_str()).collect::<Vec<_>>();

    let data_resp = tokio::try_join!(
        market::load_prices(&*cache, all_symbols.as_slice()),
        market::load_dividends(&*cache, &all_symbols),
    );

    let (prices, dividends) = data_resp.unwrap();

    let mut rate: f64 = 1.0;
    let mut currency_md: CurrencyMetadata = CurrencyMetadata::default();
    if currency != "USD" {
        let join_res = tokio::try_join!(
            market::load_exhnage_rate(&*cache, currency),
            market::load_currency_metadata(&*cache, currency)
        );

        if let Ok((rates_resp, md_resp)) = join_res {
            rate = rates_resp;
            currency_md = md_resp;
        }
    }

    let mut account_transactions = transactions_by_account(&transactions);
    let results = account_transactions
        .par_iter_mut()
        .map(|(account_id, transactions)| {
            merge_dividends(transactions, &dividends);
            sort_transactions_by_date(transactions);
            let portfolio_data = analyze_transactions(transactions, &prices).unwrap();
            (account_id, portfolio_data)
        })
        .collect::<Vec<_>>();

    let mut results_map: HashMap<String, AnalyzedPortfolio> = HashMap::with_capacity(results.len());
    for (account_id, portfolio_data) in results {
        results_map.insert(account_id.to_string(), portfolio_data);
    }

    merge_dividends(&mut transactions, &dividends);
    sort_transactions_by_date(&mut transactions);
    let all_portfolio_data = analyze_transactions(&transactions, &prices).unwrap();

    Some(UserPortfolio {
        accounts_metadata,
        accounts: results_map,
        portfolio: all_portfolio_data,
        rate,
        currency_md,
    })
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

        let result = extract_symbols(&transactions);

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
