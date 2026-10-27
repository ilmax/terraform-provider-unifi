data "unifi_countries" "all" {}

locals {
  china_code = data.unifi_countries.all.countries["China"].code
}
