package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayResourceCreate,
		Read:   resourceAwsApiGatewayResourceRead,
		Update: resourceAwsApiGatewayResourceUpdate,
		Delete: resourceAwsApiGatewayResourceDelete,

		Schema: map[string]*schema.Schema{
			"api_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"parent_resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"path_part": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"path": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsApiGatewayResourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	// Create the gateway
	log.Printf("[DEBUG] Creating API Gateway Resource")

	var err error
	resp, err := conn.CreateResource(&apigateway.CreateResourceInput{
		ParentId:  aws.String(d.Get("parent_resource_id").(string)),
		PathPart:  aws.String(d.Get("path_part").(string)),
		RestApiId: aws.String(d.Get("api_id").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error creating API Gateway Resource: %s", err)
	}

	// Get the ID and store it
	ig := *resp
	d.SetId(*ig.Id)
	d.Set("path", resp.Path)

	return nil
}

func resourceAwsApiGatewayResourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway Resource %s", d.Id())
	out, err := conn.GetResource(&apigateway.GetResourceInput{
		ResourceId: aws.String(d.Id()),
		RestApiId:  aws.String(d.Get("api_id").(string)),
	})

	if err == nil {
		d.Set("parent_resource_id", *out.ParentId)
		d.Set("path_part", *out.PathPart)
		return nil
	}

	apigatewayErr, _ := err.(awserr.Error)
	if apigatewayErr.Code() == "NotFoundException" {
		d.SetId("")
		return nil
	}

	return err
}

func resourceAwsApiGatewayResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway Resource %s", d.Id())
	operations := make([]*apigateway.PatchOperation, 0)
	if o, n := d.GetChange("path_part"); o.(string) != n.(string) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/pathPart"),
			Value: aws.String(d.Get("path_part").(string)),
		})
	}
	if o, n := d.GetChange("parent_resource_id"); o.(string) != n.(string) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/parentId"),
			Value: aws.String(d.Get("parent_resource_id").(string)),
		})
	}
	out, err := conn.UpdateResource(&apigateway.UpdateResourceInput{
		ResourceId:      aws.String(d.Id()),
		RestApiId:       aws.String(d.Get("api_id").(string)),
		PatchOperations: operations,
	})
	if err == nil {
		d.Set("path_part", *out.PathPart)
		d.Set("parent_resource_id", *out.ParentId)
		return resourceAwsApiGatewayResourceRead(d, meta)
	}
	apigatewayErr, _ := err.(awserr.Error)
	if apigatewayErr.Code() == "NotFoundException" {
		d.SetId("")
		return nil
	}
	return err
}

func resourceAwsApiGatewayResourceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway Resource: %s", d.Id())

	return resource.Retry(5*time.Minute, func() error {
		log.Printf("[DEBUG] schema is %#v", d)
		_, err := conn.DeleteResource(&apigateway.DeleteResourceInput{
			ResourceId: aws.String(d.Id()),
			RestApiId:  aws.String(d.Get("api_id").(string)),
		})
		if err == nil {
			return nil
		}

		apigatewayErr, ok := err.(awserr.Error)
		if apigatewayErr.Code() == "NotFoundException" {
			return nil
		}

		if !ok {
			return resource.RetryError{Err: err}
		}

		return resource.RetryError{Err: err}
	})
}
