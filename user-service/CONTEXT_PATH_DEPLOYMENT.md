# User Service Context Path Deployment Guide

## Overview

This guide covers the deployment and verification of the User Service with context path configuration for ALB path-based routing.

## Configuration Summary

### AWS Environment
- **Context Path**: `/api/users`
- **Spring Profile**: `aws`
- **Configuration File**: `src/main/resources/application-aws.yml`

### Local Environment
- **Context Path**: (none)
- **Spring Profile**: `local` (default)
- **Configuration File**: `src/main/resources/application.yml`

## Deployment Steps

### 1. Automatic Deployment (Recommended)

The CI/CD pipeline automatically deploys when:
- CI workflow completes successfully
- Changes are pushed to `deploy-dev` branch
- Manual workflow dispatch is triggered

**Trigger Manual Deployment:**
```bash
# Via GitHub UI:
# 1. Go to Actions tab
# 2. Select "CD - Dev User Service"
# 3. Click "Run workflow"
# 4. Select branch and provide reason (optional)
```

### 2. Manual Deployment (If needed)

If you need to deploy manually on EC2:

```bash
# SSH into EC2 instance
ssh ubuntu@<ec2-instance>

# Navigate to project directory
cd /home/ubuntu/wealist

# Pull latest code
git pull origin deploy-dev

# Load environment variables from Parameter Store
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=ap-northeast-2

# Pull latest image
aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin \
  ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

docker pull ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/wealist-dev-user-service:latest

# Restart service
docker compose -f docker/compose/docker-compose.ec2-dev.yml up -d --force-recreate user-service

# Check logs
docker compose -f docker/compose/docker-compose.ec2-dev.yml logs -f user-service
```

## Verification

### Automated Verification Script

Use the provided verification script:

```bash
# For AWS environment
./user-service/scripts/verify-context-path.sh aws

# For local environment
./user-service/scripts/verify-context-path.sh local
```

### Manual Verification

#### 1. Health Check

**AWS Environment:**
```bash
curl https://api.wealist.co.kr/api/users/actuator/health
```

**Local Environment:**
```bash
curl http://localhost:8080/actuator/health
```

Expected Response:
```json
{
  "status": "UP",
  "components": {
    "db": {"status": "UP"},
    "redis": {"status": "UP"}
  }
}
```

#### 2. Check Context Path in Logs

**AWS Environment:**
```bash
docker compose -f docker/compose/docker-compose.ec2-dev.yml logs user-service | grep "context-path"
```

Expected log output:
```
Tomcat initialized with port(s): 8080 (http)
Tomcat started on port(s): 8080 (http) with context path '/api/users'
```

**Local Environment:**
```bash
# Should show no context path or empty context path
docker compose logs user-service | grep "context-path"
```

#### 3. Test API Endpoints

**AWS Environment (through ALB):**
```bash
# The full path includes both ALB prefix and service path
curl https://api.wealist.co.kr/api/users/api/workspaces/all \
  -H "Authorization: Bearer <token>"
```

**Local Environment:**
```bash
# Direct access without ALB prefix
curl http://localhost:8080/api/workspaces/all \
  -H "Authorization: Bearer <token>"
```

#### 4. Verify Swagger UI

**AWS Environment:**
```bash
# Open in browser
https://api.wealist.co.kr/api/users/swagger-ui.html
```

**Local Environment:**
```bash
# Open in browser
http://localhost:8080/swagger-ui.html
```

## Troubleshooting

### Issue: 404 Not Found on Health Check

**Symptoms:**
```bash
curl https://api.wealist.co.kr/api/users/actuator/health
# Returns 404
```

**Possible Causes:**
1. Context path not configured correctly
2. Spring profile not activated
3. ALB routing rules not configured

**Solutions:**
1. Check environment variable:
   ```bash
   docker compose -f docker/compose/docker-compose.ec2-dev.yml exec user-service env | grep SPRING_PROFILES_ACTIVE
   # Should show: SPRING_PROFILES_ACTIVE=aws
   ```

2. Check application logs:
   ```bash
   docker compose -f docker/compose/docker-compose.ec2-dev.yml logs user-service | grep -i "context"
   ```

3. Verify configuration file exists:
   ```bash
   docker compose -f docker/compose/docker-compose.ec2-dev.yml exec user-service ls -la /app/BOOT-INF/classes/application-aws.yml
   ```

### Issue: Service Starts but Health Check Fails

**Symptoms:**
- Container is running
- Health check endpoint returns 503 or times out

**Solutions:**
1. Check database connectivity:
   ```bash
   docker compose -f docker/compose/docker-compose.ec2-dev.yml logs user-service | grep -i "database\|postgres"
   ```

2. Check Redis connectivity:
   ```bash
   docker compose -f docker/compose/docker-compose.ec2-dev.yml logs user-service | grep -i "redis"
   ```

3. Verify dependencies are healthy:
   ```bash
   docker compose -f docker/compose/docker-compose.ec2-dev.yml ps
   ```

### Issue: Context Path Applied in Local Environment

**Symptoms:**
- Local development shows context path `/api/users`
- Cannot access endpoints without prefix

**Solutions:**
1. Check Spring profile:
   ```bash
   # Should be 'local' or not set
   echo $SPRING_PROFILES_ACTIVE
   ```

2. Update docker-compose.yml for local:
   ```yaml
   environment:
     - SPRING_PROFILES_ACTIVE=local
   ```

## Rollback Procedure

If deployment fails or causes issues:

### 1. Quick Rollback (Remove Context Path)

```bash
# SSH to EC2
ssh ubuntu@<ec2-instance>

# Edit docker-compose file
cd /home/ubuntu/wealist
vi docker/compose/docker-compose.ec2-dev.yml

# Remove or comment out:
# - SPRING_PROFILES_ACTIVE=aws

# Restart service
docker compose -f docker/compose/docker-compose.ec2-dev.yml up -d --force-recreate user-service
```

### 2. Full Rollback (Previous Image)

```bash
# Pull previous image version
docker pull ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/wealist-dev-user-service:<previous-tag>

# Update docker-compose to use specific tag
# Then restart
docker compose -f docker/compose/docker-compose.ec2-dev.yml up -d --force-recreate user-service
```

## Next Steps

After successful deployment:

1. **Configure ALB** (Task 3.1):
   - Update Target Group health check path to `/api/users/actuator/health`
   - Add Listener Rule for `/api/users/*` path pattern

2. **Update Frontend** (if needed):
   - Ensure API calls use correct base URL
   - Update environment variables

3. **Monitor**:
   - Check CloudWatch logs
   - Monitor Target Group health
   - Review application metrics

## References

- [Requirements Document](../../.kiro/specs/alb-path-prefix-routing/requirements.md)
- [Design Document](../../.kiro/specs/alb-path-prefix-routing/design.md)
- [Tasks Document](../../.kiro/specs/alb-path-prefix-routing/tasks.md)
- [Spring Boot Context Path Documentation](https://docs.spring.io/spring-boot/docs/current/reference/html/web.html#web.servlet.embedded-container.context-path)
