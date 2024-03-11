ui = true
disable_mlock = "true"

listener "tcp" {
  address = "[::]:8200"
  tls_disable = true
}

api_addr = "http://0.0.0.0:8200"