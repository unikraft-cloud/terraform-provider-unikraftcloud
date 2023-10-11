resource "kraftcloud_instance" "example" {
  image     = "unikraft.io/myuser.unikraft.io/myapp/latest"
  memory_mb = 64
  port      = 8080
  autostart = true
}
