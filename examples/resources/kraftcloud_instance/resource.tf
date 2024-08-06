resource "unikraft-cloud_instance" "example" {
  image     = "myuser.unikraft.io/myapp:latest"
  memory_mb = 64
  autostart = true
  service_group = {
    services = [
      {
        port    = 80
        handler = ["http"]
      }
    ]
  }
}
