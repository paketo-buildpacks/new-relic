# `docker.io/paketobuildpacks/new-relic`
The Paketo Buildpack for New Relic is a Cloud Native Buildpack that contributes the [New Relic][n] Agent and configures it to connect to the service.

[n]: https://newrelic.com

## Behavior
This buildpack will participate if all the following conditions are met

* A binding exists with `type` of `NewRelic`

The buildpack will do the following for Java applications:

* Contributes a Java agent to a layer and configures `$JAVA_TOOL_OPTIONS` to use it
  * Contributes a default `newrelic.yml`
* Contribute extensions if available
* Transforms the contents of the binding secret to environment variables with the pattern `NEW_RELIC_<KEY>=<VALUE>`

The buildpack will do the following for NodeJS applications:

* Contributes a NodeJS agent to a layer and configures `$NODE_MODULES` to use it
  * Contributes a default `newrelic.js`
* If main module does not already require `newrelic` module, prepends the main module with `require('newrelic');`
* Transforms the contents of the binding secret to environment variables with the pattern `NEW_RELIC_<KEY>=<VALUE>`

The buildpack will do the following for PHP applications:

* Contributes a PHP agent to a layer and configures `$PHP_INI_SCAN_DIR` to use it
* Transforms the contents of the binding secret to environment variables with the pattern `NEW_RELIC_<KEY>=<VALUE>`

## Configuration
| Environment Variable | Description
| -------------------- | -----------
| `$BP_NEW_RELIC_EXT_SHA256` | Configure the SHA256 hash of the New Relic extensions archive
| `$BP_NEW_RELIC_EXT_STRIP` | Configure the number of directory components to strip from the New Relic extensions archive. Defaults to `0`.
| `$BP_NEW_RELIC_EXT_URI` | Configure the download location of the New Relic extensions
| `$BP_NEW_RELIC_EXT_VERSION` | Configure the version of the New Relic extensions

## Bindings
The buildpack optionally accepts the following bindings:

### Type: `dependency-mapping`
|Key                   | Value   | Description
|----------------------|---------|------------
|`<dependency-digest>` | `<uri>` | If needed, the buildpack will fetch the dependency with digest `<dependency-digest>` from `<uri>`

## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
