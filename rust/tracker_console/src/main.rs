use std::collections::HashMap;
use aws_sdk_dynamodb::Error;
use tracker_analyzer::helpers::transactions_by_account;
use tracker_analyzer::portfolio_analyzer::analyze_transactions;
use tracker_analyzer::store::cache;
use tracker_analyzer::store::user_data::load_user_data;

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), Error> {

    // let transactions = load_user_data("1").await?;
    // println!("Found {} transactions", transactions.len());

    let prices = cache::get("prices").await.unwrap();
    println!("Prices: {:?}", prices);

    // let account_transactions = transactions_by_account(&transactions);
    // for (account_id, transactions) in account_transactions {
    //     println!("--------\nAccount: {}", account_id);
    //     analyze_transactions(&transactions, &HashMap::with_capacity(0)).unwrap();
    // }

    Ok(())
}
