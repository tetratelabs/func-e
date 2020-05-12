use proxy_wasm::traits::StreamContext;
use proxy_wasm::types::LogLevel;

use envoy_sdk::extension;
use envoy_sdk::extension::filter::network;
use envoy_sdk::host::services::clients;
use envoy_sdk::host::services::time;

use envoy_sample_network_filter::SampleNetworkFilterFactory;

#[cfg_attr(
    all(
        target_arch = "wasm32",
        target_vendor = "unknown",
        target_os = "unknown"
    ),
    export_name = "_start"
)]
#[no_mangle]
/// Is called when a new instance of WebAssembly module gets created.
///
/// In general, a single WebAssembly module can include multiple extensions.
/// Envoy will instantiate the module on first use of any of them.
///
/// A module with an Network filter extension (like this one) will be
/// instantiated at a time Envoy creates a Listener that makes
/// use of that filter.
/// Nonetheless, lifecycle of a WebAssembly module instance is separate
/// from the one of a Listener.
/// A single WebAssembly module instance can be shared by multiple Listeners
/// and will outlive all of them.
extern "C" fn start() {
    proxy_wasm::set_log_level(LogLevel::Info);
    proxy_wasm::set_stream_context(|context_id, _| -> Box<dyn StreamContext> {
        // TODO: at the moment, extension configuration is ignored since it belongs to the RootContext
        // but `proxy-wasm` doesn't provide any way to associate StreamContext with its parent RootContext

        // Inject dependencies on Envoy host APIs
        let mut factory =
            SampleNetworkFilterFactory::new(&time::ops::Host, &clients::http::ops::Host);
        let network_filter =
            <SampleNetworkFilterFactory as extension::factory::Factory>::new_extension(
                &mut factory,
                extension::InstanceId::from(context_id),
            )
            .unwrap();
        Box::new(network::FilterContext::new(
            network_filter,
            &network::ops::Host,
            &clients::http::ops::Host,
        ))
    });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn should_start() {
        start()
    }
}
