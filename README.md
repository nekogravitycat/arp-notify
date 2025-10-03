# ARP Notify

A small Go tool that monitors devices on your local Wi-Fi/LAN by scanning for their MAC addresses using `arp-scan`.
When a target device is detected, it sends a notification via LINE Bot.

## Requirements

1. Install `arp-scan`:

```bash
sudo apt update
sudo apt install arp-scan
````

2. Grant the required capability so `arp-scan` can run without `sudo`:

```bash
sudo setcap cap_net_raw+ep /usr/sbin/arp-scan
```

## Build

```bash
go build -o arp-notify ./cmd/arp-notify/main.go
```

## Configuration

### Target Monitoring File

Create a file named `monitor_targets.json`. Example:

```json
{
  "targets": [
    {
      "mac": "e0:0f:52:1b:b9:59",
      "message": "A random mac address and a random receiver!",
      "receivers": [
        "Uufj4b2qnpmf3jj0pqj8xqz42ay1bbo8s"
      ]
    },
    {
      "mac": "e0:0f:52:1b:b9:59",
      "ip": "192.168.0.2",
      "message": "With optional IP address.",
      "receivers": [
        "Uufj4b2qnpmf3jj0pqj8xqz42ay1bbo8s"
      ]
    }
  ]
}
```

* `mac`: The MAC address of the device to monitor.
* `ip`: Optional IP address to probe when a broadcast ARP scan fails.
* `message`: The notification message to send when the device is detected.
* `receivers`: A list of LINE user IDs to receive the message.

### Environment Variables (`.env`)

Create a `.env` file with the following:

#### Required

* `LINE_BOT_CHANNEL_ACCESS_TOKEN`
* `LINE_BOT_CHANNEL_SECRET`

#### Optional (with defaults)

* `ARP_SCAN_BIN = "arp-scan"`
* `ARP_SCAN_IFACE = ""`
* `ARP_SCAN_INTERVAL_SECS = "60"`
* `ARP_SCAN_TIMEOUT_SECS = "15"`
* `MONITOR_ABSENCE_RESET_MIN = "1440"`


## Run

```bash
./arp-notify
```

The program will periodically run `arp-scan` to detect target devices and send LINE Bot notifications when matches are found.
