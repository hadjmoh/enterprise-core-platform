package cloud

import (
	"fmt"
	"strings"
	"time"
)

// Normalizer handles conversion of provider-specific data to unified format
type Normalizer struct{}

// NewNormalizer creates a new normalizer instance
func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// NormalizeEvent converts provider-specific events to CloudEvent
func (n *Normalizer) NormalizeEvent(provider Provider, raw map[string]interface{}) (*CloudEvent, error) {
	switch provider {
	case ProviderAWS:
		return n.normalizeAWSEvent(raw)
	case ProviderAzure:
		return n.normalizeAzureEvent(raw)
	case ProviderGCP:
		return n.normalizeGCPEvent(raw)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// normalizeAWSEvent converts AWS CloudTrail event to CloudEvent
func (n *Normalizer) normalizeAWSEvent(raw map[string]interface{}) (*CloudEvent, error) {
	event := &CloudEvent{
		Provider: ProviderAWS,
		RawEvent: raw,
	}
	
	// Extract common fields
	if eventID, ok := raw["eventID"].(string); ok {
		event.ID = eventID
	}
	
	if eventName, ok := raw["eventName"].(string); ok {
		event.EventName = eventName
		event.EventType = n.normalizeAWSEventType(eventName)
	}
	
	if eventTime, ok := raw["eventTime"].(string); ok {
		if t, err := time.Parse(time.RFC3339, eventTime); err == nil {
			event.Timestamp = t
		}
	}
	
	// Extract actor information
	if userIdentity, ok := raw["userIdentity"].(map[string]interface{}); ok {
		if principalID, ok := userIdentity["principalId"].(string); ok {
			event.ActorID = principalID
		}
		if userType, ok := userIdentity["type"].(string); ok {
			event.ActorType = userType
		}
		if arn, ok := userIdentity["arn"].(string); ok {
			event.ActorName = arn
		}
	}
	
	// Extract resource information
	if resources, ok := raw["resources"].([]interface{}); ok && len(resources) > 0 {
		if resource, ok := resources[0].(map[string]interface{}); ok {
			if arn, ok := resource["ARN"].(string); ok {
				event.ResourceID = arn
			}
			if resourceType, ok := resource["type"].(string); ok {
				event.ResourceType = resourceType
			}
		}
	}
	
	// Extract context
	if sourceIP, ok := raw["sourceIPAddress"].(string); ok {
		event.SourceIP = sourceIP
	}
	
	if userAgent, ok := raw["userAgent"].(string); ok {
		event.UserAgent = userAgent
	}
	
	if region, ok := raw["awsRegion"].(string); ok {
		event.Region = region
	}
	
	if accountID, ok := raw["recipientAccountId"].(string); ok {
		event.Account = accountID
	}
	
	// Determine success
	if errorCode, ok := raw["errorCode"].(string); ok {
		event.Success = false
		event.ErrorCode = errorCode
		if errorMessage, ok := raw["errorMessage"].(string); ok {
			event.ErrorMessage = errorMessage
		}
	} else {
		event.Success = true
	}
	
	return event, nil
}

// normalizeAWSEventType converts AWS event name to generic event type
func (n *Normalizer) normalizeAWSEventType(eventName string) string {
	// Map AWS event names to generic types
	// Format: <service>:<resource>:<action>
	
	lower := strings.ToLower(eventName)
	
	// IAM events
	if strings.Contains(lower, "createuser") {
		return "iam:user:create"
	}
	if strings.Contains(lower, "deleteuser") {
		return "iam:user:delete"
	}
	if strings.Contains(lower, "attachuserpolicy") {
		return "iam:policy:attach"
	}
	
	// EC2 events
	if strings.Contains(lower, "runinstances") {
		return "compute:instance:create"
	}
	if strings.Contains(lower, "terminateinstances") {
		return "compute:instance:delete"
	}
	if strings.Contains(lower, "stopinstances") {
		return "compute:instance:stop"
	}
	
	// S3 events
	if strings.Contains(lower, "putbucket") {
		return "storage:bucket:create"
	}
	if strings.Contains(lower, "deletebucket") {
		return "storage:bucket:delete"
	}
	
	// Default: use event name as-is
	return eventName
}

// normalizeAzureEvent converts Azure Activity Log to CloudEvent
func (n *Normalizer) normalizeAzureEvent(raw map[string]interface{}) (*CloudEvent, error) {
	event := &CloudEvent{
		Provider: ProviderAzure,
		RawEvent: raw,
	}
	
	// Extract common fields
	if eventID, ok := raw["eventDataId"].(string); ok {
		event.ID = eventID
	}
	
	if operationName, ok := raw["operationName"].(string); ok {
		event.EventName = operationName
		event.EventType = n.normalizeAzureEventType(operationName)
	}
	
	if eventTime, ok := raw["eventTimestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, eventTime); err == nil {
			event.Timestamp = t
		}
	}
	
	// Extract caller
	if caller, ok := raw["caller"].(string); ok {
		event.ActorName = caller
		event.ActorID = caller
		event.ActorType = "user"
	}
	
	// Extract resource
	if resourceID, ok := raw["resourceId"].(string); ok {
		event.ResourceID = resourceID
	}
	
	if resourceType, ok := raw["resourceType"].(string); ok {
		event.ResourceType = resourceType
	}
	
	// Extract status
	if status, ok := raw["status"].(map[string]interface{}); ok {
		if value, ok := status["value"].(string); ok {
			event.Success = strings.ToLower(value) == "succeeded"
		}
	}
	
	return event, nil
}

// normalizeAzureEventType converts Azure operation name to generic event type
func (n *Normalizer) normalizeAzureEventType(operationName string) string {
	lower := strings.ToLower(operationName)
	
	if strings.Contains(lower, "microsoft.compute/virtualmachines/write") {
		return "compute:instance:create"
	}
	if strings.Contains(lower, "microsoft.compute/virtualmachines/delete") {
		return "compute:instance:delete"
	}
	if strings.Contains(lower, "microsoft.storage/storageaccounts/write") {
		return "storage:account:create"
	}
	
	return operationName
}

// normalizeGCPEvent converts GCP Audit Log to CloudEvent
func (n *Normalizer) normalizeGCPEvent(raw map[string]interface{}) (*CloudEvent, error) {
	event := &CloudEvent{
		Provider: ProviderGCP,
		RawEvent: raw,
	}
	
	// Extract common fields
	if insertID, ok := raw["insertId"].(string); ok {
		event.ID = insertID
	}
	
	if protoPayload, ok := raw["protoPayload"].(map[string]interface{}); ok {
		if methodName, ok := protoPayload["methodName"].(string); ok {
			event.EventName = methodName
			event.EventType = n.normalizeGCPEventType(methodName)
		}
		
		// Extract authentication info
		if authInfo, ok := protoPayload["authenticationInfo"].(map[string]interface{}); ok {
			if principalEmail, ok := authInfo["principalEmail"].(string); ok {
				event.ActorName = principalEmail
				event.ActorID = principalEmail
				event.ActorType = "user"
			}
		}
		
		// Extract resource
		if resourceName, ok := protoPayload["resourceName"].(string); ok {
			event.ResourceID = resourceName
		}
		
		// Extract status
		if status, ok := protoPayload["status"].(map[string]interface{}); ok {
			if code, ok := status["code"].(float64); ok {
				event.Success = code == 0
				if code != 0 {
					event.ErrorCode = fmt.Sprintf("%d", int(code))
				}
			}
		}
	}
	
	if timestamp, ok := raw["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			event.Timestamp = t
		}
	}
	
	return event, nil
}

