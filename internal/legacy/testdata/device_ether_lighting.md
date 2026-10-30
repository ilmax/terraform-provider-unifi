# Per-device EtherLighting (Network 11.0.97, USM8P emulated, which does NOT advertise the capability)

    PUT /proxy/network/api/s/default/rest/device/{_id}
    {"ether_lighting":{"led_mode":"etherlighting","mode":"speed","brightness":50,"behavior":"steady"}}

The controller stores it without the switch_caps bit: GET stat/device/{mac} returns the same object.
So the resource can be tested today. Allowed values: led_mode standard|etherlighting, mode speed|network,
brightness 1-100, behavior breath|steady. Site-level colors: setting key "ether_lighting".
