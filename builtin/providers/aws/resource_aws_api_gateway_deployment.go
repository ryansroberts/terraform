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

func resourceAwsApiGatewayDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayDeploymentCreate,
		Read:   resourceAwsApiGatewayDeploymentRead,
		Delete: resourceAwsApiGatewayDeploymentDelete,

		Schema: map[string]*schema.Schema{
			"api_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"stage_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"stage_description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"variables": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func resourceAwsApiGatewayDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	// Create the gateway
	log.Printf("[DEBUG] Creating API Gateway deployment")

	variables := make(map[string]string)
	for k, v := range d.Get("variables").(map[string]interface{}) {
		variables[k] = v.(string)
	}

	var err error
	resp, err := conn.CreateDeployment(&apigateway.CreateDeploymentInput{
		RestApiId:        aws.String(d.Get("api_id").(string)),
		StageName:        aws.String(d.Get("stage_name").(string)),
		Description:      aws.String(d.Get("description").(string)),
		StageDescription: aws.String(d.Get("stage_description").(string)),
		Variables:        aws.StringMap(variables),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway deployment: %s", err)
	}

	// Get the ID and store it
	ig := *resp
	d.SetId(*ig.Id)
	log.Printf("[DEBUG] API Gateway deployment ID: %s", d.Id())

	return nil
}

func resourceAwsApiGatewayDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway deployment %s", d.Id())
	out, err := conn.GetDeployment(&apigateway.GetDeploymentInput{
		RestApiId:    aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	})
	if err != nil {
		d.SetId("")
		return err
	}
	log.Printf("[DEBUG] Received API Gateway deployment: %s", out)
	d.SetId(*out.Id)

	return nil
}

func resourceAwsApiGatewayDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway deployment: %s", d.Id())

	return resource.Retry(5*time.Minute, func() error {
		log.Printf("[DEBUG] schema is %#v", d)
		if _, err := conn.DeleteStage(&apigateway.DeleteStageInput{
			StageName: aws.String(d.Get("stage_name").(string)),
			RestApiId: aws.String(d.Get("api_id").(string)),
		}); err == nil {
			return nil
		}

		_, err := conn.DeleteDeployment(&apigateway.DeleteDeploymentInput{
			DeploymentId: aws.String(d.Id()),
			RestApiId:    aws.String(d.Get("api_id").(string)),
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
