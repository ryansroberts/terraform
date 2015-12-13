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

func resourceAwsApiGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayCreate,
		Read:   resourceAwsApiGatewayRead,
		Update: resourceAwsApiGatewayUpdate,
		Delete: resourceAwsApiGatewayDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"root_resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsApiGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	// Create the gateway
	log.Printf("[DEBUG] Creating API Gateway")

	var err error
	resp, err := conn.CreateRestApi(&apigateway.CreateRestApiInput{
		Name: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway: %s", err)
	}

	// Get the ID and store it
	ig := *resp
	d.SetId(*ig.Id)
	log.Printf("[DEBUG] API Gateway ID: %s", d.Id())

	return resourceAwsApiGatewayRefreshResources(d, meta)
}

func resourceAwsApiGatewayRefreshResources(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	resp, err := conn.GetResources(&apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	for _, item := range resp.Items {
		if *item.Path == "/" {
			d.Set("root_resource_id", item.Id)
			break
		}
	}

	return nil
}

func resourceAwsApiGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway %s", d.Id())
	out, err := conn.GetRestApi(&apigateway.GetRestApiInput{
		RestApiId: aws.String(d.Id()),
	})
	if err != nil {
		d.SetId("")
		return err
	}
	log.Printf("[DEBUG] Received API Gateway: %s", out)

	d.SetId(*out.Id)
	d.Set("name", *out.Name)
	if out.Description != nil {
		d.Set("description", *out.Description)
	}

	return nil
}

func resourceAwsApiGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Updating API Gateway %s", d.Id())
	operations := make([]*apigateway.PatchOperation, 0)
	if o, n := d.GetChange("name"); o.(string) != n.(string) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}
	if o, n := d.GetChange("description"); o.(string) != n.(string) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}
	_, err := conn.UpdateRestApi(&apigateway.UpdateRestApiInput{
		RestApiId:       aws.String(d.Id()),
		PatchOperations: operations,
	})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Updated API Gateway %s", d.Id())

	return resourceAwsApiGatewayRead(d, meta)
}

func resourceAwsApiGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway: %s", d.Id())

	return resource.Retry(5*time.Minute, func() error {
		_, err := conn.DeleteRestApi(&apigateway.DeleteRestApiInput{
			RestApiId: aws.String(d.Id()),
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
