---
layout: "aws"
page_title: "AWS: aws_api_gateway_model"
sidebar_current: "docs-aws-resource-api-gateway-model"
description: |-
  Provides an HTTP Method for an API Gateway Resource.
---

# aws\_api\_gateway\_model

Provides an HTTP Method for an API Gateway Resource.

## Example Usage

```
resource "aws_api_gateway" "MyDemoAPI" {
  name = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_model" "MyDemoModel" {
  api_id = "${aws_api_gateway.MyDemoAPI.id}"
  name = "user"
  description = "a JSON schema"
  content_type = "application/json"
  schema = <<EOF
{
  "type": "object"
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) API Gateway ID
* `name` - (Required) Name of the model
* `description` - (Optional) Model description
* `content_type` - (Required) Model content type
* `schema` - (Required) Model schema
