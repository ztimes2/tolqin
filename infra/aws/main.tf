provider "aws" {
  region = var.region
}

resource "aws_default_vpc" "vpc" {}

resource "aws_security_group" "helloworld-sg" {
  name        = "helloworld-sg"
  description = "Allow inbound traffic"
  vpc_id      = aws_default_vpc.vpc.id

  ingress {
    description      = "Allow HTTPS"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Allow HTTP"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Allow SSH"
    from_port        = 22
    to_port          = 22
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
    Name    = "helloworld-sg"
    Project = "helloworld"
  }
}

resource "aws_instance" "helloworld-ec2" {
  ami               = "ami-02f26adf094f51167"
  instance_type     = "t2.micro"
  availability_zone = var.az_a
  key_name          = "helloworld-kp"

  vpc_security_group_ids = [aws_security_group.helloworld-sg.id]

  user_data = <<-EOF
    #!/bin/bash
    sudo yum update
    sudo yum -y install httpd
    sudo service httpd start
    sudo chown ec2-user /var/www/html
    sudo chmod -R o+r /var/www/html
    sudo echo 'Hello, World! 👋🌏' > /var/www/html/index.html
    EOF

  tags = {
    Name    = "helloworld-ec2"
    Project = "helloworld"
  }
}
