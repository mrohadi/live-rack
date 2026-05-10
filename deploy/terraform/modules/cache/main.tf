resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.name_prefix}-cache-subnet"
  subnet_ids = var.private_subnet_ids
}

resource "aws_security_group" "cache" {
  name        = "${var.name_prefix}-cache"
  description = "Redis access from private subnets"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = { Name = "${var.name_prefix}-sg-cache" }
}

resource "aws_elasticache_replication_group" "main" {
  replication_group_id = "${var.name_prefix}-redis"
  description          = "live-rack Redis cache"

  node_type            = var.cache_node_type
  num_cache_clusters   = 2
  engine_version       = "7.1"
  port                 = 6379

  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [aws_security_group.cache.id]

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = random_password.cache_auth.result

  automatic_failover_enabled = true
  multi_az_enabled           = true

  snapshot_retention_limit = 1
  snapshot_window          = "03:30-04:30"

  tags = { Name = "${var.name_prefix}-redis" }
}

resource "random_password" "cache_auth" {
  length  = 32
  special = false
}

resource "aws_secretsmanager_secret" "cache" {
  name                    = "${var.name_prefix}/cache/credentials"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret_version" "cache" {
  secret_id     = aws_secretsmanager_secret.cache.id
  secret_string = jsonencode({ auth_token = random_password.cache_auth.result })
}
