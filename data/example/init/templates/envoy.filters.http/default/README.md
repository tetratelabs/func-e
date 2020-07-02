# Default example setup to demo an HTTP Filter

## Files

| File              | Description              | Purpose                                                                 |
| ----------------- | ------------------------ | ----------------------------------------------------------------------- |
| `example.yaml`    | `Example` descriptor     | Describes runtime requirements, e.g. a specific version of `Envoy`      |
| `envoy.tmpl.yaml` | `Envoy` bootstrap config | Provides `Envoy` config that demoes extension in action                 |
| `extension.json`  | `Extension` config       | Provides configuration for extension itself                             |

## Components

### Envoy config

#### Listeners

* [0.0.0.0:10000](http://0.0.0.0:10000) - represents an HTTP ingress
  * dispatches all requests to a `mock HTTP endpoint` (see below)
  * configured to use `HTTP Filter` extension
* [127.0.0.1:10001](http://127.0.0.1:10001) - represents a `mock HTTP endpoint`
  * responds to all HTTP requests with HTTP status `200`

### Extension config

Empty by default

## Request Flow

```
+--------+                +----------------------+              +----------------------------+
|        |   (requests)   | Envoy (HTTP ingress) | (dispatches) | Envoy (mock HTTP endpoint) |
| client | -------------> |                      | -----------> |                            |
|        |                | http://0.0.0.0:10000 |              |   http://127.0.0.1:10001   |
+--------+                +----------------------+              +----------------------------+
                                    | (uses)
                                    V
                          +----------------------+
                          |  HTTP Filter (Wasm)  |
                          +----------------------+
```

## How to use

1. Make HTTP request
   ```shell
   curl http://0.0.0.0:10000
   ```
2. Checkout `Envoy` stdout
