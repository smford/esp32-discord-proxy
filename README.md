# esp32-discord-proxy

A simple app which proxies between webapi and discord, and vice versa.

The problem this solves is that discord uses web sockets, and that controlling a device like an ESP32 would require a constant connection between the device and discord.

Instead, the device will call a webhook to esp32-discord-proxy, which then processes the request and passes it on to discord.  Alternatively, inbound data flows will work like this:  Discord, via websockets, will notify esp32-discord-proxy, who will process it and then call webapis presented on the device.

Outbound: `device state change -> web api call to esp32-discord-proxy -> discord message`

Inbound: `a message in discord is typed by a user -> esp32-discord-proxy detects and processes message -> web api call to device`

This tool works in conjunction with: https://github.com/smford/eeh-esp32-rfid

## Emoji Sources
- 32on - https://iconscout.com/icon/3d-89
- 3doff - https://iconscout.com/icon/3d-89
- robot icon - http://www.pngall.com/robot-png/download/22191
- backlight off - https://www.iconspng.com/image/113092/lightbulb-onoff-2
- backlight on - https://www.iconspng.com/image/113091/lightbulb-onoff-1
- laseron - https://www.cleanpng.com/png-laser-engraving-laser-cutting-laser-icon-5145931/download-png.html
- lasermaintenance - https://iconscout.com/icon/maintenance-26
- ticket - https://www.iconfinder.com/icons/1891021/approved_check_checkbox_checkmark_confirm_success_yes_icon
- userlogin - https://www.flaticon.com/free-icon/enter_1391041
- userlogout - https://www.flaticon.com/free-icon/log-out_1391043
