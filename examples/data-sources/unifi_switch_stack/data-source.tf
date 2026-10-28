data "unifi_switch_stack" "example" {
  site_id         = var.unifi_site_id
  switch_stack_id = "stack-123"
}
