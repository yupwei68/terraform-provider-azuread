package aadgraph

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

// valid types are `application` and `service_principal`
func CertificateResourceSchema(idAttribute string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		idAttribute: {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validate.UUID,
		},

		"key_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ForceNew:         true,
			ValidateDiagFunc: validate.UUID,
		},

		"type": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			ValidateFunc: validation.StringInSlice([]string{
				"AsymmetricX509Cert",
				"Symmetric",
			}, false),
		},

		"value": {
			Type:      schema.TypeString,
			Required:  true,
			ForceNew:  true,
			Sensitive: true,
		},

		"start_date": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsRFC3339Time,
		},

		"end_date": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"end_date_relative"},
			ValidateFunc:  validation.IsRFC3339Time,
		},

		"end_date_relative": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			ConflictsWith:    []string{"end_date"},
			ValidateDiagFunc: validate.NoEmptyStrings,
		},
	}
}

// valid types are `application` and `service_principal`
func PasswordResourceSchema(idAttribute string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		idAttribute: {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validate.UUID,
		},

		"key_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ForceNew:         true,
			ValidateDiagFunc: validate.UUID,
		},

		"description": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},

		"value": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Sensitive:    true,
			ValidateFunc: validation.StringLenBetween(1, 863), // Encrypted secret cannot be empty and can be at most 1024 bytes.
		},

		"start_date": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsRFC3339Time,
		},

		"end_date": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ExactlyOneOf: []string{"end_date_relative"},
			ValidateFunc: validation.IsRFC3339Time,
		},

		"end_date_relative": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			ExactlyOneOf:     []string{"end_date"},
			ValidateDiagFunc: validate.NoEmptyStrings,
		},
	}
}

type CredentialId struct {
	ObjectId string
	KeyType  string
	KeyId    string
}

func (id CredentialId) String() string {
	return id.ObjectId + "/" + id.KeyType + "/" + id.KeyId
}

func ParseCertificateId(idString string) (*CredentialId, error) {
	id, err := ParseObjectSubResourceId(idString, "certificate")
	if err != nil {
		return nil, fmt.Errorf("unable to parse Certificate ID: %v", err)
	}

	return &CredentialId{
		ObjectId: id.objectId,
		KeyType:  id.Type,
		KeyId:    id.subId,
	}, nil
}

func ParsePasswordId(idString string) (*CredentialId, error) {
	id, err := ParseObjectSubResourceId(idString, "password")
	if err != nil {
		return nil, fmt.Errorf("unable to parse Password ID: %v", err)
	}

	return &CredentialId{
		ObjectId: id.objectId,
		KeyType:  id.Type,
		KeyId:    id.subId,
	}, nil
}

func ParseOldPasswordId(id string) (*CredentialId, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Password ID expected to be in the format {objectId}/{keyId} - but got %q", id)
	}

	newId := parts[0] + "/password/" + parts[1]
	return ParsePasswordId(newId)
}

func CredentialIdFrom(objectId, keyType, keyId string) CredentialId {
	return CredentialId{
		ObjectId: objectId,
		KeyType:  keyType,
		KeyId:    keyId,
	}
}

func PasswordCredentialForResource(d *schema.ResourceData) (*graphrbac.PasswordCredential, error) {
	value := d.Get("value").(string)

	// errors should be handled by the validation
	var keyId string
	if v, ok := d.GetOk("key_id"); ok {
		keyId = v.(string)
	} else {
		kid, err := uuid.GenerateUUID()
		if err != nil {
			return nil, err
		}

		keyId = kid
	}

	var endDate time.Time
	if v := d.Get("end_date").(string); v != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("unable to parse the provided end date %q: %+v", v, err), attr: "end_date"}
		}
	} else if v := d.Get("end_date_relative").(string); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("unable to parse `end_date_relative` (%q) as a duration", v), attr: "end_date_relative"}
		}
		endDate = time.Now().Add(d)
	} else {
		return nil, CredentialError{str: "one of `end_date` or `end_date_relative` must be specified", attr: "end_date"}
	}

	credential := graphrbac.PasswordCredential{
		KeyID:   utils.String(keyId),
		Value:   utils.String(value),
		EndDate: &date.Time{Time: endDate},
	}

	if v, ok := d.GetOk("description"); ok {
		customIdentifier := []byte(v.(string))
		credential.CustomKeyIdentifier = &customIdentifier
	}

	if v, ok := d.GetOk("start_date"); ok {
		startDate, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("unable to parse the provided start date %q: %+v", v, err), attr: "start_date"}
		}
		credential.StartDate = &date.Time{Time: startDate}
	}

	return &credential, nil
}

func PasswordCredentialResultFindByKeyId(creds graphrbac.PasswordCredentialListResult, keyId string) *graphrbac.PasswordCredential {
	var cred *graphrbac.PasswordCredential

	if creds.Value != nil {
		for _, c := range *creds.Value {
			if c.KeyID == nil {
				continue
			}

			if *c.KeyID == keyId {
				cred = &c
				break
			}
		}
	}

	return cred
}

func PasswordCredentialResultAdd(existing graphrbac.PasswordCredentialListResult, cred *graphrbac.PasswordCredential) (*[]graphrbac.PasswordCredential, error) {
	if cred == nil {
		return nil, errors.New("credential to be added is null")
	}

	newCreds := make([]graphrbac.PasswordCredential, 0)

	if existing.Value != nil {
		for _, v := range *existing.Value {
			if v.KeyID == nil {
				continue
			}
			if *v.KeyID == *cred.KeyID {
				return nil, &AlreadyExistsError{"Password Credential", *cred.KeyID}
			}
		}

		newCreds = *existing.Value
	}
	newCreds = append(newCreds, *cred)

	return &newCreds, nil
}

