# ARP Notify

A small Go tool that monitors devices on your local Wi-Fi/LAN by scanning for their MAC addresses using `arp-scan`.
When a target device is detected, it sends a notification via LINE Bot.

## Requirements

1. Install `arp-scan`:

```bash
sudo apt update
sudo apt install arp-scan
```

2. Grant the required capability so `arp-scan` can run without `sudo`:

```bash
sudo setcap cap_net_raw+ep /usr/sbin/arp-scan
```

## Build

```bash
go build -o arp-notify ./cmd/arp-notify
```

## Configuration

Configuration lives in two YAML files in the working directory. They are created
automatically from a template on first run — populate them (or use the web UI) and restart.
Only the LINE secrets live in `.env`.

### `.env` (secrets only)

```dotenv
LINE_BOT_CHANNEL_ACCESS_TOKEN="..."
LINE_BOT_CHANNEL_SECRET="..."
```

These may also be supplied directly by the environment (e.g. via systemd); the `.env` file is
optional.

### `config.yaml` (system / scan behavior)

```yaml
arp_scan:
  bin: arp-scan            # path to the arp-scan binary
  iface: ""               # network interface; empty = all interfaces
  interval_sec: 60         # how often to scan
  broadcast_timeout_sec: 15
  individual_timeout_sec: 2
monitor:
  absence_reset_min: 1440  # re-notify after a device has been absent this long (minutes)
server:
  host: "127.0.0.1"        # bind address; 127.0.0.1 = loopback only, 0.0.0.0 = all interfaces
  port: 5000               # HTTP port for the LINE webhook and the /admin UI
```

### `targets.yaml` (what to watch + who to tell)

```yaml
default_message: "Welcome home!"     # used when a target/receiver has no message
contacts:                            # reusable LINE user -> friendly name registry
  - id: "Uufj4b2qnpmf3jj0pqj8xqz42ay1bbo8s"
    name: "Mom"
targets:
  - name: "Mom's phone"              # friendly label (UI + logs)
    mac: "e0:0f:52:1b:b9:59"
    enabled: true
    detection:
      mode: auto                     # ip | broadcast | auto
      ip: "192.168.0.2"              # required for ip / auto
    message: "Mom's home!"           # optional; overrides default_message
    receivers:
      - id: "Uufj4b2qnpmf3jj0pqj8xqz42ay1bbo8s"
        message: "歡迎回家！"         # optional; overrides the target message
```

- **Detection modes**
  - `ip` — only individual-scan the configured IP.
  - `broadcast` — only broadcast-scan and match the MAC in the output.
  - `auto` — individual-scan the IP first, fall back to the broadcast scan.
- **Message precedence:** `receiver.message` → `target.message` → `default_message`.
- **Contacts** map a LINE user ID to a friendly name and are auto-filled from the LINE profile
  the first time that user messages the bot.

## Web UI

The service serves a configuration UI at `http://<host>:<port>/admin/` (default
`http://localhost:5000/admin/`) on the same port as the LINE webhook. From there you can:

- edit targets, detection modes, messages and receivers;
- pick receivers from the list of users who recently messaged the bot (with their LINE names);
- send a test notification to verify a receiver ID;
- view live device status (last seen / notified);
- adjust system settings.

Changes are saved to the YAML files and take effect **immediately, without a restart** (a
changed `server.host`/`server.port` is the one exception and needs a restart). The UI has **no
authentication**, so it binds to `127.0.0.1` (loopback) by default. Set `server.host` to
`0.0.0.0` only on a trusted network if you need to reach it from another machine.

### Finding a LINE user ID

Have the person send the bot any message; they will appear in the **"Pick from recent"** picker.
Sending `whoami` makes the bot reply with the raw user ID.

## Run

```bash
./arp-notify
```

The program periodically runs `arp-scan` to detect target devices and sends LINE Bot
notifications when matches are found.

## Autostart on Linux (systemd)

A systemd service file `arp-notify.service` is included in the repository. To use it:

1.  **Edit the service file**:
    Open `arp-notify.service` and replace `<YOUR_USER>` with your username and `<PATH_TO_PROJECT>` with the absolute path to your `arp-notify` directory.
2.  **Install the service**:
    ```bash
    sudo cp arp-notify.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl enable --now arp-notify
    ```
3.  **Check status**:
    ```bash
    sudo systemctl status arp-notify
    ```

## Updating

To deploy a newer version when running under systemd:

1.  **Stop the service**:
    ```bash
    sudo systemctl stop arp-notify
    ```
2.  **Pull and rebuild** in the project directory:
    ```bash
    cd /path/to/arp-notify
    git pull
    go build -o arp-notify ./cmd/arp-notify
    ```
3.  **Restart and verify**:
    ```bash
    sudo systemctl start arp-notify
    sudo systemctl status arp-notify
    journalctl -u arp-notify -f
    ```

Your `.env`, `config.yaml` and `targets.yaml` are left untouched. If you build on a
different machine, copy the resulting `arp-notify` binary over the old one and restart
the service instead of running `git pull && go build` on the server.
