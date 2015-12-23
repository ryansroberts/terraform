---
layout: "aws"
page_title: "AWS: aws_api_gateway_deployment"
sidebar_current: "docs-aws-resource-api-gateway-deployment"
description: |-
  Provides an API Gateway Deployment.
---

# aws\_api\_gateway\_deployment

Provides an API Gateway Deployment.

## Example Usage

```
resource "aws_api_gateway" "MyDemoAPI" {
  name = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_deployment" "MyDemoDeployment" {
  api_id = "${aws_api_gateway.MyDemoAPI.id}"

  variables = {
    "answer" = "42"
  }
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) ID of the API Gateway
* `stage_name` - (Required) name of the stage
* `description` - (Optional) name of the stage
* `variables` - (Optional) Stage Variables
