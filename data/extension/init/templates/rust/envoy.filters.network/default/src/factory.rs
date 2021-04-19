use std::convert::TryFrom;
use std::rc::Rc;

use envoy::extension::{factory, ConfigStatus, ExtensionFactory, InstanceId, Result};
use envoy::host::{ByteString, Clock, Stats};

use super::config::SampleNetworkFilterConfig;
use super::filter::SampleNetworkFilter;
use super::stats::SampleNetworkFilterStats;

/// Factory for creating Sample Network Filter instances
/// (one filter instance per TCP connection).
pub struct SampleNetworkFilterFactory<'a> {
    // This example shows how multiple filter instances could share
    // the same configuration.
    config: Rc<SampleNetworkFilterConfig>,
    // This example shows how multiple filter instances could share
    // metrics.
    stats: Rc<SampleNetworkFilterStats>,
    // This example shows how to use Time API provided by Envoy host.
    clock: &'a dyn Clock,
}

impl<'a> SampleNetworkFilterFactory<'a> {
    /// Creates a new factory.
    pub fn new(clock: &'a dyn Clock, stats: &dyn Stats) -> Result<Self> {
        let stats = SampleNetworkFilterStats::new(
            stats.counter("examples.network_filter.connections_total")?,
        );
        // Inject dependencies on Envoy host APIs
        Ok(SampleNetworkFilterFactory {
            config: Rc::new(SampleNetworkFilterConfig::default()),
            stats: Rc::new(stats),
            clock,
        })
    }

    /// Creates a new factory bound to the actual Envoy ABI.
    pub fn default() -> Result<Self> {
        Self::new(Clock::default(), Stats::default())
    }
}

impl<'a> ExtensionFactory for SampleNetworkFilterFactory<'a> {
    type Extension = SampleNetworkFilter<'a>;

    /// The reference name for Sample Network Filter.
    ///
    /// This name appears in `Envoy` configuration as a value of `root_id` field.
    fn name() -> &'static str {
        "{{ .Extension.Name }}"
    }

    /// Is called when Envoy creates a new Listener that uses Sample Network Filter.
    fn on_configure(
        &mut self,
        config: ByteString,
        _ops: &dyn factory::ConfigureOps,
    ) -> Result<ConfigStatus> {
        let config = if config.is_empty() {
            SampleNetworkFilterConfig::default()
        } else {
            SampleNetworkFilterConfig::try_from(config.as_bytes())?
        };
        self.config = Rc::new(config);
        Ok(ConfigStatus::Accepted)
    }

    /// Is called to create a unique instance of Sample Network Filter
    /// for each TCP connection.
    fn new_extension(&mut self, instance_id: InstanceId) -> Result<Self::Extension> {
        Ok(SampleNetworkFilter::new(
            Rc::clone(&self.config),
            Rc::clone(&self.stats),
            instance_id,
            self.clock,
        ))
    }
}
