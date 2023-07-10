## Quick-start: InterLink
### Install binaries

```bash
curl -sfL https://intertwin-eu.github.io/interLink/itwinctl.sh | sh -s - install
```

### Start daemons

```bash
curl -sfL https://intertwin-eu.github.io/interLink/itwinctl.sh | sh -s - start
```

### Restart daemons

```bash
curl -sfL https://intertwin-eu.github.io/interLink/itwinctl.sh | sh -s - restart
```

### Stop daemons

```bash
curl -sfL https://intertwin-eu.github.io/interLink/itwinctl.sh | sh -s - stop
```

## Configuration options


### :closed_lock_with_key: Authentication
InterLink supports OAuth2 proxy authentication, allowing you to set up an authorized group (or managing single-user access) to access services. In order to use it, set the InterLinkPort field to 8080 and run InterLink executable by executing the docs/itwinctl.sh script. The provided script will run InterLink and Slurm sidecar binaries, but you can easily edit it to run another sidecar.
First time running the script, run ```source itwinctl.sh install```, to download and setup the OAuth2 proxy.
From now on, you can just use ```source itwinctl.sh start/stop/restart``` to manage your applications.
Remember to generate your token and set the VKTOKENFILE environment variable, otherwise any connection between VK and InterLink will be refused with a 403 Fobidden reply.