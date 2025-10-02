# ARP Scanner Go Tool

This is a simple Go-based ARP scanner to detect devices on your local network.

## Requirements

1. Install `arp-scan`:

```bash
sudo apt update
sudo apt install arp-scan
```

2. Check the full path to `arp-scan`:

```bash
which arp-scan
```

3. Grant the necessary capability to run without `sudo`:

```bash
sudo setcap cap_net_raw+ep /usr/sbin/arp-scan
```

> Replace `/usr/sbin/arp-scan` with the path obtained from `which arp-scan` if different.

## Notes

- This is **not the final version**; it is a simple proof-of-concept.
- Make sure your network interface is correct when running the scanner.
