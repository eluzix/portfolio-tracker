use std::collections::HashMap;

use chrono::Utc;
use rayon::prelude::*;

use crate::helpers::extract_symbols;
use crate::store::market::SymbolPrice;
use crate::types::portfolio::AnalyzedPortfolio;
use crate::types::transactions::{Transaction, TransactionType};

pub fn analyze_transactions(
    transactions: &Vec<Transaction>,
    price_table: &HashMap<String, SymbolPrice>,
) -> Option<AnalyzedPortfolio> {
    let mut portfolio = AnalyzedPortfolio::new();
    if transactions.is_empty() {
        return Some(portfolio);
    }

    portfolio.first_transaction = transactions.first().unwrap().naive_date().to_string();
    portfolio.last_transaction = transactions.last().unwrap().naive_date().to_string();

    let all_symbols_set = extract_symbols(transactions);
    let today = Utc::now().date_naive();
    let days_since_inception = today
        .signed_duration_since(transactions.first().unwrap().naive_date())
        .num_days();

    let mut total_invested: Vec<f64> = Vec::with_capacity(transactions.len());
    let mut total_withdrawn: Vec<f64> = Vec::with_capacity(transactions.len());
    let mut total_dividends: Vec<f64> = Vec::with_capacity(transactions.len());
    let mut weighted_cash_flows: Vec<f64> = Vec::with_capacity(transactions.len());
    let mut symbols_values: HashMap<&String, f64> = HashMap::with_capacity(all_symbols_set.len());
    let mut symbols_count: HashMap<&String, f64> = HashMap::with_capacity(all_symbols_set.len());

    transactions.iter().for_each(|transaction| {
        let symbol = &transaction.symbol;
        if let Some(price) = price_table.get(symbol) {
            let transaction_date = transaction.naive_date();
            let days_since_transaction = today.signed_duration_since(transaction_date).num_days();
            let transaction_cash_flow = (transaction.quantity as f64 * transaction.pps)
                * days_since_transaction as f64
                / days_since_inception as f64;
            // weighted_cash_flows = (transaction.quantity as f64 * price.adj_close) * days_since_transaction as f64 / days_since_inception as f64;

            match transaction.transaction_type {
                TransactionType::Buy => {
                    total_invested.push(transaction.quantity as f64 * transaction.pps);
                    weighted_cash_flows.push(transaction_cash_flow);

                    let current_value = symbols_values.get(symbol).unwrap_or(&0.0)
                        + transaction.quantity as f64 * price.adj_close;
                    symbols_values.insert(symbol, current_value);

                    let count =
                        symbols_count.get(&symbol).unwrap_or(&0.0) + transaction.quantity as f64;
                    symbols_count.insert(symbol, count);
                }
                TransactionType::Sell => {
                    total_withdrawn.push(transaction.quantity as f64 * transaction.pps);
                    weighted_cash_flows.push(-transaction_cash_flow);

                    let current_value = symbols_values.get(symbol).unwrap_or(&0.0)
                        - transaction.quantity as f64 * price.adj_close;
                    symbols_values.insert(symbol, current_value);

                    let count =
                        symbols_count.get(&symbol).unwrap_or(&0.0) - transaction.quantity as f64;
                    symbols_count.insert(symbol, count);
                }
                TransactionType::Dividend => {
                    let count = symbols_count.get(&symbol).unwrap_or(&0.0);
                    let tr_value = transaction.pps * count;
                    total_dividends.push(tr_value);

                    let dividend_cash_flow =
                        tr_value * days_since_transaction as f64 / days_since_inception as f64;
                    weighted_cash_flows.push(-dividend_cash_flow);
                }
            }
        } else {
            println!("No price found for symbol: {}", symbol);
        }
    });

    portfolio.current_portfolio_value = symbols_values.values().sum();
    portfolio.total_invested = total_invested.iter().sum();
    portfolio.total_withdrawn = total_withdrawn.iter().sum();
    portfolio.total_dividends = total_dividends.iter().sum();

    // todo: dividends

    let weighted_cash_flows: f64 = weighted_cash_flows.iter().sum();
    let portfolio_gain_value =
        (portfolio.current_portfolio_value + portfolio.total_withdrawn + portfolio.total_dividends)
            - portfolio.total_invested;
    portfolio.portfolio_gain_value = portfolio_gain_value;

    portfolio.modified_dietz_yield =
        portfolio_gain_value / (portfolio.total_invested + weighted_cash_flows);

    portfolio.portfolio_gain = match portfolio.total_invested {
        0.0 => 0.0,
        _ => portfolio_gain_value / portfolio.total_invested,
    };

    let years_since_start = days_since_inception as f64 / 365.0;
    portfolio.annualized_yield =
        ((1.0 + portfolio.portfolio_gain).powf(1.0 / years_since_start)) - 1.0;

    // println!("portfolio.portfolio_gain: {:?}", portfolio.portfolio_gain);
    portfolio.symbols = all_symbols_set;
    Some(portfolio)
}

