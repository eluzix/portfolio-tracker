use std::collections::HashMap;
use tracker_types::transactions::Transaction;

pub fn transactions_by_account<'a>(transactions: &'a [Transaction]) -> HashMap<&'a str, Vec<&'a Transaction>> {
    let mut map: HashMap<&'a str, Vec<&'a Transaction>> = HashMap::new();

    for transaction in transactions {
        map.entry(&transaction.account_id)
            .or_insert_with(Vec::new)
            .push(transaction);
    }

    map
}
