use std::rc::Rc;

use envoy::extension::{filter::http, HttpFilter, InstanceId, Result};
use envoy::host::log::info;
use envoy::host::Clock;

use super::config::SampleHttpFilterConfig;
use super::stats::SampleHttpFilterStats;

// Sample HTTP Filter.
pub struct SampleHttpFilter<'a> {
    // This example shows how multiple filter instances could share
    // the same configuration.
    config: Rc<SampleHttpFilterConfig>,
    // This example shows how multiple filter instances could share
    // metrics.
    stats: Rc<SampleHttpFilterStats>,
    instance_id: InstanceId,
    // This example shows how to use Time API provided by Envoy host.
    clock: &'a dyn Clock,
}

impl<'a> SampleHttpFilter<'a> {
    /// Creates a new instance of Sample HTTP Filter.
    pub fn new(
        config: Rc<SampleHttpFilterConfig>,
        stats: Rc<SampleHttpFilterStats>,
        instance_id: InstanceId,
        clock: &'a dyn Clock,
    ) -> Self {
        // Inject dependencies on Envoy host APIs
        SampleHttpFilter {
            config,
            stats,
            instance_id,
            clock,
        }
    }
}

impl<'a> HttpFilter for SampleHttpFilter<'a> {
    /// Called when HTTP request headers have been received.
    ///
    /// Use filter_ops to access and mutate request headers.
    fn on_request_headers(
        &mut self,
        _num_headers: usize,
        _end_of_stream: bool,
        filter_ops: &dyn http::RequestHeadersOps,
    ) -> Result<http::FilterHeadersStatus> {
        info!(
            "#{} new http exchange starts at {:?} with config: {:?}",
            self.instance_id,
            self.clock.now(),
            self.config,
        );

        info!("#{} observing request headers", self.instance_id);
        for (name, value) in &filter_ops.request_headers()? {
            info!("#{} -> {}: {}", self.instance_id, name, value);
        }

        Ok(http::FilterHeadersStatus::Continue)
    }

    /// Called when HTTP stream is complete.
    fn on_exchange_complete(&mut self, _ops: &dyn http::ExchangeCompleteOps) -> Result<()> {
        // Update stats
        self.stats.requests_total().inc()?;

        info!("#{} http exchange complete", self.instance_id);
        Ok(())
    }
}
