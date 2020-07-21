use std::rc::Rc;
use std::time::Duration;

use log::info;

use envoy_sdk::extension::filter::http;
use envoy_sdk::extension::{InstanceId, Result};
use envoy_sdk::host::services::clients;
use envoy_sdk::host::services::time;

use chrono::offset::Local;
use chrono::DateTime;

use super::config::SampleHttpFilterConfig;
use super::stats::SampleHttpFilterStats;

// Sample HTTP filter.
pub struct SampleHttpFilter<'a> {
    // This example shows how multiple filter instances could share
    // the same configuration.
    config: Rc<SampleHttpFilterConfig>,
    // This example shows how multiple filter instances could share
    // metrics.
    stats: Rc<SampleHttpFilterStats>,
    instance_id: InstanceId,
    // This example shows how to use Time API, HTTP Client API and
    // Metrics API provided by Envoy host.
    time_service: &'a dyn time::Service,
    http_client: &'a dyn clients::http::Client,

    active_request: Option<clients::http::RequestHandle>,
    response_body_size: u64,
}

impl<'a> SampleHttpFilter<'a> {
    /// Creates a new instance of sample HTTP filter.
    pub fn new(
        config: Rc<SampleHttpFilterConfig>,
        stats: Rc<SampleHttpFilterStats>,
        instance_id: InstanceId,
        time_service: &'a dyn time::Service,
        http_client: &'a dyn clients::http::Client,
    ) -> Self {
        // Inject dependencies on Envoy host APIs
        SampleHttpFilter {
            config,
            stats,
            instance_id,
            time_service,
            http_client,
            active_request: None,
            response_body_size: 0,
        }
    }
}

impl<'a> http::Filter for SampleHttpFilter<'a> {
    /// Is called when HTTP request headers have been received.
    ///
    /// Use filter_ops to access and mutate request headers.
    fn on_request_headers(
        &mut self,
        _num_headers: usize,
        filter_ops: &dyn http::RequestHeadersOps,
    ) -> Result<http::FilterHeadersStatus> {
        // Update stats
        self.stats.requests_active().inc()?;

        let current_time = self.time_service.get_current_time()?;
        let datetime: DateTime<Local> = current_time.into();

        info!(
            "#{} new http exchange starts at {} with config: {:?}",
            self.instance_id,
            datetime.format("%+"),
            self.config,
        );

        info!("#{} observing request headers", self.instance_id);
        for (name, value) in &filter_ops.get_request_headers()? {
            info!("#{} -> {}: {}", self.instance_id, name, value);
        }

        match filter_ops.get_request_header(":path")? {
            Some(path) if path == "/ping" => {
                filter_ops.send_response(
                    200,
                    vec![("x-sample-response", "pong")],
                    Some(b"Pong!\n"),
                )?;
                Ok(http::FilterHeadersStatus::Pause)
            }
            Some(path) if path == "/secret" => {
                self.active_request = Some(self.http_client.send_request(
                    "mock_service",
                    vec![
                        (":method", "GET"),
                        (":path", "/authz"),
                        (":authority", "mock.local"),
                    ],
                    None,
                    vec![],
                    Duration::from_secs(3),
                )?);
                info!(
                    "#{} sent authorization request: @{}",
                    self.instance_id,
                    self.active_request.as_ref().unwrap()
                );
                info!("#{} suspending http exchange processing", self.instance_id);
                Ok(http::FilterHeadersStatus::Pause)
            }
            _ => Ok(http::FilterHeadersStatus::Continue),
        }
    }

    /// Is called when HTTP response headers have been received.
    ///
    /// Use filter_ops to access and mutate response headers.
    fn on_response_headers(
        &mut self,
        _num_headers: usize,
        filter_ops: &dyn http::ResponseHeadersOps,
    ) -> Result<http::FilterHeadersStatus> {
        info!("#{} observing response headers", self.instance_id);
        for (name, value) in &filter_ops.get_response_headers()? {
            info!("#{} <- {}: {}", self.instance_id, name, value);
        }
        Ok(http::FilterHeadersStatus::Continue)
    }

    /// Is called on response body part.
    fn on_response_body(
        &mut self,
        data_size: usize,
        _end_of_stream: bool,
        _ops: &dyn http::ResponseBodyOps,
    ) -> Result<http::FilterDataStatus> {
        self.response_body_size += data_size as u64;

        Ok(http::FilterDataStatus::Continue)
    }

    /// Is called when HTTP exchange is complete.
    fn on_exchange_complete(&mut self) -> Result<()> {
        // Update stats
        self.stats.requests_active().dec()?;
        self.stats.requests_total().inc()?;
        self.stats
            .response_body_size_bytes()
            .record(self.response_body_size)?;

        info!("#{} http exchange complete", self.instance_id);
        Ok(())
    }

    // HTTP Client API callbacks

    /// Is called when an auxiliary HTTP request sent via HTTP Client API
    /// is finally complete.
    ///
    /// Use http_client_ops to get ahold of response headers, body, etc.
    ///
    /// Use filter_ops to amend and resume HTTP exchange.
    fn on_http_call_response(
        &mut self,
        request: clients::http::RequestHandle,
        num_headers: usize,
        _body_size: usize,
        _num_trailers: usize,
        filter_ops: &dyn http::Ops,
        http_client_ops: &dyn clients::http::ResponseOps,
    ) -> Result<()> {
        info!(
            "#{} received response on authorization request: @{}",
            self.instance_id, request
        );
        assert!(self.active_request == Some(request));
        self.active_request = None;

        info!("     headers[count={}]:", num_headers);
        let response_headers = http_client_ops.get_http_call_response_headers()?;
        for (name, value) in &response_headers {
            info!("       {}: {}", name, value);
        }

        info!("#{} resuming http exchange processing", self.instance_id);
        filter_ops.resume_request()?;
        Ok(())
    }
}
