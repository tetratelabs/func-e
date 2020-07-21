use std::convert::TryFrom;
use std::rc::Rc;

use envoy::extension::{factory, ConfigStatus, ExtensionFactory, InstanceId, Result};
use envoy::host::{Clock, Stats};

use super::config::SampleHttpFilterConfig;
use super::filter::SampleHttpFilter;
use super::stats::SampleHttpFilterStats;

/// Factory for creating Sample HTTP Filter instances
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
    clock: &'a dyn Clock,
}

impl<'a> SampleHttpFilterFactory<'a> {
    /// Creates a new factory.
    pub fn new(clock: &'a dyn Clock, stats: &dyn Stats) -> Result<Self> {
        let stats =
            SampleHttpFilterStats::new(stats.counter("examples.http_filter.requests_total")?);
        // Inject dependencies on Envoy host APIs
        Ok(SampleHttpFilterFactory {
            config: Rc::new(SampleHttpFilterConfig::default()),
            stats: Rc::new(stats),
            clock,
        })
    }

    /// Creates a new factory bound to the actual Envoy ABI.
    pub fn default() -> Result<Self> {
        Self::new(Clock::default(), Stats::default())
    }
}

impl<'a> ExtensionFactory for SampleHttpFilterFactory<'a> {
    type Extension = SampleHttpFilter<'a>;

    /// The reference name for Sample HTTP Filter.
    ///
    /// This name appears in `Envoy` configuration as a value of `root_id` field
    /// (also known as `group_name`).
    const NAME: &'static str = "{{ .Extension.Name }}";

    /// Is called when Envoy creates a new Listener that uses Sample HTTP Filter.
    fn on_configure(
        &mut self,
        _configuration_size: usize,
        ops: &dyn factory::ConfigureOps,
    ) -> Result<ConfigStatus> {
        let config = match ops.configuration()? {
            Some(bytes) => SampleHttpFilterConfig::try_from(bytes.as_slice())?,
            None => SampleHttpFilterConfig::default(),
        };
        self.config = Rc::new(config);
        Ok(ConfigStatus::Accepted)
    }

    /// Is called to create a unique instance of Sample HTTP Filter
    /// for each HTTP request.
    fn new_extension(&mut self, instance_id: InstanceId) -> Result<Self::Extension> {
        Ok(SampleHttpFilter::new(
            Rc::clone(&self.config),
            Rc::clone(&self.stats),
            instance_id,
            self.clock,
        ))
    }
}
