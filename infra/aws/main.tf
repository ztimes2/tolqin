provider "aws" {
  region = var.region
}

resource "aws_default_vpc" "main" {}

resource "aws_security_group" "lb" {
  name        = "lb-helloworld"
  description = "Allow inbound HTTP traffic"
  vpc_id      = aws_default_vpc.main.id

  ingress {
    description      = "Allow HTTP"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name    = "sg-lb-helloworld"
    Project = "helloworld"
  }
}

resource "aws_lb" "main" {
  name               = "helloworld"
  internal           = false
  load_balancer_type = "application" 
  security_groups    = [aws_security_group.lb.id]
  subnets            = [var.subnet_a_id, var.subnet_b_id]
 
  enable_deletion_protection = false

  tags = {
    Name    = "lb-helloworld"
    Project = "helloworld"
  }
}

resource "aws_alb_target_group" "main" {
  name        = "helloworld"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_default_vpc.main.id
  target_type = "ip"
 
  health_check {
    healthy_threshold   = "3"
    interval            = "30"
    protocol            = "HTTP"
    matcher             = "200"
    timeout             = "3"
    path                = "/"
    unhealthy_threshold = "2"
  }

  tags = {
    Name    = "lb-target-group-helloworld"
    Project = "helloworld"
  }
}

resource "aws_alb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"
 
  default_action {
    target_group_arn = aws_alb_target_group.main.arn
    type             = "forward"
  }
 
  tags = {
    Name    = "lb-listener-helloworld"
    Project = "helloworld"
  }
}

resource "aws_security_group" "ecs_task" {
  name        = "ecs-task-helloworld"
  description = "Allow inbound HTTP traffic from ELB"
  vpc_id      = aws_default_vpc.main.id

  ingress {
    description      = "Allow HTTP from ELB"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    security_groups = [ aws_security_group.lb.id ]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name    = "sg-ecs-task-helloworld"
    Project = "helloworld"
  }
}

resource "aws_ecr_repository" "main" {
  name                 = "helloworld"
  image_tag_mutability = "MUTABLE"

  tags = {
    Name    = "ecr-helloworld"
    Project = "helloworld"
  }
}

resource "aws_ecs_cluster" "main" {
  name = "helloworld"

  tags = {
    Name    = "ecs-cluster-helloworld"
    Project = "helloworld"
  }
}

resource "aws_iam_role" "ecs_task_execution" {
  name = "ecs-task-execution-helloworld"
 
  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Action": "sts:AssumeRole",
     "Principal": {
       "Service": "ecs-tasks.amazonaws.com"
     },
     "Effect": "Allow",
     "Sid": ""
   }
 ]
}
EOF

  tags = {
    Name    = "ecs-task-execution-role-helloworld"
    Project = "helloworld"
  }
}
 
resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_ecs_task_definition" "main" {
  family = "helloworld"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn = aws_iam_role.ecs_task_execution.arn 

  container_definitions = jsonencode([{
    name        = "helloworld"
    image       = aws_ecr_repository.main.repository_url
    essential   = true
    portMappings = [{
      protocol      = "tcp"
      containerPort = 80
      hostPort      = 80
    }]
  }])
  
  tags = {
    Name    = "ecs-task-definition-helloworld"
    Project = "helloworld"
  }
}

resource "aws_ecs_service" "main" {
  name = "helloworld"
  cluster = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count = 1
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"

  network_configuration {
    security_groups  = [aws_security_group.ecs_task.id]
    subnets          = [var.subnet_a_id, var.subnet_b_id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_alb_target_group.main.arn
    container_name   = "helloworld"
    container_port   = "80"
  }

  lifecycle {
   ignore_changes = [task_definition, desired_count]
  }
}

