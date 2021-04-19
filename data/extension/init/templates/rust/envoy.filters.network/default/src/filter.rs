use std::rc::Rc;

use envoy::extension::{filter::network, InstanceId, NetworkFilter, Result};
use envoy::host::log::info;
use envoy::host::Clock;

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
    // This example shows how to use Time API provided by Envoy host.
    clock: &'a dyn Clock,
}

impl<'a> SampleNetworkFilter<'a> {
    /// Creates a new instance of Sample Network Filter.
    pub fn new(
        config: Rc<SampleNetworkFilterConfig>,
        stats: Rc<SampleNetworkFilterStats>,
        instance_id: InstanceId,
        clock: &'a dyn Clock,
    ) -> Self {
        // Inject dependencies on Envoy host APIs
        SampleNetworkFilter {
            config,
            stats,
            instance_id,
            clock,
        }
    }
}

impl<'a> NetworkFilter for SampleNetworkFilter<'a> {
    /// Called when a new TCP connection is opened.
    fn on_new_connection(&mut self) -> Result<network::FilterStatus> {
        info!(
            "#{} new TCP connection starts at {:?} with config: {:?}",
            self.instance_id,
            self.clock.now(),
            self.config,
        );

        Ok(network::FilterStatus::Continue)
    }

    /// Called when TCP connection is complete.
    ///
    /// This moment happens before `Access Loggers` get called.
    fn on_connection_complete(&mut self, _ops: &dyn network::ConnectionCompleteOps) -> Result<()> {
        // Update stats
        self.stats.connections_total().inc()?;

        info!("#{} connection complete", self.instance_id);
        Ok(())
    }
}
