mod cloudflare;
mod speedtest_net;

use std::io;

use crate::{cloudflare::run_cloudflare_speedtest, speedtest_net::run_speedtest_net_speedtest};
use prometheus_exporter::{self, prometheus::register_gauge_vec};

#[derive(Debug)]
pub struct SpeedTestResults {
    pub download_bits: f64,
    pub upload_bits: f64,
    pub latency: f64,
    pub jitter: Option<f64>,
}

pub fn exporter(server_host: String, server_port: u16) {
    let binding = format!("{}:{}", server_host, server_port).parse().unwrap();
    let exporter = prometheus_exporter::start(binding).unwrap();

    println!(
        "Listening on http://{}:{}/metrics",
        binding.ip(),
        binding.port(),
    );

    let labels = vec!["name"];
    let speedtest_up = register_gauge_vec!(
        "speedtest_up",
        "Speedtest status whether the scrape worked",
        &labels
    )
    .unwrap();
    let upload_speed = register_gauge_vec!(
        "speedtest_upload_bits_per_second",
        "Speedtest current upload speed in bits/s",
        &labels
    )
    .unwrap();
    let download_speed = register_gauge_vec!(
        "speedtest_download_bits_per_second",
        "Speedtest current download speed in bit/s",
        &labels
    )
    .unwrap();
    let ping = register_gauge_vec!(
        "speedtest_ping_latency_milliseconds",
        "Speedtest current ping in ms",
        &labels
    )
    .unwrap();
    let jitter = register_gauge_vec!(
        "speedtest_jitter_milliseconds",
        "Speedtest current jitter in ms",
        &labels
    )
    .unwrap();

    let report = |name: String, r: Result<SpeedTestResults, io::Error>| {
        let labels = vec![name.as_str()];

        match r {
            Ok(r) => {
                speedtest_up.with_label_values(&labels).set(1.0);
                upload_speed.with_label_values(&labels).set(r.upload_bits);
                download_speed
                    .with_label_values(&labels)
                    .set(r.download_bits);
                ping.with_label_values(&labels).set(r.latency);

                if let Some(val) = r.jitter {
                    jitter.with_label_values(&labels).set(val);
                }
            }
            Err(e) => {
                speedtest_up.with_label_values(&labels).set(0.0);
                println!("reporting {} failed with '{}'", name, e)
            }
        }
    };

    loop {
        let guard = exporter.wait_request();

        report("cloudflare".to_string(), run_cloudflare_speedtest());
        report("speedtest_net".to_string(), run_speedtest_net_speedtest());

        drop(guard);
    }
}
