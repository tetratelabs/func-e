+++
fragment = "content"
date = "2021-07-06"
weight = 150
+++

func-e (pronounced funky) allows you to quickly see available versions of Envoy and try them out. This makes it easy to validate
configuration you would use in production. Each time you end a run, a snapshot of runtime state is taken on
your behalf. This makes knowledge sharing and troubleshooting easier, especially when upgrading. Try it out!

```sh
$ curl -L https://func-e.io/install.sh | bash -s -- -b /usr/local/bin
$ func-e run -c /path/to/envoy.yaml
# If you don't have a configuration file, you can start the admin port like this
$ func-e run --config-yaml "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}"
```
