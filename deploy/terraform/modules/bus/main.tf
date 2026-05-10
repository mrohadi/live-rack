# NATS JetStream — deployed via ECS Fargate (single node for staging, cluster for prod).
# Swap for Synadia Cloud or NGS by pointing NATS_URL at the managed endpoint instead.

resource "aws_ecs_cluster" "bus" {
  name = "${var.name_prefix}-bus"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_cloudwatch_log_group" "nats" {
  name              = "/ecs/${var.name_prefix}/nats"
  retention_in_days = 30
}

resource "aws_security_group" "nats" {
  name        = "${var.name_prefix}-nats"
  description = "NATS JetStream ports"
  vpc_id      = var.vpc_id

  ingress {
    description = "Client connections"
    from_port   = 4222
    to_port     = 4222
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    description = "HTTP monitoring"
    from_port   = 8222
    to_port     = 8222
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.name_prefix}-sg-nats" }
}

resource "aws_efs_file_system" "nats" {
  encrypted = true
  tags      = { Name = "${var.name_prefix}-nats-data" }
}

resource "aws_efs_mount_target" "nats" {
  count = length(var.private_subnet_ids)

  file_system_id  = aws_efs_file_system.nats.id
  subnet_id       = var.private_subnet_ids[count.index]
  security_groups = [aws_security_group.nats.id]
}

resource "aws_iam_role" "nats_task" {
  name = "${var.name_prefix}-nats-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "nats_task_exec" {
  role       = aws_iam_role.nats_task.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_ecs_task_definition" "nats" {
  family                   = "${var.name_prefix}-nats"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "512"
  memory                   = "1024"
  execution_role_arn       = aws_iam_role.nats_task.arn

  volume {
    name = "nats-data"
    efs_volume_configuration {
      file_system_id = aws_efs_file_system.nats.id
      root_directory = "/"
    }
  }

  container_definitions = jsonencode([{
    name  = "nats"
    image = "nats:latest"
    command = ["-js", "-sd", "/data", "-m", "8222"]

    portMappings = [
      { containerPort = 4222, protocol = "tcp" },
      { containerPort = 8222, protocol = "tcp" },
    ]

    mountPoints = [{
      sourceVolume  = "nats-data"
      containerPath = "/data"
    }]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.nats.name
        "awslogs-region"        = data.aws_region.current.name
        "awslogs-stream-prefix" = "nats"
      }
    }
  }])
}

data "aws_region" "current" {}

resource "aws_service_discovery_private_dns_namespace" "bus" {
  name = "${var.name_prefix}.internal"
  vpc  = var.vpc_id
}

resource "aws_service_discovery_service" "nats" {
  name = "nats"
  dns_config {
    namespace_id   = aws_service_discovery_private_dns_namespace.bus.id
    routing_policy = "MULTIVALUE"
    dns_records {
      ttl  = 10
      type = "A"
    }
  }
  health_check_custom_config { failure_threshold = 1 }
}

resource "aws_ecs_service" "nats" {
  name            = "${var.name_prefix}-nats"
  cluster         = aws_ecs_cluster.bus.id
  task_definition = aws_ecs_task_definition.nats.arn
  desired_count   = var.env == "production" ? 3 : 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [aws_security_group.nats.id]
    assign_public_ip = false
  }

  service_registries {
    registry_arn = aws_service_discovery_service.nats.arn
  }

  lifecycle { ignore_changes = [desired_count] }
}
