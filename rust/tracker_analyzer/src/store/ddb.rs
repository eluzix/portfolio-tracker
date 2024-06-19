use aws_config::BehaviorVersion;
use aws_config::meta::region::RegionProviderChain;
use aws_sdk_dynamodb::{Client, Error};

pub async fn get_client() -> Result<Client, Error> {
    let region_provider = RegionProviderChain::default_provider().or_else("us-east-1");
    let config = aws_config::defaults(BehaviorVersion::latest())
        .region(region_provider);

    #[cfg(feature = "local_creds")]
    let config = config.profile_name("tracker");

    let final_config = config.load().await;
    Ok(Client::new(&final_config))
}
