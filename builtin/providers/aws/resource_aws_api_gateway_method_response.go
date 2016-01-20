package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayMethodResponse() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayMethodResponseCreate,
		Read:   resourceAwsApiGatewayMethodResponseRead,
		Update: resourceAwsApiGatewayMethodResponseUpdate,
		Delete: resourceAwsApiGatewayMethodResponseDelete,

		Schema: map[string]*schema.Schema{
			"api_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"http_method": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"status_code": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"response_models": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"response_headers": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeBool,
			},
		},
	}
}

func resourceAwsApiGatewayMethodResponseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	models := make(map[string]string)
	for k, v := range d.Get("response_models").(map[string]interface{}) {
		models[k] = v.(string)
	}

	headers := make(map[string]bool)
	if d.Get("response_headers") != nil {
		v := d.Get("response_headers").(map[string]interface{})
		for k, t := range v {
			t, _ = strconv.ParseBool((t).(string))
			headers["method.response.header."+k] = t.(bool)
		}
	}

	_, err := conn.PutMethodResponse(&apigateway.PutMethodResponseInput{
		HttpMethod:         aws.String(d.Get("http_method").(string)),
		ResourceId:         aws.String(d.Get("resource_id").(string)),
		RestApiId:          aws.String(d.Get("api_id").(string)),
		StatusCode:         aws.String(d.Get("status_code").(string)),
		ResponseModels:     aws.StringMap(models),
		ResponseParameters: aws.BoolMap(headers),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Method Response: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%s-%s", d.Get("api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string), d.Get("status_code").(string)))
	log.Printf("[DEBUG] API Gateway Method ID: %s", d.Id())

	return nil
}

func resourceAwsApiGatewayMethodResponseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway Method %s", d.Id())
	out, err := conn.GetMethodResponse(&apigateway.GetMethodResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("api_id").(string)),
		StatusCode: aws.String(d.Get("status_code").(string)),
	})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Method: %s", out)
	d.SetId(fmt.Sprintf("%s-%s-%s-%s", d.Get("api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string), d.Get("status_code").(string)))

	return nil
}

func resourceAwsApiGatewayMethodResponseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway Method %s", d.Id())
	operations := make([]*apigateway.PatchOperation, 0)
	if o, n := d.GetChange("resource_id"); o.(string) != n.(string) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/resourceId"),
			Value: aws.String(d.Get("resource_id").(string)),
		})
	}

	if d.HasChange("request_models") {
		o, n := d.GetChange("request_models")
		oP := o.(map[string]interface{})
		oN := n.(map[string]interface{})

		for k, _ := range oP {
			operation := apigateway.PatchOperation{
				Op:   aws.String("remove"),
				Path: aws.String(fmt.Sprintf("/requestModels/%s", strings.Replace(k, "/", "~1", -1))),
			}
			for nK, nV := range oN {
				if nK == k {
					operation.Op = aws.String("replace")
					operation.Value = aws.String(nV.(string))
				}
			}
			operations = append(operations, &operation)
		}

		for nK, nV := range oN {
			exists := false
			for k, _ := range oP {
				if k == nK {
					exists = true
				}
			}
			if !exists {
				operation := apigateway.PatchOperation{
					Op:    aws.String("add"),
					Path:  aws.String(fmt.Sprintf("/requestModels/%s", strings.Replace(nK, "/", "~1", -1))),
					Value: aws.String(nV.(string)),
				}
				operations = append(operations, &operation)
			}
		}
	}

	out, err := conn.UpdateMethod(&apigateway.UpdateMethodInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		ResourceId:      aws.String(d.Get("resource_id").(string)),
		RestApiId:       aws.String(d.Get("api_id").(string)),
		PatchOperations: operations,
	})

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received API Gateway Method: %s", out)

	return resourceAwsApiGatewayMethodRead(d, meta)
}

func resourceAwsApiGatewayMethodResponseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway Method: %s", d.Id())

	resourceId := d.Get("resource_id").(string)
	if o, n := d.GetChange("resource_id"); o.(string) != n.(string) {
		resourceId = o.(string)
	}
	httpMethod := d.Get("http_method").(string)
	if o, n := d.GetChange("http_method"); o.(string) != n.(string) {
		httpMethod = o.(string)
	}
	restApiID := d.Get("api_id").(string)
	if o, n := d.GetChange("api_id"); o.(string) != n.(string) {
		restApiID = o.(string)
	}
	statusCode := d.Get("status_code").(string)
	if o, n := d.GetChange("status_code"); o.(string) != n.(string) {
		statusCode = o.(string)
	}

	return resource.Retry(5*time.Minute, func() error {
		log.Printf("[DEBUG] schema is %#v", d)
		_, err := conn.DeleteMethodResponse(&apigateway.DeleteMethodResponseInput{
			HttpMethod: aws.String(httpMethod),
			ResourceId: aws.String(resourceId),
			RestApiId:  aws.String(restApiID),
			StatusCode: aws.String(statusCode),
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
