use std::collections::HashMap;
use chrono::{NaiveDate, Utc};

use rayon::prelude::*;

use crate::helpers::extract_symbols;
use crate::store::market::SymbolPrice;
use crate::types::portfolio::AnalyzedPortfolio;
use crate::types::transactions::{Transaction, TransactionType};

pub fn analyze_transactions(transactions: &Vec<&Transaction>, price_table: &HashMap<String, SymbolPrice>) -> Option<AnalyzedPortfolio> {
    let mut portfolio = AnalyzedPortfolio::new();
    let all_symbols_set = extract_symbols(transactions);
    let today = Utc::now().date_naive();
    let days_since_inception = today.signed_duration_since(transactions.first().unwrap().naive_date()).num_days();


    let (total_invested, total_withdrawn, weighted_cash_flows, symbols_value) = transactions.par_iter().fold(
        || (0.0, 0.0, 0.0, HashMap::new()),
        |(mut invested, mut withdrawn, mut weighted_cash_flows, mut symbols_value), transaction| {
            let symbol = &transaction.symbol;
            let price = price_table.get(symbol).unwrap();
            let transaction_date= transaction.naive_date();
            let days_since_transaction = today.signed_duration_since(transaction_date).num_days();
            weighted_cash_flows = (transaction.quantity as f64 * price.adj_close) * days_since_transaction as f64 / days_since_inception as f64;

            match transaction.transaction_type {
                TransactionType::Buy => {
                    invested += transaction.quantity as f64 * transaction.pps;

                    let current_value = symbols_value.get(symbol).unwrap_or(&0.0) + transaction.quantity as f64 * price.adj_close;
                    symbols_value.insert(symbol, current_value);
                }
                TransactionType::Sell => {
                    withdrawn += transaction.quantity as f64 * transaction.pps;
                    weighted_cash_flows = -weighted_cash_flows;

                    let current_value = symbols_value.get(symbol).unwrap_or(&0.0) - transaction.quantity as f64 * price.adj_close;
                    symbols_value.insert(symbol, current_value);
                }
                TransactionType::Dividend => {}
            }
            (invested, withdrawn, weighted_cash_flows, symbols_value)
        },
    ).reduce(
        || (0.0, 0.0, 0.0, HashMap::new()),
        |(invested1, withdrawn1, weighted_cash_flows1, symbols_value1), (invested2, withdrawn2, weighted_cash_flows2, symbols_value2)| {
            let mut combined_symbols_value = symbols_value1;
            for (symbol, value) in symbols_value2 {
                *combined_symbols_value.entry(symbol).or_insert(0.0) += value;
            }

            (invested1 + invested2, withdrawn1 + withdrawn2, weighted_cash_flows1 + weighted_cash_flows2, combined_symbols_value)
        },
    );

    portfolio.current_portfolio_value = symbols_value.values().sum();
    portfolio.total_invested = total_invested;
    portfolio.total_withdrawn = total_withdrawn;

    // todo: dividends
    let total_dividends = 0.0;
    let portfolio_gain = (portfolio.current_portfolio_value + total_withdrawn + total_dividends) - total_invested;
    portfolio.modified_dietz_yield = portfolio_gain / (total_invested + weighted_cash_flows);

    portfolio.portfolio_gain = match total_invested {
        0.0 => 0.0,
        _ => portfolio_gain / total_invested,
    };

    let years_since_start = days_since_inception as f64 / 365.0;
    portfolio.annualized_yield = ((1.0 + portfolio.portfolio_gain).powf(1.0 / years_since_start)) - 1.0;

    // println!("portfolio.portfolio_gain: {:?}", portfolio.portfolio_gain);
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
