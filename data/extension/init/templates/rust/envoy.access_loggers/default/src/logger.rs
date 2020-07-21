use std::convert::TryFrom;
use std::time::Duration;

use log::{error, info};

use envoy_sdk::extension::access_logger;
use envoy_sdk::extension::Result;
use envoy_sdk::host::services::{clients, metrics, time};

use chrono::offset::Local;
use chrono::DateTime;

use super::config::SampleAccessLoggerConfig;
use super::stats::SampleAccessLoggerStats;

/// Sample Access Logger.
pub struct SampleAccessLogger<'a> {
    config: SampleAccessLoggerConfig,
    stats: SampleAccessLoggerStats,
    // This example shows how to use Time API, HTTP Client API and
    // Metrics API provided by Envoy host.
    time_service: &'a dyn time::Service,
    http_client: &'a dyn clients::http::Client,

    active_request: Option<clients::http::RequestHandle>,
}

impl<'a> SampleAccessLogger<'a> {
    /// Creates a new instance of sample access logger.
    pub fn new(
        time_service: &'a dyn time::Service,
        http_client: &'a dyn clients::http::Client,
        metrics_service: &'a dyn metrics::Service,
    ) -> Result<Self> {
        let stats = SampleAccessLoggerStats::new(
            metrics_service.counter("examples.access_logger.requests_total")?,
            metrics_service.gauge("examples.access_logger.reports_active")?,
            metrics_service.counter("examples.access_logger.reports_total")?,
        );
        // Inject dependencies on Envoy host APIs
        Ok(SampleAccessLogger {
            config: SampleAccessLoggerConfig::default(),
            stats,
            time_service,
            http_client,
            active_request: None,
        })
    }

    /// Creates a new instance of sample access logger
    /// bound to the actual Envoy ABI.
    pub fn with_default_ops() -> Result<Self> {
        SampleAccessLogger::new(
            &time::ops::Host,
            &clients::http::ops::Host,
            &metrics::ops::Host,
        )
    }
}

impl<'a> access_logger::Logger for SampleAccessLogger<'a> {
    /// Is called when Envoy creates a new Listener that uses sample access logger.
    ///
    /// Use logger_ops to get ahold of configuration.
    fn on_configure(
        &mut self,
        _configuration_size: usize,
        logger_ops: &dyn access_logger::ConfigureOps,
    ) -> Result<bool> {
        self.config = match logger_ops.get_configuration()? {
            Some(bytes) => match SampleAccessLoggerConfig::try_from(bytes.as_ref()) {
                Ok(value) => value,
                Err(err) => {
                    error!("failed to parse extension configuration: {}", err);
                    return Ok(false);
                }
            },
            None => SampleAccessLoggerConfig::default(),
        };
        Ok(true)
    }

    /// Is called to log a complete TCP connection or HTTP request.
    ///
    /// Use logger_ops to get ahold of request/response headers,
    /// TCP connection properties, etc.
    fn on_log(&mut self, logger_ops: &dyn access_logger::LogOps) -> Result<()> {
        // Update stats
        self.stats.requests_total().inc()?;

        let current_time = self.time_service.get_current_time()?;
        let datetime: DateTime<Local> = current_time.into();

        info!(
            "logging at {} with config: {:?}",
            datetime.format("%+"),
            self.config,
        );

        info!("  request headers:");
        let request_headers = logger_ops.get_request_headers()?;
        for (name, value) in &request_headers {
            info!("    {}: {}", name, value);
        }
        info!("  response headers:");
        let response_headers = logger_ops.get_response_headers()?;
        for (name, value) in &response_headers {
            info!("    {}: {}", name, value);
        }
        let upstream_address = logger_ops.get_property(vec!["upstream", "address"])?;
        let upstream_address = upstream_address
            .map(|value| String::from_utf8(value).unwrap())
            .unwrap_or_else(String::new);
        info!("  upstream info:");
        info!("    {}: {}", "upstream.address", upstream_address);

        // simulate sending a log entry off
        self.active_request = Some(self.http_client.send_request(
            "mock_service",
            vec![
                (":method", "GET"),
                (":path", "/mock"),
                (":authority", "mock.local"),
            ],
            None,
            vec![],
            Duration::from_secs(3),
        )?);
        info!(
            "sent request to a log collector: @{}",
            self.active_request.as_ref().unwrap()
        );
        // Update stats
        self.stats.reports_active().inc()?;

        Ok(())
    }

    // HTTP Client API callbacks

    /// Is called when an auxiliary HTTP request sent via HTTP Client API
    /// is finally complete.
    ///
    /// Use http_client_ops to get ahold of response headers, body, etc.
    fn on_http_call_response(
        &mut self,
        request: clients::http::RequestHandle,
        num_headers: usize,
        _body_size: usize,
        _num_trailers: usize,
        http_client_ops: &dyn clients::http::ResponseOps,
    ) -> Result<()> {
        info!(
            "received response from a log collector on request: @{}",
            request
        );
        assert!(self.active_request == Some(request));
        self.active_request = None;

        // Update stats
        self.stats.reports_active().dec()?;
        self.stats.reports_total().inc()?;

        info!("  headers[count={}]:", num_headers);
        let response_headers = http_client_ops.get_http_call_response_headers()?;
        for (name, value) in &response_headers {
            info!("    {}: {}", name, value);
        }

        Ok(())
    }
}
