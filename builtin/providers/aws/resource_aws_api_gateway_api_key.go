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

func resourceAwsApiGatewayApiKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayApiKeyCreate,
		Read:   resourceAwsApiGatewayApiKeyRead,
		Delete: resourceAwsApiGatewayApiKeyDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"stage_key": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"stage_name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsApiGatewayApiKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	// Create the gateway
	log.Printf("[DEBUG] Creating API Gateway")

	var stageKeys []*apigateway.StageKey
	if stageKeyData, ok := d.GetOk("stage_key"); ok {
		params := stageKeyData.(*schema.Set).List()
		for k := range params {
			data := params[k].(map[string]interface{})
			stageKeys = append(stageKeys, &apigateway.StageKey{
				RestApiId: aws.String(data["api_id"].(string)),
				StageName: aws.String(data["stage_name"].(string)),
			})
		}
	}

	var err error
	resp, err := conn.CreateApiKey(&apigateway.CreateApiKeyInput{
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Enabled:     aws.Bool(d.Get("enabled").(bool)),
		StageKeys:   stageKeys,
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway: %s", err)
	}

	// Get the ID and store it
	ig := *resp
	d.SetId(*ig.Id)
	log.Printf("[DEBUG] API Gateway ID: %s", d.Id())

	return resourceAwsApiGatewayApiKeyRefreshResources(d, meta)
}

func resourceAwsApiGatewayApiKeyRefreshResources(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	_, err := conn.GetApiKey(&apigateway.GetApiKeyInput{
		ApiKey: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	return nil
}

func resourceAwsApiGatewayApiKeyRead(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsApiGatewayApiKeyRefreshResources(d, meta)
}

func resourceAwsApiGatewayApiKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway: %s", d.Id())

	return resource.Retry(5*time.Minute, func() error {
		_, err := conn.DeleteApiKey(&apigateway.DeleteApiKeyInput{
			ApiKey: aws.String(d.Id()),
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
