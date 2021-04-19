use std::convert::TryFrom;

use envoy::extension::{access_logger, AccessLogger, ConfigStatus, Result};
use envoy::host::{log, ByteString, Stats};

use super::config::SampleAccessLoggerConfig;
use super::stats::SampleAccessLoggerStats;

/// Sample Access Logger.
pub struct SampleAccessLogger<'a> {
    config: SampleAccessLoggerConfig,
    stats: SampleAccessLoggerStats,
}

impl<'a> SampleAccessLogger<'a> {
    /// Creates a new instance of Sample Access Logger.
    pub fn new(stats: &dyn Stats) -> Result<Self> {
        let stats = SampleAccessLoggerStats::new(
            stats.counter("examples.access_logger.log_entries_total")?,
        );
        // Inject dependencies on Envoy host APIs
        Ok(SampleAccessLogger {
            config: SampleAccessLoggerConfig::default(),
            stats,
        })
    }

    /// Creates a new instance of Sample Access Logger
    /// bound to the actual Envoy ABI.
    pub fn default() -> Result<Self> {
        Self::new(Stats::default())
    }
}

impl<'a> AccessLogger for SampleAccessLogger<'a> {
    /// The reference name for Sample Access Logger.
    ///
    /// This name appears in `Envoy` configuration as a value of `root_id` field.
    fn name() -> &'static str {
        "{{ .Extension.Name }}"
    }

    /// Is called when Envoy creates a new Listener that uses Sample Access Logger.
    fn on_configure(
        &mut self,
        config: ByteString,
        _ops: &dyn access_logger::ConfigureOps,
    ) -> Result<ConfigStatus> {
        self.config = if config.is_empty() {
            SampleAccessLoggerConfig::default()
        } else {
            SampleAccessLoggerConfig::try_from(config.as_bytes())?
        };
        Ok(ConfigStatus::Accepted)
    }

    /// Is called to log a complete TCP connection or HTTP request.
    ///
    /// Use `log_ops` to get ahold of request/response headers,
    /// TCP connection properties, etc.
    fn on_log(&mut self, log_ops: &dyn access_logger::LogOps) -> Result<()> {
        // Update stats
        self.stats.log_entries_total().inc()?;

        log::info!(
            "logging with config: {:?}",
            self.config,
        );

        log::info!("  request headers:");
        let request_headers = log_ops.request_headers()?;
        for (name, value) in &request_headers {
            log::info!("    {}: {}", name, value);
        }
        log::info!("  response headers:");
        let response_headers = log_ops.response_headers()?;
        for (name, value) in &response_headers {
            log::info!("    {}: {}", name, value);
        }
        let upstream_address = log_ops.stream_info().upstream().address()?;
        log::info!("  upstream info:");
        log::info!("    {}: {:?}", "upstream.address", upstream_address);

        Ok(())
    }
}
