# Switch port PoE (captured from UI, Network 11.0.97, USW Ultra emulated)

UI action: Ports > port 2 > PoE (Off / PoE+ / 24V Passive) > Apply Changes.

Request observed (URL + method via network log; body not visible to the tool, shape taken from the
resulting device state):

    PUT /proxy/network/api/s/default/rest/device/{device _id}

Resulting `port_overrides` on `GET /api/s/default/stat/device/{mac}`:

    [{"port_idx":2,"name":"PoE Out + Data","poe_mode":"off","forward":"all","speed":1000,
      "full_duplex":true,"autoneg":false,"tagged_vlan_mgmt":"auto","setting_preference":"auto",
      "port_security_enabled":false,"port_security_mac_address":[],"ld_status":false,
      "sd_wan_underlay_port":false,"native_networkconf_id":"<network uuid>"}]

poe_mode values seen: "auto" (PoE+), "off". 24V passive not yet captured ("pasv24" in other firmwares).
port_overrides is the whole array: PUT must send every override, so read-merge-write.
Also seen on page load: GET /v2/api/site/default/poe-info?mac=... (returned {} on the emulator).
