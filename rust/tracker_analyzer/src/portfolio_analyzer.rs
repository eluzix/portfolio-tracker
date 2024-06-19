use crate::types::transactions::Transaction;

pub fn analyze_transactions(transactions: &Vec<&Transaction>) -> Option<String> {

    let all_symbols_set = transactions.iter().map(|t| &t.symbol).collect::<std::collections::HashSet<_>>();

    for transaction in transactions {
        println!("{:?}", transaction);
    }

    println!("Symbols: {:?}", all_symbols_set);

    Some("Analysis complete".to_string())
}

// #[cfg(test)]
// mod tests {
//     use super::*;
//
//     #[test]
//     fn fv_works_when_pmt_at_end_of_period() {
//         assert_eq!(
//             fv(0.1, 5.0, Some(100.0), Some(1000.0), Some(false)),
//             -2221.020000000001
//         );
//     }
// }
