# Grafana Cloud or self-hosted Grafana Stack (Loki, Tempo, Prometheus).
# This module provisions the SNS alarm topic + basic CloudWatch alarms.
# Grafana agent runs as a sidecar on each ECS task (configured separately).

resource "aws_sns_topic" "alerts" {
  name = "${var.name_prefix}-alerts"
}

resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/live-rack/${var.env}/app"
  retention_in_days = var.env == "production" ? 90 : 14
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "/live-rack/${var.env}/api"
  retention_in_days = var.env == "production" ? 90 : 14
}

resource "aws_cloudwatch_metric_alarm" "db_cpu" {
  alarm_name          = "${var.name_prefix}-db-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = 120
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "RDS CPU > 80% for 4 minutes"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "db_connections" {
  alarm_name          = "${var.name_prefix}-db-connections-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = 60
  statistic           = "Maximum"
  threshold           = 400
  alarm_description   = "RDS connections > 400"
  alarm_actions       = [aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "cache_memory" {
  alarm_name          = "${var.name_prefix}-cache-memory-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "DatabaseMemoryUsagePercentage"
  namespace           = "AWS/ElastiCache"
  period              = 120
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "Redis memory > 80%"
  alarm_actions       = [aws_sns_topic.alerts.arn]
}

# Grafana workspace (Amazon Managed Grafana) — optional, swap for self-hosted.
resource "aws_grafana_workspace" "main" {
  count = var.env == "production" ? 1 : 0

  name                     = "${var.name_prefix}-grafana"
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["AWS_SSO"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.grafana[0].arn

  data_sources             = ["CLOUDWATCH", "PROMETHEUS", "XRAY"]
  notification_destinations = ["SNS"]
}

resource "aws_iam_role" "grafana" {
  count = var.env == "production" ? 1 : 0
  name  = "${var.name_prefix}-grafana"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "grafana.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "grafana_cw" {
  count      = var.env == "production" ? 1 : 0
  role       = aws_iam_role.grafana[0].name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess"
}

# Amazon Managed Service for Prometheus workspace.
resource "aws_prometheus_workspace" "main" {
  alias = "${var.name_prefix}-amp"
}
