# Teleport (Network 11.0.97, gateway adopted: emulated UXGENT)

Per-site setting `{enabled, subnet_cidr}` (the UI's Teleport row is this setting; it is NOT the many-field
VPN server form, see wireguard_server.md). Never send subnet_cidr "": the controller stores it and then fails
to list networks (HTTP 500). The key disappears on a disabled -> enabled transition.

    GET /proxy/network/api/s/default/rest/setting          -> entry with key "teleport"
    PUT /proxy/network/api/s/default/rest/setting/teleport/{_id}   {"enabled":false}
    -> {"meta":{"rc":"ok"},"data":[{"_id":"...","key":"teleport","enabled":false,"site_id":"..."}]}

The UI has no Teleport form beyond "Generate New Guest Invite", which POSTs
/proxy/network/v2/api/site/default/teleport/invitation-history (HTTP 501 on the emulated console).
