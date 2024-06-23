use std::collections::HashMap;
use aws_sdk_dynamodb::Error;
use tracker_analyzer::helpers::transactions_by_account;
use tracker_analyzer::portfolio_analyzer::analyze_transactions;
use tracker_analyzer::store::{cache, market};
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
    for (account_id, transactions) in account_transactions {
        println!("--------\nAccount: {}", account_id);
        let portfolio_data  = analyze_transactions(&transactions, &prices).unwrap();
        println!("Portfolio: {:?}", portfolio_data);
    }

    Ok(())
}
