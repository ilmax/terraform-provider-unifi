# WireGuard VPN server (Network 11.0.97, probed through the API, UI picker unavailable)

    POST /proxy/network/api/s/default/rest/networkconf
    {"name":"probe-wg","purpose":"remote-user-vpn","vpn_type":"wireguard-server","enabled":true,
     "local_port":51820,"ip_subnet":"192.168.77.1/24","wireguard_interface":"wan",
     "x_wireguard_private_key":"<base64 32-byte key>"}

Validation errors found by iterating (each missing field is reported one at a time):
api.err.MissingLocalPort -> api.err.IncorrectIPSubnetSpec -> api.err.WireguardMissingPrivateKey.

Stored document adds: _id, external_id, site_id, wireguard_id (int), firewall_zone_id (the VPN zone),
vlan_enabled=false, dhcpd_dns_enabled=false, igmp_snooping=false, mdns_enabled=false, dpi_enabled=false.
The public key is derived by the controller. Field names/allowed values: controller jar api/fields/NetworkConf.json.
Delete: DELETE /api/s/default/rest/networkconf/{_id}.
