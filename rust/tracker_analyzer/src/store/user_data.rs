use std::collections::HashMap;

use crate::helpers::sort_transactions_by_date;
use crate::store::ddb;
use crate::types::account::AccountMetadata;
use crate::types::transactions::Transaction;
use aws_sdk_dynamodb::types::AttributeValue;
use aws_sdk_dynamodb::Error;
use openssl::sha::{sha256, Sha256};

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

    let mut transactions: Vec<Transaction> = Vec::with_capacity(350);
    let mut accounts: Vec<AccountMetadata> = Vec::with_capacity(10);

    if let Some(items) = results.items {
        // let items2 = items.iter().map(|v| v.into()).collect();
        for item in items {
            // println!("ITEM: {:?}", item);

            match item
                .get("SK")
                .unwrap()
                .as_s()
                .unwrap()
                .as_str()
                .split('#')
                .next()
                .unwrap()
            {
                "account" => {
                    let account = AccountMetadata::from_dynamodb(item);
                    accounts.push(account);
                }
                "transaction" => {
                    let transaction = Transaction::from_dynamodb(item);
                    transactions.push(transaction);
                }
                _ => (),
            }
        }
    }

    sort_transactions_by_date(&mut transactions);
    Ok((transactions, accounts))
}

pub async fn add_transaction(uid: &str, tr: &Transaction) -> Result<String, Error> {
    let hash = sha256(tr.generate_id().as_bytes());
    let tr_id = hex::encode(hash);

    let mut item: HashMap<String, AttributeValue> = tr.clone().into();
    item.insert("id".to_string(), AttributeValue::S(tr_id.clone()));
    item.insert("PK".to_string(), AttributeValue::S(format!("user#{}", uid)));
    item.insert(
        "SK".to_string(),
        AttributeValue::S(format!("transaction#{}", &tr_id)),
    );

    let client = ddb::get_client().await?;

    let res = client
        .put_item()
        .table_name("tracker-data")
        .set_item(Some(item))
        .send()
        .await?;

    println!("[add_transaction] result: {:?}", res);

    Ok(tr_id)
}