func PasswordCredentialResultRemoveByKeyId(existing graphrbac.PasswordCredentialListResult, keyId string) (*[]graphrbac.PasswordCredential, error) {
	if keyId == "" {
		return nil, errors.New("ID of key to be removed is blank")
	}

	newCreds := make([]graphrbac.PasswordCredential, 0)

	if existing.Value != nil {
		for _, v := range *existing.Value {
			if v.KeyID == nil {
				continue
			}

			if *v.KeyID == keyId {
				continue
			}

			newCreds = append(newCreds, v)
		}
	}

	return &newCreds, nil
}

func WaitForPasswordCredentialReplication(ctx context.Context, keyId string, timeout time.Duration, f func() (graphrbac.PasswordCredentialListResult, error)) (interface{}, error) {
	return (&resource.StateChangeConf{
		Pending:                   []string{"NotFound"},
		Target:                    []string{"Found"},
		Timeout:                   timeout,
		MinTimeout:                1 * time.Second,
		ContinuousTargetOccurence: 10,
		Refresh: func() (interface{}, string, error) {
			creds, err := f()
			if err != nil {
				if utils.ResponseWasNotFound(creds.Response) {
					return creds, "NotFound", nil
				}
				return creds, "Error", fmt.Errorf("unable to retrieve object, received response with status %d: %v", creds.Response.StatusCode, err)
			}

			credential := PasswordCredentialResultFindByKeyId(creds, keyId)
			if credential == nil {
				return creds, "NotFound", nil
			}

			return creds, "Found", nil
		},
	}).WaitForStateContext(ctx)
}

func KeyCredentialForResource(d *schema.ResourceData) (*graphrbac.KeyCredential, error) {
	keyType := d.Get("type").(string)
	value := d.Get("value").(string)
	encodedValue := base64.StdEncoding.EncodeToString([]byte(value))

	// errors should be handled by the validation
	var keyId string
	if v, ok := d.GetOk("key_id"); ok {
		keyId = v.(string)
	} else {
		kid, err := uuid.GenerateUUID()
		if err != nil {
			return nil, err
		}

		keyId = kid
	}

	var endDate time.Time
	if v := d.Get("end_date").(string); v != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("unable to parse the provided end date %q: %+v", v, err), attr: "end_date"}
		}
	} else if v := d.Get("end_date_relative").(string); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("unable to parse `end_date_relative` (%q) as a duration", v), attr: "end_date_relative"}
		}
		endDate = time.Now().Add(d)
	} else {
		return nil, CredentialError{str: "one of `end_date` or `end_date_relative` must be specified", attr: "end_date"}
	}

	credential := graphrbac.KeyCredential{
		KeyID:   utils.String(keyId),
		Type:    utils.String(keyType),
		Usage:   utils.String("verify"),
		Value:   utils.String(encodedValue),
		EndDate: &date.Time{Time: endDate},
	}

	if v, ok := d.GetOk("start_date"); ok {
		startDate, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("unable to parse the provided start date %q: %+v", v, err), attr: "start_date"}
		}
		credential.StartDate = &date.Time{Time: startDate}
	}

	return &credential, nil
}

func KeyCredentialResultFindByKeyId(creds graphrbac.KeyCredentialListResult, keyId string) *graphrbac.KeyCredential {
	if creds.Value != nil {
		for _, c := range *creds.Value {
			if c.KeyID == nil {
				continue
			}
			if *c.KeyID == keyId {
				return &c
			}
		}
	}

	return nil
}

func KeyCredentialResultAdd(existing graphrbac.KeyCredentialListResult, cred *graphrbac.KeyCredential) (*[]graphrbac.KeyCredential, error) {
	newCreds := make([]graphrbac.KeyCredential, 0)

	if existing.Value != nil {
		for _, v := range *existing.Value {
			if v.KeyID == nil {
				continue
			}

			if *v.KeyID == *cred.KeyID {
				return nil, &AlreadyExistsError{"Key Credential", *cred.KeyID}
			}
		}

		newCreds = *existing.Value
	}
	newCreds = append(newCreds, *cred)

	return &newCreds, nil
}

func KeyCredentialResultRemoveByKeyId(existing graphrbac.KeyCredentialListResult, keyId string) (*[]graphrbac.KeyCredential, error) {
	if keyId == "" {
		return nil, errors.New("ID of key to be removed is blank")
	}

	newCreds := make([]graphrbac.KeyCredential, 0)

	if existing.Value != nil {
		for _, v := range *existing.Value {
			if v.KeyID == nil {
				continue
			}

			if *v.KeyID == keyId {
				continue
			}

			newCreds = append(newCreds, v)
		}
	}

	return &newCreds, nil
}

func WaitForKeyCredentialReplication(ctx context.Context, keyId string, timeout time.Duration, f func() (graphrbac.KeyCredentialListResult, error)) (interface{}, error) {
	return (&resource.StateChangeConf{
		Pending:                   []string{"NotFound"},
		Target:                    []string{"Found"},
		Timeout:                   timeout,
		MinTimeout:                1 * time.Second,
		ContinuousTargetOccurence: 10,
		Refresh: func() (interface{}, string, error) {
			creds, err := f()
			if err != nil {
				if utils.ResponseWasNotFound(creds.Response) {
					return creds, "NotFound", nil
				}
				return creds, "Error", fmt.Errorf("unable to retrieve object, received response with status %d: %v", creds.Response.StatusCode, err)
			}

			credential := KeyCredentialResultFindByKeyId(creds, keyId)
			if credential == nil {
				return creds, "NotFound", nil
			}

			return creds, "Found", nil
		},
	}).WaitForStateContext(ctx)
}
