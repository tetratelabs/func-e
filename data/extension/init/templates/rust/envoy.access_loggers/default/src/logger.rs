use std::convert::TryFrom;

use chrono::{offset::Local, DateTime};

use envoy::extension::{access_logger, AccessLogger, ConfigStatus, Result};
use envoy::host::log::info;
use envoy::host::{Clock, Stats};

use super::config::SampleAccessLoggerConfig;
use super::stats::SampleAccessLoggerStats;

/// Sample Access Logger.
pub struct SampleAccessLogger<'a> {
    config: SampleAccessLoggerConfig,
    stats: SampleAccessLoggerStats,
    // This example shows how to use Time API provided by Envoy host.
    clock: &'a dyn Clock,
}

impl<'a> SampleAccessLogger<'a> {
    /// Creates a new instance of Sample Access Logger.
    pub fn new(clock: &'a dyn Clock, stats: &dyn Stats) -> Result<Self> {
        let stats = SampleAccessLoggerStats::new(
            stats.counter("examples.access_logger.log_entries_total")?,
        );
        // Inject dependencies on Envoy host APIs
        Ok(SampleAccessLogger {
            config: SampleAccessLoggerConfig::default(),
            stats,
            clock,
        })
    }

    /// Creates a new instance of Sample Access Logger
    /// bound to the actual Envoy ABI.
    pub fn default() -> Result<Self> {
        Self::new(Clock::default(), Stats::default())
    }
}

impl<'a> AccessLogger for SampleAccessLogger<'a> {
    /// The reference name for Sample Access Logger.
    ///
    /// This name appears in `Envoy` configuration as a value of `root_id` field
    /// (also known as `group_name`).
    const NAME: &'static str = "{{ .Extension.Name }}";

    /// Is called when Envoy creates a new Listener that uses Sample Access Logger.
    ///
    /// Use logger_ops to get ahold of configuration.
    fn on_configure(
        &mut self,
        _configuration_size: usize,
        logger_ops: &dyn access_logger::ConfigureOps,
    ) -> Result<ConfigStatus> {
        self.config = match logger_ops.configuration()? {
            Some(bytes) => SampleAccessLoggerConfig::try_from(bytes.as_slice())?,
            None => SampleAccessLoggerConfig::default(),
        };
        Ok(ConfigStatus::Accepted)
    }

    /// Is called to log a complete TCP connection or HTTP request.
    ///
    /// Use logger_ops to get ahold of request/response headers,
    /// TCP connection properties, etc.
    fn on_log(&mut self, logger_ops: &dyn access_logger::LogOps) -> Result<()> {
        // Update stats
        self.stats.log_entries_total().inc()?;

        let now: DateTime<Local> = self.clock.now()?.into();

        info!(
            "logging at {} with config: {:?}",
            now.format("%+"),
            self.config,
        );

        info!("  request headers:");
        let request_headers = logger_ops.request_headers()?;
        for (name, value) in &request_headers {
            info!("    {}: {}", name, value);
        }
        info!("  response headers:");
        let response_headers = logger_ops.response_headers()?;
        for (name, value) in &response_headers {
            info!("    {}: {}", name, value);
        }
        let upstream_address = logger_ops.stream_property(vec!["upstream", "address"])?;
        let upstream_address = upstream_address
            .map(String::from_utf8)
            .transpose()?
            .unwrap_or_else(String::default);
        info!("  upstream info:");
        info!("    {}: {}", "upstream.address", upstream_address);

        Ok(())
    }
}
