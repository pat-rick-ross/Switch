# 💡 ESP32-S3 Phone-Controlled Light Switch (TinyGo)

A fast, lightweight, and self-contained smart light switch built for the **ESP32-S3 DevKit** using **TinyGo**. 

Allows you to switch on, switch off, or toggle any light bulb directly from your smartphone browser—**no mobile apps or external home Wi-Fi routers required**. Also includes a photorealistic **interactive Web Simulator & Visualizer** for instant local testing.

---

## ✨ Features

- **Standalone Wi-Fi Hotspot (SoftAP):** The ESP32-S3 hosts its own secure Wi-Fi access point (`ESP32-LightBulb`). Works anywhere, off-grid or without internet.
- **Universal Mobile Compatibility:** Works seamlessly across iOS (Safari), Android (Chrome, Firefox), and desktops directly in the browser with **zero app downloads**.
- **Modern Mobile Web Interface:** Embedded responsive web controller featuring a glowing light bulb status icon, real-time state polling, and instant touch buttons.
- **RESTful API:** Lightweight HTTP endpoints (`/on`, `/off`, `/toggle`, `/status`) for easy integration with home automation scripts or shortcuts.
- **Interactive Web Simulator & Visualizer (`simulator.html`):** Test and visualize the entire circuit, glowing incandescent bulb, simulated GPIO logic, and procedural mechanical relay audio in your browser before flashing the hardware.
- **Configurable Relay Polarity:** Supports both Active-High and Active-Low 5V/3.3V relay modules and Solid State Relays (SSRs).

---

## 🔌 Hardware Requirements & Wiring

### Bill of Materials (BOM)
1. **ESP32-S3 DevKit Board** (e.g., ESP32-S3-DevKitC-1)
2. **5V or 3.3V 1-Channel Relay Module** or **Solid State Relay (SSR)**
3. **Light Bulb & Bulb Socket** (AC mains lamp or 12V DC bulb)
4. **Jumper Wires** & **USB-C Cable**

### Schematic Diagram

```
 ESP32-S3 DevKit                      Relay Module               Light Bulb Circuit
┌─────────────────┐                  ┌─────────────┐            ┌──────────────────┐
│             GND ├──────────────────┤ GND         │            │                  │
│             5V  ├──────────────────┤ VCC         │     ┌──────┤ Live / Hot (AC)  │
│                 │                  │             │     │      │                  │
│ GPIO 2 (Signal) ├──────────────────┤ IN (Signal) │     │      │   [Light Bulb]   │
└─────────────────┘                  │             │     │      │                  │
                                     │  COM  ──────┴─────┘      │                  │
                                     │  NO   ───────────────────┤ Neutral (AC Ret) │
                                     └─────────────┘            └──────────────────┘
```

> [!CAUTION]
> **High-Voltage Safety:** When working with 110V/220V AC mains electricity, disconnect all power sources before touching any wires. Enclose the relay and connections inside a non-conductive electrical junction box.

---

## 📁 Repository Structure

```
.
├── .gitignore         # Build and OS file exclusions
├── go.mod             # Go module definition & dependencies
├── go.sum             # Verified dependency checksums
├── README.md          # Project documentation
├── simulator.html     # Interactive web testing bench & visualizer
└── switch.go          # Main TinyGo firmware for ESP32-S3
```

---

## 🚀 Quick Start

### 1. Test in the Web Simulator (No Hardware Needed)

You can preview and test the complete system directly in your browser:

```bash
# Start a local HTTP server
python3 -m http.server 8000
```
Open [http://localhost:8000/simulator.html](http://localhost:8000/simulator.html) to interact with the virtual smartphone, watch the bulb illuminate with ambient room glow, and hear the mechanical relay click.

---

### 2. Configure the Firmware

Open [`switch.go`](switch.go) to verify or customize the hardware constants:

```go
// GPIO connected to the relay module signal input
const RelayPinNum = machine.GPIO2

// Set to true if your relay triggers on LOW (common for 5V relay modules)
const RelayActiveLow = false

// Access Point Settings
const (
    AP_SSID     = "ESP32-LightBulb"
    AP_PASSWORD = "lightswitch123" // 8 characters minimum (WPA2)
    AP_IP       = "192.168.4.1"
    HTTP_PORT   = 80
)
```

---

### 3. Flash to ESP32-S3

Prerequisites:
- [TinyGo](https://tinygo.org/getting-started/install/) (v0.32.0 or newer)
- Go (1.22+)

Compile and flash the firmware to your connected board:

```bash
# Build and flash to ESP32-S3
tinygo flash -target=esp32s3-generic switch.go

# (Optional) Open the serial monitor for diagnostics
tinygo monitor
```

---

### 4. Switch the Light from Your Phone

1. Power on your ESP32-S3.
2. On your phone, go to **Wi-Fi Settings** and connect to:
   - **SSID:** `ESP32-LightBulb`
   - **Password:** `lightswitch123`
3. Open Safari, Chrome, or any mobile browser and navigate to:
   - `http://192.168.4.1`
4. Tap **TURN ON** or **TURN OFF**. The relay switches the light bulb immediately!

---

## 🌐 REST API Endpoints

The embedded HTTP server exposes RESTful JSON endpoints that can be integrated with curl, Siri Shortcuts, or custom IoT scripts:

| Endpoint | Method | Description | Example Response |
| :--- | :--- | :--- | :--- |
| `/` | `GET` | Serves the interactive mobile web interface | `HTML page` |
| `/on` | `GET` | Turns the relay and light bulb **ON** | `{"state":1,"status":"on"}` |
| `/off` | `GET` | Turns the relay and light bulb **OFF** | `{"state":0,"status":"off"}` |
| `/toggle` | `GET` | Inverts the current state | `{"state":1,"status":"on"}` |
| `/status` | `GET` | Queries current light state | `{"state":1,"status":"on"}` |

---

## 🛠️ Tech Stack

- **Language:** Go / TinyGo
- **Target Hardware:** Espressif ESP32-S3 (Dual-core Xtensa LX7)
- **Networking Stack:** [`tinygo.org/x/espradio`](https://github.com/tinygo-org/espradio) (Native Wi-Fi SoftAP)
- **HTTP Server:** [`github.com/soypat/lneto`](https://github.com/soypat/lneto) (Lightweight embedded networking)
- **Web Frontend:** Pure HTML5, Modern Vanilla CSS, and Web Audio API

---

## 📄 License

This project is licensed under the MIT License - feel free to modify and use it for your own IoT projects!
