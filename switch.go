package main

import (
	"machine"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/soypat/lneto/http/httphi"
	"tinygo.org/x/espradio"
	link "tinygo.org/x/espradio/netlink"
)

// Hardware configuration
// Change RELAY_PIN to match the GPIO connected to your relay module or transistor.
// Common usable pins on ESP32-S3: GPIO2, GPIO4, GPIO5, GPIO6, GPIO7, GPIO10, GPIO15.
const (
	RelayPinNum = machine.GPIO2

	// Set to true if your relay module turns ON with LOW signal (active-low),
	// or false if it turns ON with HIGH signal (active-high).
	RelayActiveLow = false
)

// Wi-Fi Access Point credentials
// Note: If a password is used, WPA2 requires at least 8 characters.
const (
	AP_SSID     = "ESP32-LightBulb"
	AP_PASSWORD = "lightswitch123"
	AP_CHANNEL  = 6
	AP_IP       = "192.168.4.1"
	HTTP_PORT   = 80
)

// Global light bulb state: 0 = OFF, 1 = ON
var lightState atomic.Int32

// HTML & CSS Web App served to connected phones
const webUI = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>Smart Light Switch</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: radial-gradient(circle at 50% 20%, #1f2937, #0b0f19);
      color: #f3f4f6;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 24px;
      touch-action: manipulation;
    }
    .card {
      background: rgba(31, 41, 55, 0.7);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 28px;
      padding: 36px 28px;
      width: 100%;
      max-width: 360px;
      text-align: center;
      box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
    }
    h1 {
      font-size: 1.4rem;
      font-weight: 700;
      letter-spacing: -0.02em;
      margin-bottom: 6px;
      color: #f9fafb;
    }
    .device-info {
      font-size: 0.8rem;
      color: #9ca3af;
      margin-bottom: 32px;
      letter-spacing: 0.05em;
      text-transform: uppercase;
    }
    .bulb-container {
      position: relative;
      width: 140px;
      height: 140px;
      margin: 0 auto 32px;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.03);
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.4s ease;
    }
    .bulb-container.on {
      background: rgba(250, 204, 21, 0.15);
      box-shadow: 0 0 60px rgba(250, 204, 21, 0.5), inset 0 0 20px rgba(250, 204, 21, 0.3);
    }
    .bulb-icon {
      width: 76px;
      height: 76px;
      fill: none;
      stroke: #6b7280;
      stroke-width: 1.8;
      stroke-linecap: round;
      stroke-linejoin: round;
      transition: all 0.4s ease;
    }
    .bulb-container.on .bulb-icon {
      stroke: #facc15;
      filter: drop-shadow(0 0 12px #facc15);
    }
    .status-badge {
      display: inline-block;
      font-size: 0.95rem;
      font-weight: 600;
      padding: 6px 18px;
      border-radius: 9999px;
      margin-bottom: 32px;
      background: rgba(239, 68, 68, 0.15);
      color: #f87171;
      border: 1px solid rgba(239, 68, 68, 0.3);
      transition: all 0.3s ease;
    }
    .status-badge.on {
      background: rgba(34, 197, 94, 0.15);
      color: #4ade80;
      border-color: rgba(34, 197, 94, 0.3);
    }
    .btn-group {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 14px;
    }
    button {
      appearance: none;
      border: none;
      outline: none;
      padding: 16px;
      font-size: 1rem;
      font-weight: 600;
      border-radius: 18px;
      cursor: pointer;
      transition: transform 0.1s ease, filter 0.2s ease, background 0.2s ease;
      user-select: none;
      -webkit-tap-highlight-color: transparent;
    }
    button:active {
      transform: scale(0.96);
    }
    .btn-on {
      background: linear-gradient(135deg, #eab308, #ca8a04);
      color: #111827;
      box-shadow: 0 8px 20px rgba(234, 179, 8, 0.3);
    }
    .btn-off {
      background: #374151;
      color: #d1d5db;
      border: 1px solid rgba(255, 255, 255, 0.08);
    }
    .footer {
      margin-top: 28px;
      font-size: 0.75rem;
      color: #6b7280;
    }
  </style>
</head>
<body>
  <div class="card">
    <h1>Light Bulb Control</h1>
    <div class="device-info">ESP32-S3 • TinyGo Switch</div>

    <div id="bulb" class="bulb-container">
      <svg class="bulb-icon" viewBox="0 0 24 24">
        <path d="M9 18h6m-4 3h2m-4-7.5a6 6 0 1 1 6 0c0 2-2 3.5-2 4.5h-2c0-1-2-2.5-2-4.5z"/>
      </svg>
    </div>

    <div id="status" class="status-badge">LIGHT OFF</div>

    <div class="btn-group">
      <button class="btn-on" onclick="setBulb('on')">TURN ON</button>
      <button class="btn-off" onclick="setBulb('off')">TURN OFF</button>
    </div>

    <div class="footer">Connect to ESP32-LightBulb Wi-Fi</div>
  </div>

  <script>
    function updateUI(isOn) {
      const bulb = document.getElementById('bulb');
      const status = document.getElementById('status');
      if (isOn) {
        bulb.classList.add('on');
        status.classList.add('on');
        status.textContent = 'LIGHT ON';
      } else {
        bulb.classList.remove('on');
        status.classList.remove('on');
        status.textContent = 'LIGHT OFF';
      }
    }

    function setBulb(action) {
      fetch('/' + action)
        .then(res => res.json())
        .then(data => updateUI(data.state === 1))
        .catch(err => console.error(err));
    }

    // Periodically poll status
    function syncStatus() {
      fetch('/status')
        .then(res => res.json())
        .then(data => updateUI(data.state === 1))
        .catch(() => {});
    }
    setInterval(syncStatus, 3000);
    syncStatus();
  </script>
</body>
</html>`

// applyRelayState updates the physical GPIO pin according to active-high / active-low logic.
func applyRelayState(on bool) {
	if on {
		lightState.Store(1)
		if RelayActiveLow {
			RelayPinNum.Low()
		} else {
			RelayPinNum.High()
		}
		println("[SWITCH] Light turned ON")
	} else {
		lightState.Store(0)
		if RelayActiveLow {
			RelayPinNum.High()
		} else {
			RelayPinNum.Low()
		}
		println("[SWITCH] Light turned OFF")
	}
}

func main() {
	// Give serial monitor a moment to attach
	time.Sleep(2 * time.Second)
	println("\n========================================")
	println("  ESP32-S3 Phone Light Switch (TinyGo)")
	println("========================================")

	// Configure the relay output pin
	RelayPinNum.Configure(machine.PinConfig{Mode: machine.PinOutput})
	applyRelayState(false) // Start in OFF state

	// Initialize Wi-Fi stack
	lnk := &link.Esplink{}

	println("[WIFI] Starting Access Point: " + AP_SSID + "...")
	err := lnk.NetConnectAP(link.APConnectParams{
		APConfig: espradio.APConfig{
			SSID:     AP_SSID,
			Password: AP_PASSWORD,
			Channel:  AP_CHANNEL,
		},
		StaticAddr:       netip.MustParseAddr(AP_IP),
		EnableDHCPServer: true,
		MaxUDPPorts:      2,
		MaxTCPPorts:      4,
		PassivePeers:     64,
	})
	if err != nil {
		println("[ERROR] Failed to start Access Point:", err.Error())
		for {
			time.Sleep(1 * time.Second)
		}
	}

	println("[WIFI] Access Point ready!")
	println(" -> Connect your phone to Wi-Fi: " + AP_SSID)
	println(" -> Password: " + AP_PASSWORD)
	println(" -> Open browser at: http://" + AP_IP)

	// Set up HTTP Request Handlers
	var http httphi.MuxSlice

	// Serve the interactive web page
	http.Handle("/", func(exch *httphi.Exchange) {
		exch.RespondString(httphi.StatusOK, "text/html; charset=UTF-8", webUI)
	})

	// API to turn Light ON
	http.Handle("/on", func(exch *httphi.Exchange) {
		applyRelayState(true)
		exch.RespondString(httphi.StatusOK, "application/json", `{"state":1,"status":"on"}`)
	})

	// API to turn Light OFF
	http.Handle("/off", func(exch *httphi.Exchange) {
		applyRelayState(false)
		exch.RespondString(httphi.StatusOK, "application/json", `{"state":0,"status":"off"}`)
	})

	// API to toggle Light
	http.Handle("/toggle", func(exch *httphi.Exchange) {
		currentState := lightState.Load() == 1
		applyRelayState(!currentState)
		if !currentState {
			exch.RespondString(httphi.StatusOK, "application/json", `{"state":1,"status":"on"}`)
		} else {
			exch.RespondString(httphi.StatusOK, "application/json", `{"state":0,"status":"off"}`)
		}
	})

	// API to read current status
	http.Handle("/status", func(exch *httphi.Exchange) {
		if lightState.Load() == 1 {
			exch.RespondString(httphi.StatusOK, "application/json", `{"state":1,"status":"on"}`)
		} else {
			exch.RespondString(httphi.StatusOK, "application/json", `{"state":0,"status":"off"}`)
		}
	})

	// Configure HTTP router
	const maxConns, httpMemoryPerConn = 4, 2048
	var router httphi.Router
	routerCfg := httphi.DefaultRouterConfig(maxConns, httpMemoryPerConn, http.MaxPathValues())
	if err := router.Configure(&http, routerCfg); err != nil {
		println("[ERROR] Failed to configure router:", err.Error())
		return
	}
	defer router.Shutdown()

	println("[HTTP] Server listening on port", HTTP_PORT)

	// Start listening for incoming phone connections
	err = lnk.ListenAndServe(&router, HTTP_PORT)
	if err != nil {
		println("[ERROR] ListenAndServe:", err.Error())
	}
}
