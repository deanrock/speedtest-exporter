use serde::{Deserialize, Serialize};
use std::io::Error;
use std::process::Command;
use std::str;

use crate::SpeedTestResults;

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct Root {
    #[serde(rename = "type")]
    type_field: String,
    timestamp: String,
    ping: SpeedtestNetPing,
    download: SpeedtestNetDownload,
    upload: SpeedtestNetUpload,
    packet_loss: i64,
    isp: String,
    interface: SpeedtestNetInterface,
    server: SpeedtestNetServer,
    result: SpeedtestNetResult,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct SpeedtestNetPing {
    jitter: f64,
    latency: f64,
    low: f64,
    high: f64,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct SpeedtestNetDownload {
    bandwidth: i64,
    bytes: i64,
    elapsed: i64,
    latency: SpeedtestNetLatency,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct SpeedtestNetLatency {
    iqm: f64,
    low: f64,
    high: f64,
    jitter: f64,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct SpeedtestNetUpload {
    bandwidth: i64,
    bytes: i64,
    elapsed: i64,
    latency: SpeedtestNetLatency,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct SpeedtestNetInterface {
    internal_ip: String,
    name: String,
    mac_addr: String,
    is_vpn: bool,
    external_ip: String,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct SpeedtestNetServer {
    id: i64,
    host: String,
    port: i64,
    name: String,
    location: String,
    country: String,
    ip: String,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct SpeedtestNetResult {
    id: String,
    url: String,
    persisted: bool,
}

pub fn run_speedtest_net_speedtest() -> Result<SpeedTestResults, Error> {
    let output = Command::new("speedtest")
        .args([
            "--format=json-pretty",
            "--progress=no",
            "--accept-license",
            "--accept-gdpr",
        ])
        .output()?
        .stdout;

    let data: Root = serde_json::from_slice(&output)?;

    Ok(SpeedTestResults {
        download_bits: (data.download.bandwidth * 8) as f64,
        upload_bits: (data.upload.bandwidth * 8) as f64,
        latency: data.ping.latency,
        jitter: Some(data.ping.jitter),
    })
}