// normalizeGCPEventType converts GCP method name to generic event type
func (n *Normalizer) normalizeGCPEventType(methodName string) string {
	lower := strings.ToLower(methodName)
	
	if strings.Contains(lower, "compute.instances.insert") {
		return "compute:instance:create"
	}
	if strings.Contains(lower, "compute.instances.delete") {
		return "compute:instance:delete"
	}
	if strings.Contains(lower, "storage.buckets.create") {
		return "storage:bucket:create"
	}
	
	return methodName
}

// NormalizeAsset converts provider-specific asset to CloudAsset
func (n *Normalizer) NormalizeAsset(provider Provider, raw map[string]interface{}) (*CloudAsset, error) {
	switch provider {
	case ProviderAWS:
		return n.normalizeAWSAsset(raw)
	case ProviderAzure:
		return n.normalizeAzureAsset(raw)
	case ProviderGCP:
		return n.normalizeGCPAsset(raw)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// normalizeAWSAsset converts AWS EC2 instance to CloudAsset
func (n *Normalizer) normalizeAWSAsset(raw map[string]interface{}) (*CloudAsset, error) {
	asset := &CloudAsset{
		Provider:    ProviderAWS,
		Type:        "ec2",
		Tags:        make(map[string]string),
		RawMetadata: raw,
	}
	
	// Extract instance ID
	if instanceID, ok := raw["InstanceId"].(string); ok {
		asset.ID = fmt.Sprintf("aws:ec2:%s", instanceID)
		asset.Name = instanceID
	}
	
	// Extract state
	if state, ok := raw["State"].(map[string]interface{}); ok {
		if name, ok := state["Name"].(string); ok {
			asset.State = name
		}
	}
	
	// Extract network info
	if privateIP, ok := raw["PrivateIpAddress"].(string); ok {
		asset.PrivateIP = privateIP
	}
	
	if publicIP, ok := raw["PublicIpAddress"].(string); ok {
		asset.PublicIP = publicIP
	}
	
	if vpcID, ok := raw["VpcId"].(string); ok {
		asset.VPC = vpcID
	}
	
	// Extract security groups
	if sgs, ok := raw["SecurityGroups"].([]interface{}); ok {
		asset.SecurityGroups = make([]string, 0, len(sgs))
		for _, sg := range sgs {
			if sgMap, ok := sg.(map[string]interface{}); ok {
				if sgID, ok := sgMap["GroupId"].(string); ok {
					asset.SecurityGroups = append(asset.SecurityGroups, sgID)
				}
			}
		}
	}
	
	// Extract tags
	if tags, ok := raw["Tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagMap, ok := tag.(map[string]interface{}); ok {
				if key, ok := tagMap["Key"].(string); ok {
					if value, ok := tagMap["Value"].(string); ok {
						asset.Tags[key] = value
						if key == "Name" {
							asset.Name = value
						}
					}
				}
			}
		}
	}
	
	// Extract IAM role
	if iamProfile, ok := raw["IamInstanceProfile"].(map[string]interface{}); ok {
		if arn, ok := iamProfile["Arn"].(string); ok {
			asset.IAMRole = arn
		}
	}
	
	asset.LastSeen = time.Now()
	
	return asset, nil
}

