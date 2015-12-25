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
		Update: resourceAwsApiGatewayApiKeyUpdate,
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
			},

			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"stage_key": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
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

func resourceAwsApiGatewayStageKeys(d *schema.ResourceData) []*apigateway.StageKey {
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

	return stageKeys
}

func resourceAwsApiGatewayApiKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Creating API Gateway API Key")

	apiKey, err := conn.CreateApiKey(&apigateway.CreateApiKeyInput{
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Enabled:     aws.Bool(d.Get("enabled").(bool)),
		StageKeys:   resourceAwsApiGatewayStageKeys(d),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway: %s", err)
	}

	d.SetId(*(*apiKey).Id)

	return resourceAwsApiGatewayApiKeyRead(d, meta)
}

func resourceAwsApiGatewayApiKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Reading API Gateway API Key: %s", d.Id())

	apiKey, err := conn.GetApiKey(&apigateway.GetApiKeyInput{
		ApiKey: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	d.Set("name", *apiKey.Name)
	d.Set("description", *apiKey.Description)
	d.Set("enabled", *apiKey.Enabled)

	return nil
}

func resourceAwsApiGatewayApiKeyUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("enabled") {
		isEnabled := "false"
		if d.Get("enabled").(bool) {
			isEnabled = "true"
		}
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/enabled"),
			Value: aws.String(isEnabled),
		})
	}

	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	if d.HasChange("stage_key") {
		prev, curr := d.GetChange("stage_key")
		prevList := prev.(*schema.Set).List()
		currList := curr.(*schema.Set).List()

		for i := range prevList {
			p := prevList[i].(map[string]interface{})
			exists := false

			for j := range currList {
				c := currList[j].(map[string]interface{})
				if c["api_id"].(string) == p["api_id"].(string) && c["stage_name"].(string) == p["stage_name"].(string) {
					exists = true
				}
			}

			if !exists {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String("remove"),
					Path:  aws.String("/stages"),
					Value: aws.String(fmt.Sprintf("%s/%s", p["api_id"].(string), p["stage_name"].(string))),
				})
			}
		}

		for i := range currList {
			c := currList[i].(map[string]interface{})
			exists := false

			for j := range prevList {
				p := prevList[j].(map[string]interface{})
				if c["api_id"].(string) == p["api_id"].(string) && c["stage_name"].(string) == p["stage_name"].(string) {
					exists = true
				}
			}

			if !exists {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String("add"),
					Path:  aws.String("/stages"),
					Value: aws.String(fmt.Sprintf("%s/%s", c["api_id"].(string), c["stage_name"].(string))),
				})
			}
		}
	}
	return operations
}

func resourceAwsApiGatewayApiKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Updating API Gateway API Key: %s", d.Id())

	apiKey, err := conn.UpdateApiKey(&apigateway.UpdateApiKeyInput{
		ApiKey:          aws.String(d.Id()),
		PatchOperations: resourceAwsApiGatewayApiKeyUpdateOperations(d),
	})
	if err != nil {
		return err
	}

	d.Set("name", *apiKey.Name)
	d.Set("description", *apiKey.Description)
	d.Set("enabled", *apiKey.Enabled)

	// FIXME update stages from remote. api id is missing, tho, so it might be hard

	return nil
}

func resourceAwsApiGatewayApiKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway API Key: %s", d.Id())

	return resource.Retry(5*time.Minute, func() error {
		_, err := conn.DeleteApiKey(&apigateway.DeleteApiKeyInput{
			ApiKey: aws.String(d.Id()),
		})

		if err == nil {
			return nil
		}

		if apigatewayErr, ok := err.(awserr.Error); ok && apigatewayErr.Code() == "NotFoundException" {
			return nil
		}

		return resource.RetryError{Err: err}
	})
}
