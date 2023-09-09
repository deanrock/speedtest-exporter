use std::io::{Error, ErrorKind};

use cfspeedtest::{
    measurements::Measurement,
    speedtest::{run_latency_test, run_tests, test_download, test_upload, PayloadSize, TestType},
    OutputFormat, SpeedTestCLIOptions,
};

use crate::SpeedTestResults;

fn measurement_to_bits(m: Vec<Measurement>) -> Result<f64, Error> {
    if let Some(measurement) = m.last() {
        return Ok(measurement.mbit * 1024.0 * 1024.0);
    } else {
        return Err(Error::new(
            ErrorKind::Other,
            format!("measurement not available in Vec"),
        ));
    }
}

pub fn run_cloudflare_speedtest() -> Result<SpeedTestResults, Error> {
    let options = SpeedTestCLIOptions {
        output_format: OutputFormat::None, // don't write to stdout
        ipv4: false,                       // don't force ipv4 usage
        ipv6: false,                       // don't force ipv6 usage
        verbose: false,
        nr_tests: 5,
        nr_latency_tests: 20,
        max_payload_size: PayloadSize::M10,
    };

    let client = reqwest::blocking::Client::new();
    let latency = run_latency_test(&client, options.nr_latency_tests, options.output_format).1;

    let payload_sizes = PayloadSize::sizes_from_max(options.max_payload_size);
    let download_test = run_tests(
        &client,
        test_download,
        TestType::Download,
        payload_sizes.clone(),
        options.nr_tests,
        options.output_format,
    );
    let upload_test = run_tests(
        &client,
        test_upload,
        TestType::Upload,
        payload_sizes.clone(),
        options.nr_tests,
        options.output_format,
    );

    return Ok(SpeedTestResults {
        download_bits: measurement_to_bits(download_test)?,
        upload_bits: measurement_to_bits(upload_test)?,
        latency,
        jitter: None,
    });
}
