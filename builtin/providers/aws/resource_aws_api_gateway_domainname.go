package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

// AWS APIGateway domain name declaration
func resourceAwsApiGatewayDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayDomainCreate,
		Read:   resourceAwsApiGatewayDomainRead,
		Update: resourceAwsApiGatewayDomainUpdate,
		Delete: resourceAwsApiGatewayDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate_body": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate_private_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate_chain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsApiGatewayDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	r, err := conn.CreateDomainName(&apigateway.CreateDomainNameInput{
		DomainName:            aws.String(d.Get("domain_name").(string)),
		CertificateName:       aws.String(d.Get("certificate_name").(string)),
		CertificateBody:       aws.String(d.Get("certificate_body").(string)),
		CertificateChain:      aws.String(d.Get("certificate_chain").(string)),
		CertificatePrivateKey: aws.String(d.Get("certificate_private_key").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error creating API Gateway Domain: %s", err)
	}

	d.SetId(*r.DomainName)

	return resourceAwsApiGatewayDomainRead(d, meta)
}

func resourceAwsApiGatewayDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	r, err := conn.GetDomainName(&apigateway.GetDomainNameInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error creating API Gateway Domain: %s", err)
	}

	d.Set("certificate_name", *r.CertificateName)

	return nil
}

func resourceAwsApiGatewayDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	resourceAwsApiGatewayDomainDelete(d, meta)
	return resourceAwsApiGatewayDomainCreate(d, meta)
}

func resourceAwsApiGatewayDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	r, err := conn.DeleteDomainName(&apigateway.DeleteDomainNameInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err == nil {
		return nil
	}

	awsErr, ok := err.(awserr.Error)
	if awsErr.Code() == "NotFoundException" {
		return nil
	}

	return fmt.Errorf("Error Deleting Gateway Domain Name: %s", err)

}
