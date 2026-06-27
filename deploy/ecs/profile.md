# ECS Deployment Profile

This profile provides an AWS ECS/Fargate baseline.

## Services

- `ticketstream-api` service behind ALB
- `ticketstream-worker` service (no public LB)

## Dependencies

- RDS PostgreSQL (Multi-AZ)
- ElastiCache Redis (Replication Group)
- Amazon MQ / RabbitMQ-compatible managed broker

## Runtime Configuration

- Inject all secrets from AWS Secrets Manager.
- Enable CloudWatch logs for both services.
- Use `/health/ready` for target group health checks.
- Enable autoscaling on CPU and request count.
