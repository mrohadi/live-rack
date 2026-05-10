output "vpc_id" {
  value = module.network.vpc_id
}

output "db_endpoint" {
  value     = module.db.endpoint
  sensitive = true
}

output "cache_endpoint" {
  value     = module.cache.endpoint
  sensitive = true
}

output "nats_endpoint" {
  value     = module.bus.nats_endpoint
  sensitive = true
}