#[cfg(test)]
mod tests {
    use chrono::NaiveDate;
    use rand::Rng;

    use crate::helpers::to_transactions_slice;

    use super::*;

    fn default_price_table() -> HashMap<String, SymbolPrice> {
        let mut price_table = HashMap::new();
        price_table.insert(
            "AAPL".to_string(),
            SymbolPrice {
                symbol: "AAPL".to_string(),
                adj_close: 100.0,
            },
        );
        price_table
    }

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

        let portfolio = analyze_transactions(&transactions, &HashMap::with_capacity(0)).unwrap();

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

        let portfolio = analyze_transactions(&transactions, &default_price_table()).unwrap();

        assert_eq!(portfolio.total_invested, 600.0);
        assert_eq!(portfolio.total_withdrawn, 400.0);
    }

    #[test]
    fn test_symbols_value() {
        let mut rng = rand::thread_rng();

        let transactions = vec![
            Transaction {
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: rng.gen_range(5..100),
                date: "2023-01-01".to_string(),
                ..Default::default()
            },
            Transaction {
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: rng.gen_range(5..100),
                date: "2023-02-01".to_string(),
                ..Default::default()
            },
            Transaction {
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: rng.gen_range(5..100),
                date: "2023-03-01".to_string(),
                ..Default::default()
            },
            Transaction {
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Sell,
                quantity: rng.gen_range(5..10),
                date: "2023-04-01".to_string(),
                ..Default::default()
            },
        ];

        let price: f64 = rng.gen_range(100.0..200.0);
        let all_transactions_value: f64 = transactions
            .iter()
            .map(|t| {
                if t.transaction_type == TransactionType::Sell {
                    return 0.0 - t.quantity as f64 * price;
                }

                t.quantity as f64 * price
            })
            .sum();

        let mut price_table = HashMap::new();
        price_table.insert(
            "AAPL".to_string(),
            SymbolPrice {
                symbol: "AAPL".to_string(),
                adj_close: price,
            },
        );

        let portfolio = analyze_transactions(&transactions, &price_table).unwrap();

        assert_eq!(
            format!("{:.5}", portfolio.current_portfolio_value),
            format!("{:.5}", all_transactions_value)
        );
    }

    #[test]
    fn test_yields() {
        let mut rng = rand::thread_rng();
        let today = Utc::now().date_naive();
        let price: f64 = rng.gen_range(100.0..200.0);

        let num_of_transactions = 4;
        let quantities: Vec<u32> = (0..num_of_transactions)
            .map(|_| rng.gen_range(5..100))
            .collect();
        let transaction_types = vec![
            TransactionType::Buy,
            TransactionType::Buy,
            TransactionType::Buy,
            TransactionType::Sell,
        ];
        let mut dates: Vec<NaiveDate> = (0..num_of_transactions)
            .map(|i| today - chrono::Duration::days((i + 1) * 365))
            .collect();
        dates.reverse();

        let first_transaction = dates.first().unwrap().clone();

        let transactions: Vec<Transaction> = (0..num_of_transactions)
            .map(|i| Transaction {
                symbol: "AAPL".to_string(),
                transaction_type: transaction_types[i as usize].clone(),
                quantity: quantities[i as usize],
                pps: 1.0,
                date: dates[i as usize].format("%Y-%m-%d").to_string(),
                ..Default::default()
            })
            .collect();

        let days_since_inception = today.signed_duration_since(first_transaction).num_days();
        let years_since_inception = days_since_inception as f64 / 365.0;
        let current_portfolio_value: f64 = transactions
            .iter()
            .map(|t| {
                if t.transaction_type == TransactionType::Sell {
                    return 0.0 - t.quantity as f64 * price;
                }

                t.quantity as f64 * price
            })
            .sum();
        let total_withdrawn: f64 = transactions
            .iter()
            .map(|t| {
                if t.transaction_type == TransactionType::Sell {
                    return t.quantity as f64 * t.pps;
                }
                0.0
            })
            .sum();
        let total_invested: f64 = transactions
            .iter()
            .map(|t| {
                if t.transaction_type == TransactionType::Buy {
                    return t.quantity as f64 * t.pps;
                }
                0.0
            })
            .sum();
        let total_dividends = 0.0;
        let portfolio_gain_value: f64 =
            (current_portfolio_value + total_withdrawn + total_dividends) - total_invested;
        let portfolio_gain = portfolio_gain_value / total_invested;
        let annualized_yield = ((1.0 + portfolio_gain).powf(1.0 / years_since_inception)) - 1.0;
        let weighted_cash_flows: f64 = transactions
            .iter()
            .map(|t| {
                let transaction_date = t.naive_date();
                let days_since_transaction =
                    today.signed_duration_since(transaction_date).num_days();
                match t.transaction_type {
                    TransactionType::Buy => {
                        return t.quantity as f64 * t.pps * days_since_transaction as f64
                            / days_since_inception as f64;
                    }
                    TransactionType::Sell => {
                        return 0.0
                            - t.quantity as f64 * t.pps * days_since_transaction as f64
                                / days_since_inception as f64;
                    }
                    TransactionType::Dividend => 0.0,
                }
            })
            .sum();
        let modified_dietz_yield: f64 =
            portfolio_gain_value / (total_invested + weighted_cash_flows);

        let mut price_table = HashMap::new();
        price_table.insert(
            "AAPL".to_string(),
            SymbolPrice {
                symbol: "AAPL".to_string(),
                adj_close: price,
            },
        );

        let portfolio = analyze_transactions(&transactions, &price_table).unwrap();

        assert_eq!(
            format!("{:.5}", portfolio.annualized_yield),
            format!("{:.5}", annualized_yield)
        );
        assert_eq!(
            format!("{:.5}", portfolio.modified_dietz_yield),
            format!("{:.5}", modified_dietz_yield)
        );
    }

    #[test]
    fn test_dividends() {
        let transactions = vec![
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: 2,
                pps: 100.0,
                date: "2024-01-01".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Buy,
                quantity: 3,
                pps: 100.0,
                date: "2024-02-01".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Dividend,
                quantity: 0,
                pps: 10.0,
                date: "2024-02-15".to_string(),
                ..Default::default()
            },
            Transaction {
                account_id: "1".to_string(),
                symbol: "AAPL".to_string(),
                transaction_type: TransactionType::Sell,
                quantity: 4,
                pps: 100.0,
                date: "2024-05-01".to_string(),
                ..Default::default()
            },
        ];

        let portfolio = analyze_transactions(&transactions, &default_price_table()).unwrap();

        assert_eq!(portfolio.total_invested, 500.0);
        assert_eq!(portfolio.total_withdrawn, 400.0);
        assert_eq!(portfolio.total_dividends, 50.0);
    }
}
