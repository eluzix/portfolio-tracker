use std::collections::HashMap;
use rayon::prelude::*;

use crate::helpers::extract_symbols;
use crate::types::portfolio::AnalyzedPortfolio;
use crate::types::transactions::{Transaction, TransactionType};

pub fn analyze_transactions(transactions: &Vec<&Transaction>, price_table: &HashMap<String, f64>) -> Option<AnalyzedPortfolio> {
    let mut portfolio = AnalyzedPortfolio::new();
    let all_symbols_set = extract_symbols(transactions);

    let (total_invested, total_withdrawn) = transactions.par_iter().fold(
        || (0.0, 0.0),
        |(mut invested, mut withdrawn), transaction| {
            match transaction.transaction_type {
                TransactionType::Buy => {
                    invested += transaction.quantity as f64 * transaction.pps;
                }
                TransactionType::Sell => {
                    withdrawn += transaction.quantity as f64 * transaction.pps;
                }
                TransactionType::Dividend => {}
            }
            (invested, withdrawn)
        }
    ).reduce(
        || (0.0, 0.0),
        |(invested1, withdrawn1), (invested2, withdrawn2)| {
            (invested1 + invested2, withdrawn1 + withdrawn2)
        }
    );

    portfolio.total_invested = total_invested;
    portfolio.total_withdrawn = total_withdrawn;

    println!("Symbols: {:?}", all_symbols_set);

    portfolio.symbols = all_symbols_set;
    Some(portfolio)
}

#[cfg(test)]
mod tests {
    use crate::helpers::to_transactions_slice;

    use super::*;

    #[test]
    fn empty_transactions_empty_analyzer() {
        let price_table = HashMap::with_capacity(0);
        assert_eq!(
            analyze_transactions(&vec![], &price_table),
            Some(AnalyzedPortfolio::new())
        );
    }

    #[test]
    fn test_symbols_extraction() {
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

        let portfolio = analyze_transactions(&to_transactions_slice(transactions.as_slice()), &HashMap::with_capacity(0)).unwrap();


        assert_eq!(portfolio.symbols.len(), 1);
        assert!(portfolio.symbols.contains("AAPL"));
    }

    #[test]
    fn test_totals() {
        let transactions = vec![
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: 2,
                pps: 100.0,
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: 3,
                pps: 100.0,
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: 1,
                pps: 100.0,
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Sell,
                quantity: 4,
                pps: 100.0,
                ..Default::default()
            },
        ];

        let portfolio = analyze_transactions(&to_transactions_slice(transactions.as_slice()), &HashMap::with_capacity(0)).unwrap();

        assert_eq!(portfolio.total_invested, 600.0);
        assert_eq!(portfolio.total_withdrawn, 400.0);
    }
}
