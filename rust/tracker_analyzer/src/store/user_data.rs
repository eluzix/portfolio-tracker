use aws_sdk_dynamodb::Error;
use aws_sdk_dynamodb::types::AttributeValue;
use crate::helpers::sort_transactions_by_date;
use crate::store::ddb;
use crate::types::account::AccountMetadata;
use crate::types::transactions::Transaction;

/// Load user data from the DynamoDB table
/// and sort the transactions by date
pub async fn load_user_data(uid: &str) -> Result<(Vec<Transaction>, Vec<AccountMetadata>), Error> {
    let client = ddb::get_client().await?;
    let results = client
        .query()
        .table_name("tracker-data")
        // .key_condition_expression("#pk = :pk AND begins_with(#sk, :sk)")
        .key_condition_expression("#pk = :pk")
        .expression_attribute_names("#pk", "PK")
        // .expression_attribute_names("#sk", "SK")
        .expression_attribute_values(":pk", AttributeValue::S(format!("user#{uid}")))
        // .expression_attribute_values(":sk", AttributeValue::S("transaction#".to_string()))
        .send()
        .await?;

    let mut transactions: Vec<Transaction> = Vec::with_capacity(250);
    let mut accounts: Vec<AccountMetadata> = Vec::with_capacity(10);

    if let Some(items) = results.items {
        // let items2 = items.iter().map(|v| v.into()).collect();
        for item in items {
            // println!("ITEM: {:?}", item);

            match item.get("SK").unwrap().as_s().unwrap().as_str().split("#").next().unwrap() {
                "account" => {
                    let account = AccountMetadata::from_dynamodb(item);
                    accounts.push(account);
                },
                "transaction" => {
                    let transaction = Transaction::from_dynamodb(item);
                    transactions.push(transaction);
                },
                _ => (),
            }
        }
    }

    sort_transactions_by_date(&mut transactions);
    Ok((transactions, accounts))
}
