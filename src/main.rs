use argh::FromArgs;
use speedtest_exporter::exporter;

#[derive(FromArgs)]
/// Promethues exporter for running Speedtest via Cloudflare and Speedtest.net.
struct Args {
    /// server hostname
    #[argh(option, default = "String::from(\"127.0.0.1\")")]
    server_host: String,

    /// server port
    #[argh(option, default = "9100")]
    server_port: u16,
}

fn main() {
    let args: Args = argh::from_env();
    exporter(args.server_host, args.server_port);
}
