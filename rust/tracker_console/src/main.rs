use aws_sdk_dynamodb::Error;
use tracker_analyzer::helpers::transactions_by_account;

use tracker_types::transactions::Transaction;
use tracker_storage::user_data::load_user_data;

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), Error> {

    let results = load_user_data("1").await?;
    let mut transactions: Vec<Transaction> = Vec::with_capacity(results.count as usize);

    if let Some(items) = results.items {
        // let items2 = items.iter().map(|v| v.into()).collect();
        for item in items {
            // println!("ITEM: {:?}", item);
            let transaction = Transaction::from_dynamodb(item);
            transactions.push(transaction);
        }
    }

    println!("Found {} transactions", transactions.len());

    let account_transactions = transactions_by_account(&transactions);
    for (account_id, transactions) in account_transactions {
        println!("--------\nAccount: {}", account_id);
        for transaction in transactions {
            println!("{:?}", transaction);
        }
    }

    Ok(())
}
