use std::collections::HashMap;

use numfmt::{Formatter, Precision};
use tera::{to_value, Tera, Value};

pub fn currency_filter(value: &Value, args: &HashMap<String, Value>) -> tera::Result<Value> {
    let currency = match args.get("sign") {
        Some(currency) => currency.as_str().unwrap(),
        _ => "$",
    };

    let mut f = Formatter::new() // start with blank representation
        .separator(',')
        .unwrap()
        .prefix(currency)
        .unwrap()
        .precision(Precision::Decimals(0));
    if let Some(val) = value.as_f64() {
        return Ok(to_value(f.fmt2(val)).unwrap());
    }

    Ok(to_value(value.to_string()).unwrap())
}

pub fn percent_filter(value: &Value, _: &HashMap<String, Value>) -> tera::Result<Value> {
    let mut f = Formatter::new() // start with blank representation
        .separator(',')
        .unwrap()
        .suffix("%")
        .unwrap()
        .precision(Precision::Decimals(2));
    if let Some(val) = value.as_f64() {
        return Ok(to_value(f.fmt2(val * 100.0)).unwrap());
    }

    Ok(to_value(value.to_string()).unwrap())
}

pub fn load_tera() -> Tera {
    let mut tera = match Tera::new("templates/**/*.html") {
        Ok(t) => t,
        Err(e) => {
            println!("Parsing error(s): {}", e);
            panic!("EEEE");
        }
    };
    tera.register_filter("currency_filter", currency_filter);
    tera.register_filter("percent_filter", percent_filter);

    tera
}
