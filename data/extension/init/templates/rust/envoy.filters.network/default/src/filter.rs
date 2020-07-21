use std::rc::Rc;
use std::time::Duration;

use log::info;

use envoy_sdk::extension::filter::network;
use envoy_sdk::extension::{InstanceId, Result};
use envoy_sdk::host::services::clients;
use envoy_sdk::host::services::time;

use chrono::offset::Local;
use chrono::DateTime;

use super::config::SampleNetworkFilterConfig;
use super::stats::SampleNetworkFilterStats;

/// Sample network filter.
pub struct SampleNetworkFilter<'a> {
    // This example shows how multiple filter instances could share
    // the same configuration.
    config: Rc<SampleNetworkFilterConfig>,
    // This example shows how multiple filter instances could share
    // metrics.
    stats: Rc<SampleNetworkFilterStats>,
    instance_id: InstanceId,
    // This example shows how to use Time API, HTTP Client API and
    // Metrics API provided by Envoy host.
    time_service: &'a dyn time::Service,
    http_client: &'a dyn clients::http::Client,

    active_request: Option<clients::http::RequestHandle>,
    response_body_size: u64,
}

impl<'a> SampleNetworkFilter<'a> {
    /// Creates a new instance of sample network filter.
    pub fn new(
        config: Rc<SampleNetworkFilterConfig>,
        stats: Rc<SampleNetworkFilterStats>,
        instance_id: InstanceId,
        time_service: &'a dyn time::Service,
        http_client: &'a dyn clients::http::Client,
    ) -> Self {
        // Inject dependencies on Envoy host APIs
        SampleNetworkFilter {
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

impl<'a> network::Filter for SampleNetworkFilter<'a> {
    /// Is called when a new TCP connection is opened.
    fn on_new_connection(&mut self) -> Result<network::FilterStatus> {
        // Update stats
        self.stats.requests_active().inc()?;

        let current_time = self.time_service.get_current_time()?;
        let datetime: DateTime<Local> = current_time.into();

        info!(
            "#{} new TCP connection starts at {} with config: {:?}",
            self.instance_id,
            datetime.format("%+"),
            self.config,
        );

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
            "#{} sent outgoing request: @{}",
            self.instance_id,
            self.active_request.as_ref().unwrap()
        );

        Ok(network::FilterStatus::Pause)
    }

    /// Is called on response body part.
    fn on_upstream_data(
        &mut self,
        data_size: usize,
        _end_of_stream: bool,
        _ops: &dyn network::UpstreamDataOps,
    ) -> Result<network::FilterStatus> {
        self.response_body_size += data_size as u64;

        Ok(network::FilterStatus::Continue)
    }

    /// Is called when the TCP connection is complete.
    fn on_connection_complete(&mut self) -> Result<()> {
        // Update stats
        self.stats.requests_active().dec()?;
        self.stats.requests_total().inc()?;
        self.stats
            .response_body_size_bytes()
            .record(self.response_body_size)?;

        info!("#{} TCP connection ended", self.instance_id);
        Ok(())
    }

    // HTTP Client API callbacks

    /// Is called when an auxiliary HTTP request sent via HTTP Client API
    /// is finally complete.
    ///
    /// Use http_client_ops to get ahold of response headers, body, etc.
    ///
    /// Use filter_ops to amend and resume TCP flow.
    fn on_http_call_response(
        &mut self,
        request: clients::http::RequestHandle,
        num_headers: usize,
        _body_size: usize,
        _num_trailers: usize,
        _filter_ops: &dyn network::Ops,
        http_client_ops: &dyn clients::http::ResponseOps,
    ) -> Result<()> {
        info!(
            "#{} received response on outgoing request: @{}",
            self.instance_id, request
        );
        assert!(self.active_request == Some(request));
        self.active_request = None;

        info!("     headers[count={}]:", num_headers);
        let response_headers = http_client_ops.get_http_call_response_headers()?;
        for (name, value) in &response_headers {
            info!("       {}: {}", name, value);
        }

        // TODO: no way to resume tcp stream
        Ok(())
    }
}
