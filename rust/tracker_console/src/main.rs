use aws_sdk_dynamodb::Error;
use rayon::prelude::*;

use tracker_analyzer::helpers::{to_transactions_slice, transactions_by_account};
use tracker_analyzer::portfolio_analyzer::analyze_transactions;
use tracker_analyzer::store::market;
use tracker_analyzer::store::user_data::load_user_data;

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), Error> {
    let prices = market::load_prices().await.unwrap();
    // println!("Prices: {:?}", prices.get("AAPL"));

    // todo: Load account metadata
    let transactions = load_user_data("1").await?;
    println!("Found {} transactions", transactions.len());

    let account_transactions = transactions_by_account(&transactions);
    let results = account_transactions.par_iter().map(|(account_id, transactions)| {
        let portfolio_data = analyze_transactions(&transactions, &prices).unwrap();
        (account_id, portfolio_data)
    }).collect::<Vec<_>>();

    for (account_id, portfolio_data) in results {
        println!("--------\nAccount: {}", account_id);
        println!("Portfolio: {:?}", portfolio_data);
    }

    let all_portfolio_data = analyze_transactions(&to_transactions_slice(&transactions), &prices).unwrap();
    println!("--------\nAll Accounts");
    println!("Portfolio: {:?}", all_portfolio_data);

    Ok(())
}
