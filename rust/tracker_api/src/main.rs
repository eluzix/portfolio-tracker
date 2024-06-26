use lambda_http::{Body, Error, Request, Response, run, service_fn, tracing};

use tracker_analyzer::helpers::analyze_user_portfolio;

async fn function_handler(_event: Request) -> Result<Response<Body>, Error> {
    let portfolio = analyze_user_portfolio("1").await.unwrap();
    let js = serde_json::to_value(&portfolio).unwrap();

    let resp = Response::builder()
        .status(200)
        .header("content-type", "application/json")
        .body(Body::from(serde_json::to_vec(&js).unwrap()))
        .map_err(Box::new)?;

    Ok(resp)
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    tracing::init_default_subscriber();

    run(service_fn(function_handler)).await
}
