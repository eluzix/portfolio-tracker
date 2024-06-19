use aws_sdk_dynamodb::operation::query::QueryOutput;
use aws_sdk_dynamodb::types::AttributeValue;
use aws_sdk_dynamodb::{Error};
use tracker_types::transactions::Transaction;
use crate::ddb;


pub async fn load_user_data(uid: &str) -> Result<Vec<Transaction>, Error> {
    let client = ddb::get_client().await?;
    let results = client
        .query()
        .table_name("tracker-data")
        .key_condition_expression("#pk = :pk AND begins_with(#sk, :sk)")
        .expression_attribute_names("#pk", "PK")
        .expression_attribute_names("#sk", "SK")
        .expression_attribute_values(":pk", AttributeValue::S(format!("user#{uid}")))
        .expression_attribute_values(":sk", AttributeValue::S("transaction#".to_string()))
        .send()
        .await?;

    let mut transactions: Vec<Transaction> = Vec::with_capacity(results.count as usize);

    if let Some(items) = results.items {
        // let items2 = items.iter().map(|v| v.into()).collect();
        for item in items {
            // println!("ITEM: {:?}", item);
            let transaction = Transaction::from_dynamodb(item);
            transactions.push(transaction);
        }
    } else {
        return Err(Error::from("No items found"));
    }

    Ok(transactions)
}
