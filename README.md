# srtunectl

**srtunectl** is a lightweight utility for managing system routes (`ip route`) based on a configuration file. It is designed for setups using `tun` interfaces, such as Shadowsocks `ss-tun` or other VPN/proxy tunneling tools.

This tool adds or removes routes dynamically according to your configuration and can run as a persistent systemd service.

All JSON files located in the `data` folder will be processed and attempted to be added as routes through the specified tun device. Additionally, any routes defined in the `default_route` folder will be treated as forced routes for another interface, for example `eth0`.

---

## Features
- Manage routes via a simple configuration file
- Automatically route selected subnets through a tun interface
- Process JSON route files from `data` folder
- Force default routes via `default_route` folder
- Can be run manually or as a systemd service
- Easy build and installation via `make`

---

## Requirements
- Linux with `iproute2`
- Go toolchain (for building from source)
- Root privileges for modifying system routes

---

## Installation

1. Copy the example configuration:
```bash
cp iproute.conf.example iproute.conf
```

2. Build the binary:
```bash
make
```

3. Install:
```bash
sudo make install
```

After installation, the binary will be available as:
```
/usr/local/bin/srtunectl
```

---

## Configuration

The configuration file format is illustrated in `iproute.conf.example`.
Your configuration will typically include:

- List of networks to route
- Tun interface name (e.g., `tun0`)
- Gateway address
- default gateway address to set default route via them
- default interface name

Additionally:
- JSON files in the `data` folder will be parsed and added as routes through the tun device.
- Routes in the `default_route` folder will be processed as forced routes through other interfaces (e.g., `eth0`).

You may create or edit `iproute.conf` and manage JSON route files to match your routing needs.

---

## Usage

### Manual run
```bash
sudo srtunectl
```

### Verify routes
```bash
ip route
```

---

## Systemd Service

The repository includes a sample systemd unit file `route.service`. To use it:

```bash
sudo cp route.service /etc/systemd/system/srtunectl.service
sudo systemctl daemon-reload
sudo systemctl enable --now srtunectl.service
```

View service logs:
```bash
sudo journalctl -u srtunectl.service -f
```

---

## Debugging
- Ensure the tun interface exists:
  ```bash
  ip a
  ```
- Verify gateway, interface name, and subnets in `iproute.conf`
- Check systemd logs when running as a service
- Check that JSON files in `data` are valid and properly formatted

---

## Contributing
Pull requests, bug reports, and feature suggestions are welcome.

---
