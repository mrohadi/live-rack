output "nats_endpoint" {
  value = "nats.${var.name_prefix}.internal:4222"
}

output "ecs_cluster_arn" {
  value = aws_ecs_cluster.bus.arn
}
