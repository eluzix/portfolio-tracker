use config::Config;
use once_cell::sync::Lazy;

static CONFIG: Lazy<Config> = Lazy::new(|| {
    Config::builder()
        .add_source(config::File::with_name("config.json"))
        .build()
        .unwrap()
});


pub fn get<'a, T: serde::Deserialize<'a>>(key: &str) -> Option<T> {
    CONFIG.get::<T>(key).ok()
}

pub fn get_str(key: &str) -> Option<String> {
    CONFIG.get_string(key).ok()
}