// normalizeAzureAsset converts Azure VM to CloudAsset
func (n *Normalizer) normalizeAzureAsset(raw map[string]interface{}) (*CloudAsset, error) {
	asset := &CloudAsset{
		Provider:    ProviderAzure,
		Type:        "vm",
		Tags:        make(map[string]string),
		RawMetadata: raw,
	}
	
	// Extract ID and name
	if id, ok := raw["id"].(string); ok {
		asset.ID = fmt.Sprintf("azure:vm:%s", id)
	}
	
	if name, ok := raw["name"].(string); ok {
		asset.Name = name
	}
	
	// Extract location (region)
	if location, ok := raw["location"].(string); ok {
		asset.Region = location
	}
	
	// Extract tags
	if tags, ok := raw["tags"].(map[string]interface{}); ok {
		for key, value := range tags {
			if strValue, ok := value.(string); ok {
				asset.Tags[key] = strValue
			}
		}
	}
	
	asset.LastSeen = time.Now()
	
	return asset, nil
}

// normalizeGCPAsset converts GCP Compute Instance to CloudAsset
func (n *Normalizer) normalizeGCPAsset(raw map[string]interface{}) (*CloudAsset, error) {
	asset := &CloudAsset{
		Provider:    ProviderGCP,
		Type:        "compute-instance",
		Tags:        make(map[string]string),
		RawMetadata: raw,
	}
	
	// Extract ID and name
	if id, ok := raw["id"].(string); ok {
		asset.ID = fmt.Sprintf("gcp:compute:%s", id)
	}
	
	if name, ok := raw["name"].(string); ok {
		asset.Name = name
	}
	
	// Extract zone (region)
	if zone, ok := raw["zone"].(string); ok {
		asset.Region = zone
	}
	
	// Extract status
	if status, ok := raw["status"].(string); ok {
		asset.State = strings.ToLower(status)
	}
	
	// Extract labels (GCP's version of tags)
	if labels, ok := raw["labels"].(map[string]interface{}); ok {
		for key, value := range labels {
			if strValue, ok := value.(string); ok {
				asset.Tags[key] = strValue
			}
		}
	}
	
	asset.LastSeen = time.Now()
	
	return asset, nil
}
