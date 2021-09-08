//  ISV Serverless Nodepool Tagger, Ed Shnekendorf, September 2021
//  Copyright (c) 2021 Oracle and/or its affiliates.
//  Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
package main

import "time"

// OCI Events Spec:  https://docs.oracle.com/en-us/iaas/Content/Events/Reference/eventenvelopereference.htm
type OCIEvent struct {
	CloudEventsVersion string      `json:"cloudEventsVersion"`
	EventID            string      `json:"eventID"`
	EventType          string      `json:"eventType"`
	Source             string      `json:"source"`
	EventTypeVersion   string      `json:"eventTypeVersion"`
	EventTime          time.Time   `json:"eventTime"`
	SchemaURL          interface{} `json:"schemaURL"`
	ContentType        string      `json:"contentType"`
	Extensions         Extensions  `json:"extensions"`
	Data               Data        `json:"data"`
}

type AdditionalDetails struct {
	ETag          string      `json:"eTag"`
	Namespace     string      `json:"namespace"`
	ArchieveState interface{} `json:"archieveState"`
	BucketName    string      `json:"bucketName"`
	BucketID      string      `json:"bucketId"`
}

type Data struct {
	CompartmentID      string            `json:"compartmentId"`
	CompartmentName    string            `json:"compartmentName"`
	ResourceName       string            `json:"resourceName"`
	ResourceID         string            `json:"resourceId"`
	AvailabilityDomain string            `json:"availabilityDomain"`
	FreeFormTags       FreeFormTags      `json:"freeFormTags"`
	DefinedTags        DefinedTags       `json:"definedTags"`
	AdditionalDetails  AdditionalDetails `json:"additionalDetails"`
}

type Extensions struct {
	CompartmentID string `json:"compartmentId"`
}

type FreeFormTags struct {
}

type DefinedTags struct {
}
