use std::convert::TryFrom;
use std::rc::Rc;

use log::error;

use envoy_sdk::extension;
use envoy_sdk::extension::{InstanceId, Result};
use envoy_sdk::host::services::{clients, metrics, time};

use super::config::SampleHttpFilterConfig;
use super::filter::SampleHttpFilter;
use super::stats::SampleHttpFilterStats;

/// Factory for creating sample HTTP filter instances
/// (one filter instance per HTTP request).
pub struct SampleHttpFilterFactory<'a> {
    // This example shows how multiple filter instances could share
    // the same configuration.
    config: Rc<SampleHttpFilterConfig>,
    // This example shows how multiple filter instances could share
    // metrics.
    stats: Rc<SampleHttpFilterStats>,
    // This example shows how to use Time API, HTTP Client API and
    // Metrics API provided by Envoy host.
    time_service: &'a dyn time::Service,
    http_client: &'a dyn clients::http::Client,
}

impl<'a> SampleHttpFilterFactory<'a> {
    /// Creates a new factory.
    pub fn new(
        time_service: &'a dyn time::Service,
        http_client: &'a dyn clients::http::Client,
        metrics_service: &'a dyn metrics::Service,
    ) -> Result<Self> {
        let stats = SampleHttpFilterStats::new(
            metrics_service.counter("examples.http_filter.requests_total")?,
            metrics_service.gauge("examples.http_filter.requests_active")?,
            metrics_service.histogram("examples.http_filter.response_body_size_bytes")?,
        );
        // Inject dependencies on Envoy host APIs
        Ok(SampleHttpFilterFactory {
            config: Rc::new(SampleHttpFilterConfig::default()),
            stats: Rc::new(stats),
            time_service,
            http_client,
        })
    }

    /// Creates a new factory bound to the actual Envoy ABI.
    pub fn with_default_ops() -> Result<Self> {
        SampleHttpFilterFactory::new(
            &time::ops::Host,
            &clients::http::ops::Host,
            &metrics::ops::Host,
        )
    }
}

impl<'a> extension::Factory for SampleHttpFilterFactory<'a> {
    type Extension = SampleHttpFilter<'a>;

    /// The reference name for sample HTTP filter.
    ///
    /// This name appears in Envoy configuration as a value of group_name (aka, root_id) field.
    const NAME: &'static str = "";

    /// Is called when Envoy creates a new Listener that uses sample HTTP filter.
    fn on_configure(
        &mut self,
        _configuration_size: usize,
        ops: &dyn extension::factory::ConfigureOps,
    ) -> Result<bool> {
        let config = match ops.get_configuration()? {
            Some(bytes) => match SampleHttpFilterConfig::try_from(bytes.as_ref()) {
                Ok(value) => value,
                Err(err) => {
                    error!("failed to parse extension configuration: {}", err);
                    return Ok(false);
                }
            },
            None => SampleHttpFilterConfig::default(),
        };
        self.config = Rc::new(config);
        Ok(true)
    }

    /// Is called to create a unique instance of sample HTTP filter
    /// for each HTTP request.
    fn new_extension(&mut self, instance_id: InstanceId) -> Result<SampleHttpFilter<'a>> {
        Ok(SampleHttpFilter::new(
            Rc::clone(&self.config),
            Rc::clone(&self.stats),
            instance_id,
            self.time_service,
            self.http_client,
        ))
    }
}
